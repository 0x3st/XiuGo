package bbs

import (
	"errors"
	"database/sql"
	"encoding/json"
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
	name := g.Cfg().MustGet(ctx, "xiuno.timezone", "").String()
	if name == "" {
		content, err := os.ReadFile(filepath.Join(service.phpRoot(ctx), "conf", "conf.php"))
		if err != nil {
			// No legacy conf and no yaml timezone — keep process default.
			return nil
		}
		name = phpConfigString(content, "timezone")
	}
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
	result.DiskFreeSpace = diskFreeSpace(s.uploadRoot(ctx))
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

const (
	kvSiteSettings = "xiugo_site"
	kvSMTPAccounts = "xiugo_smtp"
)

func (s *Service) SiteSettings(ctx context.Context) (settings view.SiteSettings, err error) {
	if raw, ok, loadErr := s.loadKV(ctx, kvSiteSettings); loadErr != nil {
		return settings, loadErr
	} else if ok && raw != "" {
		if err = json.Unmarshal([]byte(raw), &settings); err != nil {
			return settings, gerror.Wrap(err, "解析站点配置失败")
		}
		return settings, nil
	}
	// First run: import from legacy conf.php if present, else yaml defaults.
	settings = s.bootstrapSiteSettingsFromLegacy(ctx)
	if err = s.saveKV(ctx, kvSiteSettings, settings); err != nil {
		return settings, err
	}
	return settings, nil
}

func (s *Service) bootstrapSiteSettingsFromLegacy(ctx context.Context) view.SiteSettings {
	var settings view.SiteSettings
	if root := s.phpRoot(ctx); root != "" {
		path := filepath.Join(root, "conf", "conf.php")
		if content, err := os.ReadFile(path); err == nil {
			settings.Sitename = phpConfigString(content, "sitename")
			settings.Sitebrief = phpConfigString(content, "sitebrief")
			settings.Runlevel = phpConfigInt(content, "runlevel")
			settings.RunlevelReason = phpConfigString(content, "runlevel_reason")
			settings.UserCreateOn = phpConfigInt(content, "user_create_on")
			settings.UserCreateEmailOn = phpConfigInt(content, "user_create_email_on")
			settings.UserResetpwOn = phpConfigInt(content, "user_resetpw_on")
			settings.Lang = phpConfigString(content, "lang")
		}
	}
	if settings.Sitename == "" {
		settings.Sitename = g.Cfg().MustGet(ctx, "xiuno.sitename", "XiuGo").String()
	}
	if settings.Sitebrief == "" {
		settings.Sitebrief = g.Cfg().MustGet(ctx, "xiuno.sitebrief", "").String()
	}
	if settings.Lang == "" {
		settings.Lang = g.Cfg().MustGet(ctx, "xiuno.lang", "zh-cn").String()
	}
	if s.phpRoot(ctx) == "" {
		if settings.Runlevel == 0 {
			settings.Runlevel = g.Cfg().MustGet(ctx, "xiuno.runlevel", 5).Int()
		}
		if settings.UserCreateOn == 0 {
			settings.UserCreateOn = g.Cfg().MustGet(ctx, "xiuno.userCreateOn", 1).Int()
		}
	}
	return settings
}

func (s *Service) UpdateSiteSettings(ctx context.Context, settings view.SiteSettings) error {
	if strings.TrimSpace(settings.Sitename) == "" {
		return gerror.New("站点名称不能为空")
	}
	if settings.Runlevel < 0 || settings.Runlevel > 5 {
		return gerror.New("运行级别不正确")
	}
	if settings.Lang == "" {
		settings.Lang = "zh-cn"
	}
	return s.saveKV(ctx, kvSiteSettings, settings)
}

func (s *Service) SMTPAccounts(ctx context.Context) ([]view.SMTPAccount, error) {
	if raw, ok, err := s.loadKV(ctx, kvSMTPAccounts); err != nil {
		return nil, err
	} else if ok && raw != "" {
		var accounts []view.SMTPAccount
		if err = json.Unmarshal([]byte(raw), &accounts); err != nil {
			return nil, gerror.Wrap(err, "解析 SMTP 配置失败")
		}
		if accounts == nil {
			accounts = []view.SMTPAccount{}
		}
		return accounts, nil
	}
	accounts := s.bootstrapSMTPFromLegacy(ctx)
	if err := s.saveKV(ctx, kvSMTPAccounts, accounts); err != nil {
		return accounts, err
	}
	return accounts, nil
}

func (s *Service) bootstrapSMTPFromLegacy(ctx context.Context) []view.SMTPAccount {
	root := s.phpRoot(ctx)
	if root == "" {
		return []view.SMTPAccount{}
	}
	content, err := os.ReadFile(filepath.Join(root, "conf", "smtp.conf.php"))
	if err != nil {
		return []view.SMTPAccount{}
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
	return accounts
}

func (s *Service) UpdateSMTPAccounts(ctx context.Context, accounts []view.SMTPAccount) error {
	if accounts == nil {
		accounts = []view.SMTPAccount{}
	}
	cleaned := make([]view.SMTPAccount, 0, len(accounts))
	for _, account := range accounts {
		if strings.TrimSpace(account.Email) == "" && strings.TrimSpace(account.Host) == "" {
			continue
		}
		cleaned = append(cleaned, account)
	}
	return s.saveKV(ctx, kvSMTPAccounts, cleaned)
}


func (s *Service) loadKV(ctx context.Context, key string) (value string, ok bool, err error) {
	var row entity.BbsKv
	if err = dao.BbsKv.Ctx(ctx).Where(do.BbsKv{K: key}).Scan(&row); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, gerror.Wrap(err, "读取配置存储失败")
	}
	if row.K == "" {
		return "", false, nil
	}
	return row.V, true, nil
}

func (s *Service) saveKV(ctx context.Context, key string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return gerror.Wrap(err, "序列化配置失败")
	}
	var row entity.BbsKv
	_ = dao.BbsKv.Ctx(ctx).Where(do.BbsKv{K: key}).Scan(&row)
	if row.K == "" {
		if _, err = dao.BbsKv.Ctx(ctx).Data(do.BbsKv{K: key, V: string(encoded), Expiry: 0}).Insert(); err != nil {
			return gerror.Wrap(err, "写入配置存储失败")
		}
		return nil
	}
	if _, err = dao.BbsKv.Ctx(ctx).Where(do.BbsKv{K: key}).Data(do.BbsKv{V: string(encoded), Expiry: 0}).Update(); err != nil {
		return gerror.Wrap(err, "更新配置存储失败")
	}
	return nil
}

