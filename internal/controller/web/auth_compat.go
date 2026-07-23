package web

import (
	"html/template"
	"strings"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gview"
)

const (
	registerEmailSession    = "xiuno_go_register_email"
	registerCodeSession     = "xiuno_go_register_code"
	registerCodeTimeSession = "xiuno_go_register_code_time"
	resetEmailSession       = "xiuno_go_reset_email"
	resetCodeSession        = "xiuno_go_reset_code"
	resetCodeTimeSession    = "xiuno_go_reset_code_time"
	resetVerifiedSession    = "xiuno_go_reset_verified"
)

func (c *Controller) Register(r *ghttp.Request) {
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	if settings.UserCreateOn == 0 {
		r.Response.WriteStatus(http.StatusForbidden, "未开启用户注册")
		return
	}
	if r.Method == http.MethodPost {
		email := r.GetForm("email").String()
		if settings.UserCreateEmailOn != 0 && !verificationMatches(
			r, registerEmailSession, registerCodeSession, registerCodeTimeSession,
			email, r.GetForm("code").String(),
		) {
			c.writeXiunoMessage(r, "code", "邮箱验证码不正确或已过期")
			return
		}
		user, createErr := c.service.RegisterUser(
			r.Context(), email, r.GetForm("username").String(),
			r.GetForm("password").String(), remoteIPv4(r),
		)
		if createErr != nil {
			msg := createErr.Error()
			if strings.Contains(msg, "邮箱") {
				c.writeXiunoMessage(r, "email", msg)
				return
			}
			if strings.Contains(msg, "用户名") {
				c.writeXiunoMessage(r, "username", msg)
				return
			}
			if strings.Contains(msg, "密码") {
				c.writeXiunoMessage(r, "password", msg)
				return
			}
			c.writeXiunoMessage(r, -1, msg)
			return
		}
		r.Session.MustRemove(registerEmailSession)
		r.Session.MustRemove(registerCodeSession)
		r.Session.MustRemove(registerCodeTimeSession)
		r.Session.MustSet(sessionUID, user.Uid)
		c.writeXiunoMessage(r, 0, "注册成功")
		return
	}
	if err = c.renderPage(r, "pages/register.html", gview.Params{
		"Title": "创建用户", "User": c.currentUser(r),
		"ExtraJS": template.HTML(registerPageScript()),
	}); err != nil {
		c.fail(r, err)
	}
}

func registerPageScript() string {
	return `<script src="/view/js/md5.js"></script>
<script>
var jform = $('#form');
var jsubmit = $('#submit');
var jsend = $('#sendcode');
var referer = '/';
if (jsend.length) {
jsend.on('click', function() {
	jform.reset();
	jsend.button('loading');
	var postdata = jform.serialize();
	$.xpost(jsend.attr('action'), postdata, function(code, message) {
		if(code == 0) {
			$('#code').focus();
			var t = 60;
			jsend.button('发送成功 60 ');
			jsubmit.button('reset');
			var handler = setInterval(function() {
				jsend.button('发送成功 '+(--t)+' ');
				if(t == 0) {
					clearInterval(handler);
					jsend.button('reset');
				}
			}, 1000);
		} else if(code < 0) {
			$.alert(message, -1);
			jsend.button('reset');
		} else {
			jform.find('[name="'+code+'"]').alert(message).focus();
			jsend.button('reset');
		}
	});
	return false;
});
}
jform.on('submit', function() {
	var postdata = jform.serializeObject();
	jsubmit.button('loading');
	postdata.password = $.md5(postdata.password);
	$.xpost(jform.attr('action'), postdata, function(code, message) {
		if(code == 0) {
			jsubmit.button(message).delay(1000).location(referer);
		} else if(xn.is_number(code)) {
			alert(message);
			jsubmit.button('reset');
		} else {
			jform.find('[name="'+code+'"]').alert(message).focus();
			jsubmit.button('reset');
		}
	});
	return false;
});
</script>`
}

