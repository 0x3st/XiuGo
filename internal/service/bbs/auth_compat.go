package bbs

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"mime"
	"net/smtp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/0x3st/XiuGo/internal/dao"
	"github.com/0x3st/XiuGo/internal/model/do"
	"github.com/0x3st/XiuGo/internal/model/entity"
	"github.com/0x3st/XiuGo/internal/model/view"
)

func (s *Service) RegisterUser(
	ctx context.Context, email, username, password string, userIP uint32,
) (user view.User, err error) {
	email = strings.TrimSpace(email)
	username = strings.TrimSpace(username)
	if err = validatePublicUser(email, username, password); err != nil {
		return user, err
	}
	if err = s.ensureAdminUserUnique(ctx, 0, email, username, true); err != nil {
		return user, err
	}
	salt, err := xiunoSalt()
	if err != nil {
		return user, gerror.Wrap(err, "生成密码盐失败")
	}
	serverPassword, err := xiunoPassword(password, salt)
	if err != nil {
		return user, err
	}
	now := uint(time.Now().Unix())
	uid, err := dao.BbsUser.Ctx(ctx).Data(do.BbsUser{
		Email: email, Username: username, Password: serverPassword, Salt: salt,
		Gid: 101, CreateIp: userIP, CreateDate: now, LoginIp: userIP, LoginDate: now, Logins: 1,
	}).InsertAndGetId()
	if err != nil {
		return user, gerror.Wrap(err, "创建用户失败")
	}
	if err = s.SyncPHPRuntime(ctx, map[string]int{"todayusers": 1}); err != nil {
		return user, err
	}
	return view.User{Uid: uint(uid), Username: username, Gid: 101}, nil
}

func (s *Service) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var user entity.BbsUser
	if err := dao.BbsUser.Ctx(ctx).Fields(dao.BbsUser.Columns().Uid).
		Where(do.BbsUser{Email: strings.TrimSpace(email)}).Scan(&user); err != nil {
		return false, gerror.Wrap(err, "查询邮箱失败")
	}
	return user.Uid != 0, nil
}

func (s *Service) ResetPasswordByEmail(ctx context.Context, email, password, repeat string) error {
	if password == "" {
		return gerror.New("请输入新密码")
	}
	if password != repeat {
		return gerror.New("两次输入的密码不一致")
	}
	if _, err := normalizeClientPasswordHash(password); err != nil {
		return err
	}
	var user entity.BbsUser
	if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Email: strings.TrimSpace(email)}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取用户失败")
	}
	if user.Uid == 0 {
		return gerror.New("邮箱不存在")
	}
	serverPassword, err := xiunoPassword(password, user.Salt)
	if err != nil {
		return err
	}
	if _, err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: user.Uid}).
		Data(do.BbsUser{Password: serverPassword}).Update(); err != nil {
		return gerror.Wrap(err, "重设密码失败")
	}
	return nil
}

func (s *Service) SendVerificationCode(
	ctx context.Context, email, purpose string,
) (code string, err error) {
	email = strings.TrimSpace(email)
	if !adminEmailPattern.MatchString(email) {
		return "", gerror.New("邮箱格式不正确")
	}
	accounts, err := s.SMTPAccounts(ctx)
	if err != nil {
		return "", err
	}
	if len(accounts) == 0 {
		return "", gerror.New("尚未配置 SMTP 账号")
	}
	code, err = verificationCode()
	if err != nil {
		return "", gerror.Wrap(err, "生成验证码失败")
	}
	settings, err := s.SiteSettings(ctx)
	if err != nil {
		return "", err
	}
	subject := fmt.Sprintf("【%s】您的验证码：%s，该验证码5分钟内有效。", settings.Sitename, code)
	if purpose == "reset" {
		subject = fmt.Sprintf("重设密码验证码：%s - 【%s】", code, settings.Sitename)
	}
	if err = sendSMTPMessage(accounts[0], settings.Sitename, email, subject, subject); err != nil {
		return "", err
	}
	return code, nil
}

func validatePublicUser(email, username, password string) error {
	if email == "" {
		return gerror.New("请输入邮箱")
	}
	if !adminEmailPattern.MatchString(email) {
		return gerror.New("邮箱格式不正确")
	}
	if username == "" {
		return gerror.New("请输入用户名")
	}
	if utf8.RuneCountInString(username) > 16 {
		return gerror.New("用户名不能超过 16 个字符")
	}
	if !adminUsernamePattern.MatchString(username) {
		return gerror.New("用户名格式不正确")
	}
	if password == "" {
		return gerror.New("请输入密码")
	}
	_, err := normalizeClientPasswordHash(password)
	return err
}

func verificationCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(value.Int64()+100000, 10), nil
}

func sendSMTPMessage(account view.SMTPAccount, senderName, recipient, subject, body string) error {
	if account.Host == "" || account.Port == 0 || account.Email == "" {
		return gerror.New("SMTP 配置不完整")
	}
	address := account.Host + ":" + strconv.Itoa(account.Port)
	var auth smtp.Auth
	if account.User != "" {
		auth = smtp.PlainAuth("", account.User, account.Pass, account.Host)
	}
	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)
	message := strings.Join([]string{
		"From: " + senderName + " <" + account.Email + ">",
		"To: " + recipient,
		"Subject: " + encodedSubject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}, "\r\n")
	if err := smtp.SendMail(address, auth, account.Email, []string{recipient}, []byte(message)); err != nil {
		return gerror.Wrap(err, "发送验证码失败")
	}
	return nil
}
