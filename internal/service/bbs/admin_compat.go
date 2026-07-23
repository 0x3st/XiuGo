package bbs

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/0x3st/XiuGo/internal/dao"
	"github.com/0x3st/XiuGo/internal/model/do"
	"github.com/0x3st/XiuGo/internal/model/entity"
	"github.com/0x3st/XiuGo/internal/model/view"
)

func (s *Service) VerifyAdminPassword(ctx context.Context, uid uint, browserPassword string) error {
	var user entity.BbsUser
	if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取管理员失败")
	}
	if user.Uid == 0 || user.Gid != 1 {
		return gerror.New("只允许管理员登录后台")
	}
	serverHash, err := xiunoPassword(browserPassword, user.Salt)
	if err != nil {
		return err
	}
	if serverHash != user.Password {
		return gerror.New("密码不正确")
	}
	return nil
}

func ApplyOriginalTimezone(ctx context.Context) error {
	service := New()
	content, err := os.ReadFile(filepath.Join(service.phpRoot(ctx), "conf", "conf.php"))
	if err != nil {
		return gerror.Wrap(err, "读取 Xiuno 时区配置失败")
	}
	name := phpConfigString(content, "timezone")
	if name == "" {
		return nil
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return gerror.Wrapf(err, "加载 Xiuno 时区 %s 失败", name)
	}
	time.Local = location
	return nil
}

func (s *Service) OriginalAdminDashboard(ctx context.Context, clientIP string) (result view.AdminDashboard, err error) {
	if result.Threads, err = dao.BbsThread.Ctx(ctx).Count(); err != nil {
		return result, gerror.Wrap(err, "统计主题失败")
	}
	if result.Posts, err = dao.BbsPost.Ctx(ctx).Count(); err != nil {
		return result, gerror.Wrap(err, "统计帖子失败")
	}
	if result.Users, err = dao.BbsUser.Ctx(ctx).Count(); err != nil {
		return result, gerror.Wrap(err, "统计用户失败")
	}
	if result.Attachs, err = dao.BbsAttach.Ctx(ctx).Count(); err != nil {
		return result, gerror.Wrap(err, "统计附件失败")
	}
	if result.Onlines, err = dao.BbsSession.Ctx(ctx).
		WhereGT(
			dao.BbsSession.Columns().LastDate,
			uint(time.Now().Unix()-int64(s.originalOnlineHoldSeconds(ctx))),
		).Count(); err != nil {
		return result, gerror.Wrap(err, "统计在线人数失败")
	}
	if result.Onlines < 1 {
		result.Onlines = 1
	}
	result.DiskFreeSpace = diskFreeSpace(s.phpRoot(ctx))
	result.OS = runtime.GOOS
	result.WebServer = "GoFrame HTTP Server"
	result.GoVersion = runtime.Version()
	result.Database = "mysql"
	result.PostMaxSize = "8M"
	result.UploadMaxSize = "2M"
	result.ClientIP = clientIP
	result.ServerIP = "127.0.0.1"
	result.AllowURLFopen = "支持"
	result.SafeMode = "关闭"
	result.MaxExecution = "0"
	result.MemoryLimit = "由 Go 运行时动态管理"
	return result, nil
}

func (s *Service) SiteSettings(ctx context.Context) (settings view.SiteSettings, err error) {
	content, err := os.ReadFile(filepath.Join(s.phpRoot(ctx), "conf", "conf.php"))
	if err != nil {
		return settings, gerror.Wrap(err, "读取 Xiuno 配置失败")
	}
	settings.Sitename = phpConfigString(content, "sitename")
	settings.Sitebrief = phpConfigString(content, "sitebrief")
	settings.Runlevel = phpConfigInt(content, "runlevel")
	settings.RunlevelReason = phpConfigString(content, "runlevel_reason")
	settings.UserCreateOn = phpConfigInt(content, "user_create_on")
	settings.UserCreateEmailOn = phpConfigInt(content, "user_create_email_on")
	settings.UserResetpwOn = phpConfigInt(content, "user_resetpw_on")
	settings.Lang = phpConfigString(content, "lang")
	return settings, nil
}