func (c *Controller) RegisterSendCode(r *ghttp.Request) {
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	if settings.UserCreateOn == 0 || settings.UserCreateEmailOn == 0 {
		c.writeXiunoMessage(r, -1, "未开启注册邮箱验证")
		return
	}
	email := r.GetForm("email").String()
	exists, err := c.service.UserExistsByEmail(r.Context(), email)
	if err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	if exists {
		c.writeXiunoMessage(r, "email", "邮箱已经被使用")
		return
	}
	code, err := c.service.SendVerificationCode(r.Context(), email, "register")
	if err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	storeVerification(r, registerEmailSession, registerCodeSession, registerCodeTimeSession, email, code)
	c.writeXiunoMessage(r, 0, "验证码发送成功")
}

func (c *Controller) ResetPassword(r *ghttp.Request) {
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	if settings.UserResetpwOn == 0 {
		r.Response.WriteStatus(http.StatusForbidden, "未开启密码找回功能")
		return
	}
	var message string
	if r.Method == http.MethodPost {
		email := r.GetForm("email").String()
		if verificationMatches(
			r, resetEmailSession, resetCodeSession, resetCodeTimeSession,
			email, r.GetForm("code").String(),
		) {
			r.Session.MustSet(resetVerifiedSession, 1)
			r.Response.RedirectTo("/reset-password/complete")
			return
		}
		message = "邮箱验证码不正确或已过期"
	}
	if err = c.renderPage(r, "pages/reset_password.html", gview.Params{
		"Title": "找回密码", "Error": message,
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) ResetPasswordSendCode(r *ghttp.Request) {
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	if settings.UserResetpwOn == 0 {
		c.writeXiunoMessage(r, -1, "未开启密码找回功能")
		return
	}
	email := r.GetForm("email").String()
	exists, err := c.service.UserExistsByEmail(r.Context(), email)
	if err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	if !exists {
		c.writeXiunoMessage(r, "email", "邮箱不存在")
		return
	}
	code, err := c.service.SendVerificationCode(r.Context(), email, "reset")
	if err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	storeVerification(r, resetEmailSession, resetCodeSession, resetCodeTimeSession, email, code)
	c.writeXiunoMessage(r, 0, "验证码发送成功")
}

func (c *Controller) ResetPasswordComplete(r *ghttp.Request) {
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	email := r.Session.MustGet(resetEmailSession).String()
	if settings.UserResetpwOn == 0 || email == "" || r.Session.MustGet(resetVerifiedSession).Int() != 1 {
		r.Response.RedirectTo("/reset-password")
		return
	}
	var message string
	if r.Method == http.MethodPost {
		if err = c.service.ResetPasswordByEmail(
			r.Context(), email, r.GetForm("password").String(), r.GetForm("password_repeat").String(),
		); err == nil {
			r.Session.MustRemove(resetEmailSession)
			r.Session.MustRemove(resetCodeSession)
			r.Session.MustRemove(resetCodeTimeSession)
			r.Session.MustRemove(resetVerifiedSession)
			r.Response.RedirectTo("/login")
			return
		}
		message = err.Error()
	}
	if err = c.renderPage(r, "pages/reset_password_complete.html", gview.Params{
		"Title": "设置新密码", "Error": message, "Email": email,
	}); err != nil {
		c.fail(r, err)
	}
}

func storeVerification(r *ghttp.Request, emailKey, codeKey, timeKey, email, code string) {
	r.Session.MustSet(emailKey, email)
	r.Session.MustSet(codeKey, code)
	r.Session.MustSet(timeKey, time.Now().Unix())
}

func verificationMatches(
	r *ghttp.Request, emailKey, codeKey, timeKey, email, code string,
) bool {
	createdAt := r.Session.MustGet(timeKey).Int64()
	return code != "" && email == r.Session.MustGet(emailKey).String() &&
		code == r.Session.MustGet(codeKey).String() && createdAt > 0 && time.Now().Unix()-createdAt <= 300
}
