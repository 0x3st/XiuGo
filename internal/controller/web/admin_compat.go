package web

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gview"

	"github.com/0x3st/XiuGo/internal/model/view"
)

const adminVerifiedAtSession = "xiuno_go_admin_verified_at"

type xiunoMessage struct {
	Code    any `json:"code"`
	Message any `json:"message"`
}

func (c *Controller) AdminCompatRoot(r *ghttp.Request) {
	target := "/admin/"
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	r.Response.RedirectTo(target)
}

func (c *Controller) AdminCompat(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 || user.Gid != 1 {
		r.Response.RedirectTo("/login")
		return
	}
	route, action, arguments := parseXiunoAdminRoute(r.URL.RawQuery)
	if route == "index" && action == "login" {
		c.adminCompatLogin(r, user)
		return
	}
	if route == "index" && action == "logout" {
		r.Session.MustRemove(adminVerifiedAtSession)
		r.Response.RedirectTo("/admin/?index-login.htm")
		return
	}
	if !c.adminCompatVerified(r) {
		c.adminCompatLogin(r, user)
		return
	}

	switch route {
	case "index":
		if action == "phpinfo" {
			c.adminCompatRuntimeInfo(r, user)
		} else {
			c.adminCompatDashboard(r, user)
		}
	case "setting":
		c.adminCompatSetting(r, user, action)
	case "forum":
		c.adminCompatForum(r, user, action, arguments)
	case "group":
		c.adminCompatGroup(r, user, action, arguments)
	case "user":
		c.adminCompatUser(r, user, action, arguments)
	case "thread":
		c.adminCompatThread(r, user, action, arguments)
	case "other":
		c.adminCompatOther(r, user, action)
	case "plugin":
		// XiuGo does not load Xiuno PHP plugins (no Hook/Overwrite runtime).
		c.writeXiunoMessage(r, -1, "XiuGo 不支持原版 PHP 插件；后续如需扩展将使用 Go 侧扩展点，而非 Xiuno 插件目录")
	default:
		c.adminCompatDashboard(r, user)
	}
}

func (c *Controller) adminCompatLogin(r *ghttp.Request, user view.User) {
	if r.Method == http.MethodPost {
		if err := c.service.VerifyAdminPassword(r.Context(), user.Uid, requestFormString(r, "password")); err != nil {
			c.writeXiunoMessage(r, "password", err.Error())
			return
		}
		r.Session.MustSet(adminVerifiedAtSession, time.Now().Unix())
		c.writeXiunoMessage(r, 0, "登录成功")
		return
	}
	c.writeAdminCompatTemplate(r, "admin_compat/index_login.html", gview.Params{
		"Title": "管理员登录", "Active": "home", "User": user,
	})
}

func (c *Controller) adminCompatDashboard(r *ghttp.Request, user view.User) {
	dashboard, err := c.service.OriginalAdminDashboard(r.Context(), requestIP(r))
	if err != nil {
		c.fail(r, err)
		return
	}
	c.writeAdminCompatTemplate(r, "admin_compat/index.html", gview.Params{
		"Title": "后台首页", "Active": "home", "User": user, "Dashboard": dashboard,
	})
}

func (c *Controller) adminCompatRuntimeInfo(r *ghttp.Request, user view.User) {
	dashboard, err := c.service.OriginalAdminDashboard(r.Context(), requestIP(r))
	if err != nil {
		c.fail(r, err)
		return
	}
	c.writeAdminCompatTemplate(r, "admin_compat/runtime_info.html", gview.Params{
		"Title": "Go 运行环境", "Active": "home", "User": user, "Dashboard": dashboard,
	})
}