func (s *Service) UpdateSiteSettings(ctx context.Context, settings view.SiteSettings) error {
	path := filepath.Join(s.phpRoot(ctx), "conf", "conf.php")
	content, err := os.ReadFile(path)
	if err != nil {
		return gerror.Wrap(err, "读取 Xiuno 配置失败")
	}
	values := []struct {
		key     string
		literal string
	}{
		{"sitename", phpQuote(settings.Sitename)},
		{"sitebrief", phpQuote(settings.Sitebrief)},
		{"runlevel", strconv.Itoa(settings.Runlevel)},
		{"user_create_on", strconv.Itoa(settings.UserCreateOn)},
		{"user_create_email_on", strconv.Itoa(settings.UserCreateEmailOn)},
		{"user_resetpw_on", strconv.Itoa(settings.UserResetpwOn)},
		{"lang", phpQuote(settings.Lang)},
	}
	for _, value := range values {
		var replaced bool
		content, replaced = replacePHPConfigLiteral(content, value.key, value.literal)
		if !replaced {
			return gerror.Newf("Xiuno 配置缺少字段：%s", value.key)
		}
	}
	if err = writeFileAtomic(path, content); err != nil {
		return gerror.Wrap(err, "保存 Xiuno 配置失败")
	}
	return nil
}

func (s *Service) SMTPAccounts(ctx context.Context) ([]view.SMTPAccount, error) {
	content, err := os.ReadFile(filepath.Join(s.phpRoot(ctx), "conf", "smtp.conf.php"))
	if err != nil {
		if os.IsNotExist(err) {
			return []view.SMTPAccount{}, nil
		}
		return nil, gerror.Wrap(err, "读取 SMTP 配置失败")
	}
	pattern := regexp.MustCompile(`(?s)array\s*\(\s*'email'\s*=>\s*'((?:\\.|[^'])*)',\s*'host'\s*=>\s*'((?:\\.|[^'])*)',\s*'port'\s*=>\s*(?:'((?:\\.|[^'])*)'|(-?\d+)),\s*'user'\s*=>\s*'((?:\\.|[^'])*)',\s*'pass'\s*=>\s*'((?:\\.|[^'])*)',\s*\)`)
	matches := pattern.FindAllSubmatch(content, -1)
	accounts := make([]view.SMTPAccount, 0, len(matches))
	for _, match := range matches {
		portText := phpUnescape(string(match[3]))
		if portText == "" {
			portText = string(match[4])
		}
		port, _ := strconv.Atoi(portText)
		accounts = append(accounts, view.SMTPAccount{
			Email: phpUnescape(string(match[1])), Host: phpUnescape(string(match[2])), Port: port,
			User: phpUnescape(string(match[5])), Pass: phpUnescape(string(match[6])),
		})
	}
	return accounts, nil
}

func (s *Service) UpdateSMTPAccounts(ctx context.Context, accounts []view.SMTPAccount) error {
	var builder strings.Builder
	builder.WriteString("<?php\r\nreturn array (\r\n")
	for index, account := range accounts {
		builder.WriteString(fmt.Sprintf("  %d => \r\n  array (\r\n", index))
		builder.WriteString("    'email' => " + phpQuote(account.Email) + ",\r\n")
		builder.WriteString("    'host' => " + phpQuote(account.Host) + ",\r\n")
		builder.WriteString("    'port' => " + strconv.Itoa(account.Port) + ",\r\n")
		builder.WriteString("    'user' => " + phpQuote(account.User) + ",\r\n")
		builder.WriteString("    'pass' => " + phpQuote(account.Pass) + ",\r\n")
		builder.WriteString("  ),\r\n")
	}
	builder.WriteString(");\r\n?>")
	path := filepath.Join(s.phpRoot(ctx), "conf", "smtp.conf.php")
	if err := writeFileAtomic(path, []byte(builder.String())); err != nil {
		return gerror.Wrap(err, "保存 SMTP 配置失败")
	}
	return nil
}