func (s *Service) phpRoot(ctx context.Context) string {
	root := strings.TrimSpace(g.Cfg().MustGet(ctx, "xiuno.phpRoot", "").String())
	if root == "" {
		return ""
	}
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
	prefixes := []string{"", "bbs_"}
	if root := s.phpRoot(ctx); root != "" {
		content, err := os.ReadFile(filepath.Join(root, "conf", "conf.php"))
		if err == nil {
			pattern := regexp.MustCompile(`'cachepre'\s*=>\s*'((?:\\.|[^'])*)'`)
			for _, match := range pattern.FindAllSubmatch(content, -1) {
				if len(match) > 1 {
					prefixes = append(prefixes, phpUnescape(string(match[1])))
				}
			}
		}
	}
	keys := make([]string, 0, len(prefixes))
	seen := map[string]bool{}
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

const (
	kvActiveTheme   = "xiugo_theme"
	kvPluginEnabled = "xiugo_plugins"
)

// LoadActiveThemeID returns persisted theme id for theme.Init.
func (s *Service) LoadActiveThemeID(ctx context.Context) (string, error) {
	raw, ok, err := s.loadKV(ctx, kvActiveTheme)
	if err != nil || !ok || raw == "" {
		return "", err
	}
	var id string
	if err := json.Unmarshal([]byte(raw), &id); err != nil {
		return strings.Trim(raw, "\""), nil
	}
	return id, nil
}

// SaveActiveThemeID persists theme selection (JSON string).
func (s *Service) SaveActiveThemeID(ctx context.Context, id string) error {
	return s.saveKV(ctx, kvActiveTheme, id)
}

// LoadPluginEnabledMap returns plugin id -> enabled.
func (s *Service) LoadPluginEnabledMap(ctx context.Context) map[string]bool {
	raw, ok, err := s.loadKV(ctx, kvPluginEnabled)
	if err != nil || !ok || raw == "" {
		return map[string]bool{}
	}
	var m map[string]bool
	if json.Unmarshal([]byte(raw), &m) != nil || m == nil {
		return map[string]bool{}
	}
	return m
}

// SavePluginEnabledMap persists plugin enable flags.
func (s *Service) SavePluginEnabledMap(ctx context.Context, m map[string]bool) error {
	if m == nil {
		m = map[string]bool{}
	}
	return s.saveKV(ctx, kvPluginEnabled, m)
}