func (c *Controller) adminCompatSetting(r *ghttp.Request, user view.User, action string) {
	if action == "smtp" {
		c.adminCompatSMTP(r, user)
		return
	}
	if r.Method == http.MethodPost {
		settings := view.SiteSettings{
			Sitename: r.GetForm("sitename").String(), Sitebrief: r.GetForm("sitebrief").String(),
			Runlevel: r.GetForm("runlevel").Int(), UserCreateOn: r.GetForm("user_create_on").Int(),
			UserCreateEmailOn: r.GetForm("user_create_email_on").Int(),
			UserResetpwOn:     r.GetForm("user_resetpw_on").Int(), Lang: r.GetForm("lang").String(),
		}
		if err := c.service.UpdateSiteSettings(r.Context(), settings); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, "修改成功")
		return
	}
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	c.writeAdminCompatTemplate(r, "admin_compat/setting_base.html", gview.Params{
		"Title": "站点设置", "Active": "setting", "User": user, "Settings": settings,
	})
}

func (c *Controller) adminCompatSMTP(r *ghttp.Request, user view.User) {
	if r.Method == http.MethodPost {
		var (
			emails = flatFormValues(r, "email")
			hosts  = flatFormValues(r, "host")
			ports  = flatFormValues(r, "port")
			users  = flatFormValues(r, "user")
			passes = flatFormValues(r, "pass")
		)
		accounts := make([]view.SMTPAccount, 0, len(emails))
		for index, email := range emails {
			account := view.SMTPAccount{Email: email}
			if index < len(hosts) {
				account.Host = hosts[index]
			}
			if index < len(ports) {
				account.Port, _ = strconv.Atoi(ports[index])
			}
			if index < len(users) {
				account.User = users[index]
			}
			if index < len(passes) {
				account.Pass = passes[index]
			}
			accounts = append(accounts, account)
		}
		if err := c.service.UpdateSMTPAccounts(r.Context(), accounts); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		if testEmail := strings.TrimSpace(r.GetForm("test_email").String()); testEmail != "" {
			if err := c.service.TestSMTPSend(r.Context(), testEmail); err != nil {
				c.writeXiunoMessage(r, -1, "配置已保存，但测试发送失败："+err.Error())
				return
			}
			c.writeXiunoMessage(r, 0, "保存成功，测试邮件已发送")
			return
		}
		c.writeXiunoMessage(r, 0, "保存成功")
		return
	}
	accounts, err := c.service.SMTPAccounts(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	c.writeAdminCompatTemplate(r, "admin_compat/setting_smtp.html", gview.Params{
		"Title": "SMTP 设置", "Active": "setting", "User": user, "Accounts": accounts,
	})
}

func (c *Controller) adminCompatVerified(r *ghttp.Request) bool {
	verifiedAt := r.Session.MustGet(adminVerifiedAtSession).Int64()
	if verifiedAt == 0 || time.Now().Unix()-verifiedAt > 3600 {
		r.Session.MustRemove(adminVerifiedAtSession)
		return false
	}
	if time.Now().Unix()-verifiedAt > 1800 {
		r.Session.MustSet(adminVerifiedAtSession, time.Now().Unix())
	}
	return true
}

func (c *Controller) writeAdminCompatTemplate(r *ghttp.Request, template string, params gview.Params) {
	if err := r.Response.WriteTpl(template, params); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) writeXiunoMessage(r *ghttp.Request, code, message any) {
	r.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.Response.WriteJson(xiunoMessage{Code: code, Message: message})
}

func parseXiunoAdminRoute(rawQuery string) (route, action string, arguments []string) {
	if rawQuery == "" {
		return "index", "", nil
	}
	token := strings.Split(rawQuery, "&")[0]
	token = strings.TrimSuffix(token, ".htm")
	parts := strings.Split(token, "-")
	for index := 2; index < len(parts); index++ {
		encoded := strings.ReplaceAll(parts[index], "_", "%")
		if decoded, err := url.QueryUnescape(encoded); err == nil {
			parts[index] = decoded
		}
	}
	if len(parts) == 0 || parts[0] == "" {
		return "index", "", nil
	}
	route = parts[0]
	if len(parts) > 1 {
		action = parts[1]
	}
	if len(parts) > 2 {
		arguments = parts[2:]
	}
	return route, action, arguments
}

func adminArgumentUint(arguments []string, index int) uint {
	if index >= len(arguments) {
		return 0
	}
	value, _ := strconv.ParseUint(arguments[index], 10, 64)
	return uint(value)
}

func requestIP(r *ghttp.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