func (s *Service) phpRoot(ctx context.Context) string {
	root := g.Cfg().MustGet(ctx, "xiuno.phpRoot", "../xiuno-bbs").String()
	if absolute, err := filepath.Abs(root); err == nil {
		return absolute
	}
	return root
}

func phpConfigString(content []byte, key string) string {
	pattern := regexp.MustCompile(`(?m)^\s*'` + regexp.QuoteMeta(key) + `'\s*=>\s*'((?:\\.|[^'])*)'\s*,`)
	match := pattern.FindSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return phpUnescape(string(match[1]))
}

func phpConfigInt(content []byte, key string) int {
	pattern := regexp.MustCompile(`(?m)^\s*'` + regexp.QuoteMeta(key) + `'\s*=>\s*(-?\d+)\s*,`)
	match := pattern.FindSubmatch(content)
	if len(match) < 2 {
		return 0
	}
	value, _ := strconv.Atoi(string(match[1]))
	return value
}

func (s *Service) phpCacheKeys(ctx context.Context, key string) []string {
	content, err := os.ReadFile(filepath.Join(s.phpRoot(ctx), "conf", "conf.php"))
	prefixes := []string{"", "bbs_"}
	if err == nil {
		pattern := regexp.MustCompile(`'cachepre'\s*=>\s*'((?:\\.|[^'])*)'`)
		for _, match := range pattern.FindAllSubmatch(content, -1) {
			if len(match) > 1 {
				prefixes = append(prefixes, phpUnescape(string(match[1])))
			}
		}
	}
	keys := make([]string, 0, len(prefixes))
	seen := make(map[string]bool)
	for _, prefix := range prefixes {
		candidate := prefix + key
		if len(candidate) > 32 {
			candidate = fmt.Sprintf("%x", md5.Sum([]byte(candidate)))
		}
		if !seen[candidate] {
			seen[candidate] = true
			keys = append(keys, candidate)
		}
	}
	return keys
}

func replacePHPConfigLiteral(content []byte, key, literal string) ([]byte, bool) {
	pattern := regexp.MustCompile(`(?m)(^\s*'` + regexp.QuoteMeta(key) + `'\s*=>\s*)(?:'(?:\\.|[^'])*'|-?\d+|true|false)(\s*,)`)
	indexes := pattern.FindSubmatchIndex(content)
	if len(indexes) < 6 {
		return content, false
	}
	result := make([]byte, 0, len(content)+len(literal))
	result = append(result, content[:indexes[2]]...)
	result = append(result, content[indexes[2]:indexes[3]]...)
	result = append(result, literal...)
	result = append(result, content[indexes[4]:indexes[5]]...)
	result = append(result, content[indexes[1]:]...)
	return result, true
}

func phpQuote(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return `'` + value + `'`
}

func phpUnescape(value string) string {
	value = strings.ReplaceAll(value, `\'`, `'`)
	return strings.ReplaceAll(value, `\\`, `\`)
}

func writeFileAtomic(path string, content []byte) error {
	info, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	mode := os.FileMode(0o644)
	if info != nil {
		mode = info.Mode()
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".xiuno-go-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err = temporary.Write(content); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func diskFreeSpace(path string) string {
	var stats syscall.Statfs_t
	if err := syscall.Statfs(path, &stats); err != nil {
		return "未知"
	}
	bytes := uint64(stats.Bavail) * uint64(stats.Bsize)
	const (
		kib = 1024
		mib = 1024 * kib
		gib = 1024 * mib
	)
	switch {
	case bytes > gib:
		return fmt.Sprintf("%.2fG", float64(bytes)/gib)
	case bytes > mib:
		return fmt.Sprintf("%.2fM", float64(bytes)/mib)
	case bytes > kib:
		return fmt.Sprintf("%.2fK", float64(bytes)/kib)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
