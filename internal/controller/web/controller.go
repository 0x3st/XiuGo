package web

import (
	"fmt"
	htmltemplate "html/template"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gview"

	"github.com/0x3st/XiuGo/internal/model/view"
	"github.com/0x3st/XiuGo/internal/plugin"
	"github.com/0x3st/XiuGo/internal/theme"
	bbsService "github.com/0x3st/XiuGo/internal/service/bbs"
)

const sessionUID = "xiuno_go_uid"

type Controller struct {
	service *bbsService.Service
}

func New() *Controller {
	return &Controller{service: bbsService.New()}
}

func (c *Controller) Bind(group *ghttp.RouterGroup) {
	group.Middleware(c.EnforceRunlevel)
	group.GET("/", c.Home)
	group.GET("/forum/:fid", c.Forum)
	group.GET("/thread/:tid", c.Thread)
	group.POST("/thread/:tid/reply", c.Reply)
	group.GET("/thread/:tid/reply", c.AdvancedReply)
	group.ALL("/post/:pid/edit", c.EditPost)
	group.POST("/post/:pid/delete", c.DeletePost)
	group.POST("/attach/create", c.AttachCreate)
	group.POST("/attach/:aid/delete", c.AttachDelete)
	group.GET("/attach/:aid/download", c.AttachDownload)
	group.ALL("/login", c.Login)
	group.ALL("/logout", c.Logout)
	group.ALL("/register", c.Register)
	group.POST("/register/send-code", c.RegisterSendCode)
	group.ALL("/reset-password", c.ResetPassword)
	group.POST("/reset-password/send-code", c.ResetPasswordSendCode)
	group.ALL("/reset-password/complete", c.ResetPasswordComplete)
	group.ALL("/thread/create", c.CreateThread)
	group.GET("/user/:uid", c.UserProfile)
	group.GET("/user/:uid/threads", c.UserThreads)
	group.GET("/my", c.MyProfile)
	group.GET("/my/threads", c.MyThreads)
	group.ALL("/my/password", c.MyPassword)
	group.ALL("/my/avatar", c.MyAvatar)
	group.ALL("/admin", c.AdminCompat)
	group.ALL("/admin/login", c.AdminLogin)
	group.ALL("/admin/settings", c.AdminSettings)
	group.ALL("/admin/settings/smtp", c.AdminSettingsSMTP)
	group.ALL("/admin/maintenance", c.AdminMaintenance)
	group.GET("/admin/runtime", c.AdminRuntime)
	group.ALL("/admin/themes", c.AdminThemes)
	group.ALL("/admin/plugins", c.AdminPlugins)
	group.GET("/admin/threads", c.AdminThreads)
	group.POST("/admin/thread/:tid/close", c.AdminThreadClose)
	group.POST("/admin/thread/:tid/delete", c.AdminThreadDelete)
	group.POST("/thread/:tid/top", c.ThreadTop)
	group.GET("/mod/top", c.ModTopForm)
	group.POST("/mod/top", c.ModTop)
	group.GET("/mod/close", c.ModCloseForm)
	group.POST("/mod/close", c.ModClose)
	group.GET("/mod/delete", c.ModDeleteForm)
	group.POST("/mod/delete", c.ModDelete)
	group.GET("/mod/move", c.ModMoveForm)
	group.POST("/mod/move", c.ModMove)
	group.GET("/admin/users", c.AdminUsers)
	group.POST("/admin/user/:uid/group", c.AdminUserGroup)
	group.GET("/admin/forums", c.AdminForums)
	group.ALL("/admin/forum/create", c.AdminForumCreate)
	group.ALL("/admin/forum/:fid/edit", c.AdminForumEdit)
	group.POST("/admin/forum/:fid/delete", c.AdminForumDelete)
	group.GET("/admin/groups", c.AdminGroups)
	group.ALL("/admin/group/:gid/edit", c.AdminGroupEdit)
}

// EnforceRunlevel mirrors model/misc.func.php:check_runlevel() from the
// original Xiuno runtime. Login, logout, registration and password recovery
// stay reachable so a visitor can become an authenticated user.
func (c *Controller) EnforceRunlevel(r *ghttp.Request) {
	path := r.URL.Path
	if strings.HasPrefix(path, "/admin") || runlevelExemptPath(path) {
		r.Middleware.Next()
		return
	}
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	user := c.currentUser(r)
	if user.Gid == 1 || runlevelAllows(settings.Runlevel, user.Uid != 0, r.Method) {
		r.Middleware.Next()
		return
	}
	reason := runlevelReason(settings)
	r.Response.WriteStatus(http.StatusForbidden, reason)
}

func runlevelExemptPath(path string) bool {
	return path == "/login" || path == "/logout" || strings.HasPrefix(path, "/register") ||
		strings.HasPrefix(path, "/reset-password")
}

func runlevelAllows(runlevel int, loggedIn bool, method string) bool {
	readOnly := method == http.MethodGet || method == http.MethodHead
	switch runlevel {
	case 0, 1:
		return false
	case 2:
		return loggedIn && readOnly
	case 3:
		return loggedIn
	case 4:
		return readOnly
	default:
		return true
	}
}

func runlevelReason(settings view.SiteSettings) string {
	switch settings.Runlevel {
	case 0:
		if settings.RunlevelReason != "" {
			return settings.RunlevelReason
		}
		return "站点已关闭"
	case 1:
		return "当前站点设置状态：管理员可读写"
	case 2:
		return "当前站点设置状态：会员只读"
	case 3:
		return "当前站点设置状态：会员可读写"
	case 4:
		return "当前站点设置状态：所有人只读"
	default:
		return "当前请求不可用"
	}
}

func (c *Controller) renderPage(r *ghttp.Request, template string, params gview.Params) error {
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		return err
	}
	params["Site"] = settings
	// Do not globally rewrite page titles: thread subjects may contain legacy product names.
	// Shared layout context for original-style header/footer.
	user, _ := params["User"].(view.User)
	if _, ok := params["NavForums"]; !ok {
		if forums, _, _, homeErr := c.service.Home(r.Context(), user); homeErr == nil {
			params["NavForums"] = forums
			params["ForumArrJSON"] = forumArrJSON(forums)
		} else {
			params["NavForums"] = []view.ForumSummary{}
			params["ForumArrJSON"] = "{}"
		}
	} else if forums, ok := params["NavForums"].([]view.ForumSummary); ok {
		params["ForumArrJSON"] = forumArrJSON(forums)
	}
	if _, ok := params["ActiveFid"]; !ok {
		params["ActiveFid"] = 0
	}
	if _, ok := params["HeaderDescription"]; !ok {
		params["HeaderDescription"] = settings.Sitebrief
	}
	// Theme stylesheets (active skin).
	params["ThemeID"] = theme.Global().ActiveID()
	params["ThemeCSS"] = theme.Global().Stylesheets()
	// Plugin page hook
	ev := &plugin.PageRenderEvent{
		Template: template,
		Params:   map[string]any{},
	}
	_ = plugin.Global().Fire(r.Context(), plugin.HookPageRender, ev)
	params["PluginFooterHTML"] = htmltemplate.HTML(ev.FooterHTML)
	if len(ev.ExtraCSS) > 0 {
		params["PluginCSS"] = ev.ExtraCSS
	}
	if len(ev.ExtraJS) > 0 {
		params["PluginJS"] = ev.ExtraJS
	}
	return r.Response.WriteTpl(template, params)
}

func forumArrJSON(forums []view.ForumSummary) string {
	// Mirrors PHP $forumarr used by bbs.js: {fid: name, ...}
	parts := make([]string, 0, len(forums))
	for _, forum := range forums {
		name := strings.ReplaceAll(forum.Name, `\`, `\`)
		name = strings.ReplaceAll(name, `"`, `"`)
		parts = append(parts, fmt.Sprintf(`"%d":"%s"`, forum.Fid, name))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func (c *Controller) Home(r *ghttp.Request) {
	user := c.currentUser(r)
	page := r.GetQuery("page").Int()
	if page < 1 {
		page = 1
	}
	forums, threads, stats, pager, err := c.service.HomePaged(r.Context(), user, page)
	if err != nil {
		c.fail(r, err)
		return
	}
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	haveAllowTop, annErr := c.service.AnnotateThreadModFlags(r.Context(), threads, user)
	if annErr != nil {
		c.fail(r, annErr)
		return
	}
	if err = c.renderPage(r, "pages/home.html", gview.Params{
		"Title": settings.Sitename, "Forums": forums, "NavForums": forums, "Threads": threads,
		"Stats": stats, "User": user, "ActiveFid": 0, "HaveAllowTop": haveAllowTop,
		"Pagination": pager,
		"ExtraJS": htmltemplate.HTML(`<script>$('li[data-active="fid-0"]').addClass('active');</script>`),
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) Forum(r *ghttp.Request) {
	fid := r.GetRouter("fid").Uint()
	user := c.currentUser(r)
	orderby := r.GetQuery("orderby", "lastpid").String()
	if orderby != "tid" {
		orderby = "lastpid"
	}
	page := r.GetQuery("page").Int()
	if page < 1 {
		page = 1
	}
	forum, threads, pager, err := c.service.ForumOrdered(r.Context(), fid, user, orderby, page)
	if err != nil {
		c.fail(r, err)
		return
	}
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	haveAllowTop, annErr := c.service.AnnotateThreadModFlags(r.Context(), threads, user)
	if annErr != nil {
		c.fail(r, annErr)
		return
	}
	if err = c.renderPage(r, "pages/forum.html", gview.Params{
		"Title": forum.Name + "-" + settings.Sitename, "Forum": forum,
		"Threads": threads, "User": user, "ActiveFid": fid, "Orderby": orderby, "HaveAllowTop": haveAllowTop,
		"Pagination": pager,
		"MobileTitle": forum.Name, "MobileLink": "/forum/" + strconv.FormatUint(uint64(fid), 10),
		"ExtraJS": htmltemplate.HTML(fmt.Sprintf(`<script>$('li[data-active="fid-%d"]').addClass('active');</script>`, fid)),
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) Thread(r *ghttp.Request) {
	tid := r.GetRouter("tid").Uint()
	user := c.currentUser(r)
	pageNum := r.GetQuery("page").Int()
	if pageNum < 1 {
		pageNum = 1
	}
	keyword := r.GetQuery("keyword").String()
	page, err := c.service.Thread(r.Context(), tid, user, pageNum, keyword)
	if err != nil {
		c.fail(r, err)
		return
	}
	settings, err := c.service.SiteSettings(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	fid := uint(page.Thread.Fid)
	extra := threadPageExtraJS(fid, page.Thread.Closed != 0, user.Gid)
	if err = c.renderPage(r, "pages/thread.html", gview.Params{
		"Title": page.Thread.Subject + "-" + page.Thread.ForumName + "-" + settings.Sitename,
		"Page": page, "User": user, "ActiveFid": fid,
		"MobileTitle": page.Thread.Subject,
		"MobileLink": "/thread/" + strconv.FormatUint(uint64(page.Thread.Tid), 10),
		"ExtraJS": htmltemplate.HTML(extra),
	}); err != nil {
		c.fail(r, err)
	}
}


func postFormExtraJS(location string, fid uint, isCreate bool) string {
	loc := strings.ReplaceAll(location, `\`, `\\`)
	loc = strings.ReplaceAll(loc, `'`, `\'`)
	return fmt.Sprintf(`
<script>
var jform = $('#form');
var jsubmit = $('#submit');
var jfid = jform.find('select[name="fid"]');
jform.on('submit', function() {
	jform.reset();
	jsubmit.button('loading');
	var postdata = jform.serialize();
	$.xpost(jform.attr('action'), postdata, function(code, message) {
		if(code == 0) {
			$.alert(message);
			var loc = '%s';
			if(loc === '__forum__' && jfid.length) { loc = '/forum/' + jfid.val(); }
			jsubmit.button(message).delay(1000).location(loc);
		} else if(xn.is_number(code)) {
			alert(message);
			jsubmit.button('reset');
		} else {
			$.alert(message);
			jsubmit.button('reset');
		}
	});
	return false;
});

var jattachparent = $('.attachlist_parent');
$('#addattach').on('change', function(e) {
	var files = xn.get_files_from_event(e);
	if (!files) return;
	if (!jattachparent.find('.attachlist').length) {
		jattachparent.append('<fieldset class="fieldset"><legend>上传的附件：</legend><ul class="attachlist"></ul></fieldset>');
	}
	var jprogress = jattachparent.find('.progress');
	if(!jprogress.length) {
		jprogress = $('<div class="progress"><div class="progress-bar" role="progressbar" style="width: 25%%;" aria-valuenow="25" aria-valuemin="0" aria-valuemax="100">25%%</div></div>').appendTo(jattachparent);
	}
	var jprogressbar = jprogress.find('.progress-bar');
	$.each_sync(files, function(i, callback) {
		var file = files[i];
		xn.upload_file(file, '/attach/create', {is_image: 0}, function(code, message) {
			if (code != 0) return $.alert(message);
			var aid = message.aid;
			$('.attachlist').append('<li aid="' + aid + '"><a href="' + message.url + '" target="_blank"><i class="icon filetype ' + message.filetype + '"></i> ' + message.orgfilename + '</a> <a href="javascript:void(0);" class="delete ml-2"><i class="icon-remove"></i> 删除</a></li>');
			callback();
			jprogress.hide();
		}, function(percent) {
			percent = xn.intval(percent);
			jprogressbar.css('width', percent+'%%');
			jprogressbar.text(percent+'%%');
			jprogress.show();
		});
	});
});
jattachparent.on('click', 'a.delete', function() {
	var jlink = $(this);
	var jli = jlink.parents('li');
	var aid = jli.attr('aid');
	if(!window.confirm(lang.confirm_delete || '确定删除吗？')) return false;
	$.xpost('/attach/' + aid + '/delete', function(code, message) {
		if(code != 0) return $.alert(message);
		jlink.parent().remove();
	});
	return false;
});
if (jfid.length) { jfid.val('%d'); }
$('li[data-active="fid-%d"]').addClass('active');
</script>
`, loc, fid, fid)
}

func threadPageExtraJS(fid uint, closed bool, gid uint) string {
	js := fmt.Sprintf(`
<script>
var jform = $('#quick_reply_form');
var jsubmit = $('#submit');
if (jform.length) {
jform.on('submit', function() {
	jform.reset();
	jsubmit.button('loading');
	var postdata = jform.serialize();
	$.xpost(jform.attr('action'), postdata, function(code, message) {
		if(code == 0) {
			var s = '<ul>'+message+'</ul>';
			var jli = $(s).find('li');
			if(jli.length && $('.postlist > .post').length) {
				jli.insertBefore($('.postlist > .post').last());
				jsubmit.button('reset');
				$('#message').val('');
				// bump counters
				var jposts = $('.posts');
				if(jposts.length) { jposts.text(xn.intval(jposts.text())+1); }
				var jfloor = $('#newfloor');
				if(jfloor.length) { jfloor.text(xn.intval(jfloor.text())+1); }
			} else {
				window.location.reload();
			}
		} else if(xn.is_number(code)) {
			$.alert(message);
			jsubmit.button('reset');
		} else {
			jform.find('[name="'+code+'"]').alert(message).focus();
			jsubmit.button('reset');
		}
	});
	return false;
});
}
$('.post_reply').on('click', function() {
	var pid = $(this).data('pid');
	$('input[name="quotepid"]').val(pid);
	$('#message').focus();
});
$('.post_delete').on('click', function() {
	var href = $(this).data('href');
	if(!href) return;
	if(!window.confirm('确定删除吗？')) return;
	$.xpost(href, {}, function(code, message) {
		if(code == 0) {
			window.location.href = message || '/';
		} else {
			alert(message);
		}
	});
});
var jmessage = $('#message');
if (jmessage.length) {
	jmessage.on('focus', function() {if(jmessage.t) { clearTimeout(jmessage.t); jmessage.t = null; } jmessage.css('height', '8rem'); });
	jmessage.on('blur', function() {jmessage.t = setTimeout(function() { jmessage.css('height', '2.5rem');}, 1000); });
}
$('li[data-active="fid-%d"]').addClass('active');
</script>
`, fid)
	if closed && (gid == 0 || gid > 5) {
		js += `<script>if($('#message').length){$('#message').val('主题已经关闭').attr('readonly','readonly');}</script>`
	}
	return js
}

func (c *Controller) EditPost(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		if wantsXiunoJSON(r) {
			c.writeXiunoMessage(r, -1, "请先登录")
			return
		}
		r.Response.RedirectTo("/login")
		return
	}
	pid := r.GetRouter("pid").Uint()
	if r.Method == http.MethodPost {
		pending := c.pendingAttachments(r)
		if _, err := c.service.UpdatePost(
			r.Context(), pid, user.Uid, r.GetForm("subject").String(), r.GetForm("message").String(),
			r.GetForm("doctype").Int(), pending, r.GetForm("fid").Uint(),
		); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.clearPendingAttachments(r)
		c.writeXiunoMessage(r, 0, "编辑成功")
		return
	}
	post, err := c.service.PostForEdit(r.Context(), pid, user.Uid)
	if err != nil {
		c.fail(r, err)
		return
	}
	loc := "/thread/" + strconv.FormatUint(uint64(post.Tid), 10)
	forums, _ := c.service.ForumsAllowThread(r.Context(), user)
	if err = c.renderPage(r, "pages/post_edit.html", gview.Params{
		"Title": "编辑帖子", "Post": post, "User": user, "Forums": forums,
		"Pending": c.pendingAttachments(r), "ActiveFid": post.Fid,
		"ExtraJS": htmltemplate.HTML(postFormExtraJS(loc, uint(post.Fid), false)),
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) DeletePost(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		if wantsXiunoJSON(r) {
			c.writeXiunoMessage(r, -1, "请先登录")
			return
		}
		r.Response.RedirectTo("/login")
		return
	}
	tid, err := c.service.DeletePost(r.Context(), r.GetRouter("pid").Uint(), user.Uid)
	if err != nil {
		if wantsXiunoJSON(r) {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.fail(r, err)
		return
	}
	dest := "/"
	if tid != 0 {
		dest = "/thread/" + strconv.FormatUint(uint64(tid), 10)
	}
	if wantsXiunoJSON(r) {
		c.writeXiunoMessage(r, 0, dest)
		return
	}
	r.Response.RedirectTo(dest)
}

func (c *Controller) Reply(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		if wantsXiunoJSON(r) {
			c.writeXiunoMessage(r, -1, "请先登录")
			return
		}
		r.Response.RedirectTo("/login")
		return
	}
	tid := r.GetRouter("tid").Uint()
	pid, err := c.service.Reply(
		r.Context(), tid, user.Uid, r.GetForm("message").String(), r.GetForm("doctype").Int(),
		r.GetForm("quotepid").Uint(), remoteIPv4(r), c.pendingAttachments(r),
	)
	if err != nil {
		if wantsXiunoJSON(r) || requestFormString(r, "return_html") != "" {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.fail(r, err)
		return
	}
	c.clearPendingAttachments(r)
	if requestFormString(r, "return_html") != "" {
		post, postErr := c.service.PostViewForPID(r.Context(), pid, user)
		if postErr != nil {
			c.writeXiunoMessage(r, -1, postErr.Error())
			return
		}
		params := gview.Params{
			"Pid": post.Pid, "Tid": post.Tid, "Uid": post.Uid, "Username": post.Username,
			"AvatarURL": post.AvatarURL, "CreateTime": post.CreateTime,
			"MessageFmt": htmltemplate.HTML(post.MessageFmt),
			"Floor": post.Floor, "CanQuote": post.CanQuote, "CanEdit": post.CanEdit, "CanDelete": post.CanDelete,
			"Files": post.Files,
		}
		html, renderErr := r.Response.ParseTpl("partials/post_list_item.html", params)
		if renderErr != nil {
			c.writeXiunoMessage(r, -1, renderErr.Error())
			return
		}
		c.writeXiunoMessage(r, 0, html)
		return
	}
	if wantsXiunoJSON(r) {
		c.writeXiunoMessage(r, 0, "回帖成功")
		return
	}
	r.Response.RedirectTo("/thread/" + strconv.FormatUint(uint64(tid), 10))
}


func parseTidarr(r *ghttp.Request) []uint {
	_ = r.Request.ParseForm()
	rawValues := append([]string{}, r.Request.PostForm["tidarr"]...)
	rawValues = append(rawValues, r.Request.PostForm["tidarr[]"]...)
	if len(rawValues) == 0 {
		rawValues = append(rawValues, r.GetForm("tidarr").Strings()...)
		rawValues = append(rawValues, r.GetForm("tidarr[]").Strings()...)
	}
	seen := map[uint]bool{}
	tids := make([]uint, 0, len(rawValues))
	for _, value := range rawValues {
		parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		if err != nil || parsed == 0 || seen[uint(parsed)] {
			continue
		}
		seen[uint(parsed)] = true
		tids = append(tids, uint(parsed))
	}
	return tids
}

func requestFormString(r *ghttp.Request, key string) string {
	if v := strings.TrimSpace(r.GetForm(key).String()); v != "" {
		return v
	}
	_ = r.Request.ParseForm()
	if v := strings.TrimSpace(r.Request.PostFormValue(key)); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Request.FormValue(key)); v != "" {
		return v
	}
	return strings.TrimSpace(r.Get(key).String())
}

func wantsXiunoJSON(r *ghttp.Request) bool {

	if r.IsAjaxRequest() || r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		return true
	}
	if requestFormString(r, "return_html") != "" {
		return true
	}
	if r.Method == http.MethodPost && r.Header.Get("Sec-Fetch-Mode") != "navigate" {
		return true
	}
	return false
}

func (c *Controller) Login(r *ghttp.Request) {
	if r.Method == http.MethodPost {
		account := r.GetForm("email").String()
		if account == "" {
			account = r.GetForm("account").String()
		}
		user, err := c.service.Authenticate(
			r.Context(), account, r.GetForm("password").String(), remoteIPv4(r),
		)
		if err == nil {
			r.Session.MustSet(sessionUID, user.Uid)
			c.writeXiunoMessage(r, 0, "登录成功")
			return
		}
		msg := err.Error()
		if strings.Contains(msg, "密码") {
			c.writeXiunoMessage(r, "password", msg)
			return
		}
		if strings.Contains(msg, "账号") || strings.Contains(msg, "邮箱") || strings.Contains(msg, "用户") {
			c.writeXiunoMessage(r, "email", msg)
			return
		}
		c.writeXiunoMessage(r, -1, msg)
		return
	}
	referer := r.GetQuery("referer", "/").String()
	if referer == "" {
		referer = "/"
	}
	if err := c.renderPage(r, "pages/login.html", gview.Params{
		"Title":   "用户登录",
		"User":    c.currentUser(r),
		"Referer": referer,
		"ExtraJS": htmltemplate.HTML(loginPageScript(referer)),
	}); err != nil {
		c.fail(r, err)
	}
}

func loginPageScript(referer string) string {
	ref := strings.ReplaceAll(referer, `\`, `\`)
	ref = strings.ReplaceAll(ref, `'`, `\'`)
	return fmt.Sprintf(`<script src="/view/js/md5.js"></script>
<script>
var jform = $('#form');
var jsubmit = $('#submit');
var referer = '%s';
jform.on('submit', function() {
	jform.reset();
	jsubmit.button('loading');
	var postdata = jform.serializeObject();
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
</script>`, ref)
}

func (c *Controller) Logout(r *ghttp.Request) {
	r.Session.MustRemove(sessionUID)
	r.Response.RedirectTo("/")
}

func (c *Controller) CreateThread(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		if wantsXiunoJSON(r) {
			c.writeXiunoMessage(r, -1, "请先登录")
			return
		}
		r.Response.RedirectTo("/login")
		return
	}
	fid := r.GetForm("fid").Uint()
	if fid == 0 {
		fid = r.Get("fid", 1).Uint()
	}
	if r.Method == http.MethodPost {
		pending := c.pendingAttachments(r)
		if _, err := c.service.CreateThread(
			r.Context(), fid, user.Uid,
			r.GetForm("subject").String(), r.GetForm("message").String(),
			r.GetForm("doctype").Int(), remoteIPv4(r), pending,
		); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.clearPendingAttachments(r)
		// Original xpost success then client redirects using form JS location.
		c.writeXiunoMessage(r, 0, "发帖成功")
		return
	}
	forums, err := c.service.ForumsAllowThread(r.Context(), user)
	if err != nil {
		c.fail(r, err)
		return
	}
	if err = c.renderPage(r, "pages/create.html", gview.Params{
		"Title": "发帖", "User": user, "Forums": forums, "ActiveFid": fid,
		"IsFirst": true, "FormTitle": "发表主题", "FormAction": "/thread/create",
		"SubmitText": "发表主题", "Doctype": 1, "Pending": c.pendingAttachments(r),
		"ExtraJS": htmltemplate.HTML(postFormExtraJS("__forum__", fid, true)),
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) UserProfile(r *ghttp.Request) {
	user := c.currentUser(r)
	profile, err := c.service.Profile(r.Context(), r.GetRouter("uid").Uint())
	if err != nil {
		c.fail(r, err)
		return
	}
	if err = c.renderPage(r, "pages/user_profile.html", gview.Params{
		"Title": profile.Username, "Profile": profile, "User": user, "Mine": false,
		"ExtraJS": htmltemplate.HTML(`<script>$('a[data-active="menu-user"]').addClass('active');$('a[data-active="user-profile"]').addClass('active');</script>`),
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) UserThreads(r *ghttp.Request) {
	user := c.currentUser(r)
	uid := r.GetRouter("uid").Uint()
	profile, err := c.service.Profile(r.Context(), uid)
	if err != nil {
		c.fail(r, err)
		return
	}
	page := r.GetQuery("page").Int()
	if page < 1 {
		page = 1
	}
	threads, pager, err := c.service.UserThreadsPaged(r.Context(), uid, page)
	if err != nil {
		c.fail(r, err)
		return
	}
	// Filter unreadable forums for viewers other than owner.
	if user.Uid != uid {
		visible := threads[:0]
		for _, th := range threads {
			ok, permErr := c.service.ForumReadable(r.Context(), uint(th.Fid), user.Gid)
			if permErr != nil {
				c.fail(r, permErr)
				return
			}
			if ok {
				visible = append(visible, th)
			}
		}
		threads = visible
	}
	if err = c.renderPage(r, "pages/user_threads.html", gview.Params{
		"Title": profile.Username, "Profile": profile,
		"Threads": threads, "User": user, "Mine": false, "Pagination": pager,
		"ExtraJS": htmltemplate.HTML(`<script>$('a[data-active="menu-user-thread"]').addClass('active');$('a[data-active="user-thread"]').addClass('active');</script>`),
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) MyProfile(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		r.Response.RedirectTo("/login")
		return
	}
	profile, err := c.service.Profile(r.Context(), user.Uid)
	if err != nil {
		c.fail(r, err)
		return
	}
	if err = c.renderPage(r, "pages/user_profile.html", gview.Params{
		"Title": "个人中心", "Profile": profile, "User": user, "Mine": true,
		"ExtraJS": htmltemplate.HTML(`<script>$('a[data-active="menu-my"]').addClass('active');$('a[data-active="my-profile"]').addClass('active');</script>`),
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) MyThreads(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		r.Response.RedirectTo("/login")
		return
	}
	profile, err := c.service.Profile(r.Context(), user.Uid)
	if err != nil {
		c.fail(r, err)
		return
	}
	page := r.GetQuery("page").Int()
	if page < 1 {
		page = 1
	}
	threads, pager, err := c.service.UserThreadsPaged(r.Context(), user.Uid, page)
	if err != nil {
		c.fail(r, err)
		return
	}
	// My threads: rewrite pager links to /my/threads
	if pager.HTML != "" {
		pager.HTML = strings.ReplaceAll(pager.HTML, "/user/"+strconv.FormatUint(uint64(user.Uid), 10)+"/threads", "/my/threads")
	}
	settings, _ := c.service.SiteSettings(r.Context())
	if err = c.renderPage(r, "pages/user_threads.html", gview.Params{
		"Title": settings.Sitename, "Profile": profile, "Threads": threads, "User": user, "Mine": true, "Pagination": pager,
		"ExtraJS": htmltemplate.HTML(`<script>$('a[data-active="menu-my-thread"]').addClass('active');$('a[data-active="my-thread"]').addClass('active');</script>`),
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) MyPassword(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		if wantsXiunoJSON(r) {
			c.writeXiunoMessage(r, -1, "请先登录")
			return
		}
		r.Response.RedirectTo("/login")
		return
	}
	if r.Method == http.MethodPost {
		err := c.service.ChangePassword(
			r.Context(), user.Uid, r.GetForm("password_old").String(),
			r.GetForm("password_new").String(), r.GetForm("password_new_repeat").String(),
		)
		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "旧密码") {
				c.writeXiunoMessage(r, "password_old", msg)
				return
			}
			c.writeXiunoMessage(r, -1, msg)
			return
		}
		c.writeXiunoMessage(r, 0, "修改成功")
		return
	}
	settings, _ := c.service.SiteSettings(r.Context())
	if err := c.renderPage(r, "pages/my_password.html", gview.Params{
		"Title": settings.Sitename, "User": user,
		"ExtraJS": htmltemplate.HTML(`<script src="/view/js/md5.js"></script>
<script>
$('a[data-active="menu-my"]').addClass('active');
$('a[data-active="my-password"]').addClass('active');
var jform = $('#form');
var jsubmit = $('#submit');
jform.on('submit', function() {
	jform.reset();
	jsubmit.button('loading');
	var postdata = jform.serializeObject();
	postdata.password_old = $.md5(postdata.password_old);
	postdata.password_new = $.md5(postdata.password_new);
	postdata.password_new_repeat = $.md5(postdata.password_new_repeat);
	$.xpost(jform.attr('action'), postdata, function(code, message) {
		if(code == 0) {
			$.alert(message);
			jsubmit.button(message).delay(3000).button('reset');
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
</script>`),
	}); err != nil {
		c.fail(r, err)
	}
}


func (c *Controller) MyAvatar(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		if wantsXiunoJSON(r) {
			c.writeXiunoMessage(r, -1, "请先登录")
			return
		}
		r.Response.RedirectTo("/login")
		return
	}
	if r.Method == http.MethodPost {
		url, err := c.service.UpdateAvatar(r.Context(), user.Uid, r.GetForm("data").String())
		if err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, map[string]any{"url": url})
		return
	}
	settings, _ := c.service.SiteSettings(r.Context())
	if err := c.renderPage(r, "pages/my_avatar.html", gview.Params{
		"Title": settings.Sitename, "User": user,
		"ExtraJS": htmltemplate.HTML(`<script>
$('a[data-active="menu-my"]').addClass('active');
$('a[data-active="my-avatar"]').addClass('active');
var javatar_upload = $('#avatar_upload');
var jprogress = $('#avatar_progress');
var jimg = $('#avatar_img');
jprogress.hide();
javatar_upload.on('change', function(e) {
	var files = xn.get_files_from_event(e);
	if(!files || !files.length) return;
	xn.upload_file(
		files[0],
		'/my/avatar',
		{width: 128, height: 128, action: 'clip', filetype: 'png'},
		function(code, message) {
			if(code == 0) {
				jimg.attr('src', message.url+'?'+Math.random());
				$.alert('成功');
				jprogress.delay(1000).hide();
			} else {
				$.alert(message);
			}
		},
		function(percent) {
			jprogress.show();
			jprogress.width(percent+'%');
		}
	);
});
</script>`),
	}); err != nil {
		c.fail(r, err)
	}
}


func (c *Controller) AdminThreads(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	keyword := r.GetQuery("keyword").String()
	threads, err := c.service.AdminThreads(r.Context(), keyword)
	if err != nil {
		c.fail(r, err)
		return
	}
	if err = r.Response.WriteTpl("admin/threads.html", gview.Params{
		"Title": "主题管理", "Threads": threads, "Keyword": keyword, "User": user,
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) AdminThreadClose(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	tid := r.GetRouter("tid").Uint()
	if err := c.service.SetThreadClosed(r.Context(), tid, user.Uid, r.GetForm("closed").Uint()); err != nil {
		c.fail(r, err)
		return
	}
	r.Response.RedirectTo("/admin/threads")
}

func (c *Controller) AdminThreadDelete(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	if err := c.service.DeleteThread(r.Context(), r.GetRouter("tid").Uint(), user.Uid); err != nil {
		c.fail(r, err)
		return
	}
	r.Response.RedirectTo("/admin/threads")
}

// ThreadTop is the single-thread pin control used on the thread detail page.
func (c *Controller) ThreadTop(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		r.Response.RedirectTo("/login")
		return
	}
	tid := r.GetRouter("tid").Uint()
	if err := c.service.SetThreadTop(r.Context(), tid, user.Uid, r.GetForm("top").Int()); err != nil {
		c.fail(r, err)
		return
	}
	r.Response.RedirectTo("/thread/" + strconv.FormatUint(uint64(tid), 10))
}

// ModTop mirrors route/mod.php?action=top for multi-select lists.

func (c *Controller) ModTopForm(r *ghttp.Request) {
	if c.currentUser(r).Uid == 0 {
		r.Response.WriteStatus(http.StatusForbidden, "请先登录")
		return
	}
	_ = r.Response.WriteTpl("mod/top.html", gview.Params{})
}

func (c *Controller) ModCloseForm(r *ghttp.Request) {
	if c.currentUser(r).Uid == 0 {
		r.Response.WriteStatus(http.StatusForbidden, "请先登录")
		return
	}
	_ = r.Response.WriteTpl("mod/close.html", gview.Params{})
}

func (c *Controller) ModDeleteForm(r *ghttp.Request) {
	if c.currentUser(r).Uid == 0 {
		r.Response.WriteStatus(http.StatusForbidden, "请先登录")
		return
	}
	_ = r.Response.WriteTpl("mod/delete.html", gview.Params{})
}

func (c *Controller) ModMoveForm(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		r.Response.WriteStatus(http.StatusForbidden, "请先登录")
		return
	}
	forums, _ := c.service.ForumsAllowThread(r.Context(), user)
	_ = r.Response.WriteTpl("mod/move.html", gview.Params{"Forums": forums})
}

func (c *Controller) ModClose(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		c.writeXiunoMessage(r, -1, "请先登录")
		return
	}
	if err := c.service.SetThreadsClosed(r.Context(), parseTidarr(r), user.Uid, r.GetForm("close").Uint()); err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	c.writeXiunoMessage(r, 0, "设置完成")
}

func (c *Controller) ModDelete(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		c.writeXiunoMessage(r, -1, "请先登录")
		return
	}
	if err := c.service.DeleteThreads(r.Context(), parseTidarr(r), user.Uid); err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	c.writeXiunoMessage(r, 0, "删除完成")
}

func (c *Controller) ModMove(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		c.writeXiunoMessage(r, -1, "请先登录")
		return
	}
	if err := c.service.MoveThreads(r.Context(), parseTidarr(r), user.Uid, r.GetForm("newfid").Uint()); err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	c.writeXiunoMessage(r, 0, "移动完成")
}

func (c *Controller) AdvancedReply(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		r.Response.RedirectTo("/login")
		return
	}
	tid := r.GetRouter("tid").Uint()
	page, err := c.service.Thread(r.Context(), tid, user, 1, "")
	if err != nil {
		c.fail(r, err)
		return
	}
	quotePID := r.GetQuery("quotepid").Uint()
	if err = c.renderPage(r, "pages/create.html", gview.Params{
		"Title": "回帖", "User": user, "IsFirst": false,
		"FormTitle": "回帖", "FormAction": "/thread/" + strconv.FormatUint(uint64(tid), 10) + "/reply",
		"SubmitText": "回帖", "Doctype": 1, "QuotePID": quotePID,
		"ActiveFid": uint(page.Thread.Fid), "Pending": c.pendingAttachments(r),
		"MobileTitle": "回帖", "MobileLink": "/thread/" + strconv.FormatUint(uint64(tid), 10),
		"ExtraJS": htmltemplate.HTML(postFormExtraJS("/thread/"+strconv.FormatUint(uint64(tid), 10), uint(page.Thread.Fid), false)),
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) ModTop(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		c.writeXiunoMessage(r, -1, "请先登录")
		return
	}
	if err := c.service.SetThreadsTop(r.Context(), parseTidarr(r), user.Uid, r.GetForm("top").Int()); err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	c.writeXiunoMessage(r, 0, "设置完成")
}

func (c *Controller) AdminUsers(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	keyword := r.GetQuery("keyword").String()
	users, err := c.service.AdminUsers(r.Context(), keyword)
	if err != nil {
		c.fail(r, err)
		return
	}
	groups, err := c.service.Groups(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	if err = r.Response.WriteTpl("admin/users.html", gview.Params{
		"Title": "用户管理", "Users": users, "Groups": groups,
		"Keyword": keyword, "User": user,
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) AdminUserGroup(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	if err := c.service.UpdateUserGroup(
		r.Context(), r.GetRouter("uid").Uint(), r.GetForm("gid").Uint(), user.Uid,
	); err != nil {
		c.fail(r, err)
		return
	}
	r.Response.RedirectTo("/admin/users")
}

func (c *Controller) AdminForums(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	forums, err := c.service.AdminForums(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	if err = r.Response.WriteTpl("admin/forums.html", gview.Params{
		"Title": "板块管理", "Forums": forums, "User": user,
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) AdminForumCreate(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	var message string
	if r.Method == http.MethodPost {
		fid, err := c.service.CreateForum(
			r.Context(), r.GetForm("name").String(), r.GetForm("rank").Uint(), r.GetForm("brief").String(),
		)
		if err == nil {
			r.Response.RedirectTo("/admin/forum/" + strconv.FormatUint(uint64(fid), 10) + "/edit")
			return
		}
		message = err.Error()
	}
	if err := r.Response.WriteTpl("admin/forum_create.html", gview.Params{
		"Title": "新建板块", "Error": message, "User": user,
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) AdminForumEdit(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	fid := r.GetRouter("fid").Uint()
	forum, rules, err := c.service.ForumEditor(r.Context(), fid)
	if err != nil {
		c.fail(r, err)
		return
	}
	var message string
	if r.Method == http.MethodPost {
		for i := range rules {
			suffix := strconv.FormatUint(uint64(rules[i].Gid), 10)
			rules[i].Allowread = r.GetForm("allowread_" + suffix).Uint()
			rules[i].Allowthread = r.GetForm("allowthread_" + suffix).Uint()
			rules[i].Allowpost = r.GetForm("allowpost_" + suffix).Uint()
			rules[i].Allowattach = r.GetForm("allowattach_" + suffix).Uint()
			rules[i].Allowdown = r.GetForm("allowdown_" + suffix).Uint()
		}
		err = c.service.UpdateForum(
			r.Context(), fid, r.GetForm("name").String(), r.GetForm("rank").Uint(),
			r.GetForm("brief").String(), r.GetForm("announcement").String(), r.GetForm("moderators").String(),
			r.GetForm("accesson").Uint(), rules,
		)
		if err == nil {
			r.Response.RedirectTo("/admin/forums")
			return
		}
		message = err.Error()
	}
	if err = r.Response.WriteTpl("admin/forum_edit.html", gview.Params{
		"Title": "编辑板块", "Forum": forum, "Rules": rules,
		"Error": message, "User": user,
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) AdminForumDelete(r *ghttp.Request) {
	_, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	if err := c.service.DeleteForum(r.Context(), r.GetRouter("fid").Uint()); err != nil {
		c.fail(r, err)
		return
	}
	r.Response.RedirectTo("/admin/forums")
}

func (c *Controller) AdminGroups(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	groups, err := c.service.AdminGroups(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	if err = r.Response.WriteTpl("admin/groups.html", gview.Params{
		"Title": "用户组管理", "Groups": groups, "User": user,
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) AdminGroupEdit(r *ghttp.Request) {
	user, ok := c.requireAdmin(r)
	if !ok {
		return
	}
	gid := r.GetRouter("gid").Uint()
	groups, err := c.service.AdminGroups(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	var group view.GroupPermission
	for _, item := range groups {
		if item.Gid == gid {
			group = item
			break
		}
	}
	if group.Name == "" {
		c.fail(r, http.ErrMissingFile)
		return
	}
	var message string
	if r.Method == http.MethodPost {
		group.Name = r.GetForm("name").String()
		group.Creditsfrom = r.GetForm("creditsfrom").Int()
		group.Creditsto = r.GetForm("creditsto").Int()
		group.Allowread = r.GetForm("allowread").Int()
		group.Allowthread = r.GetForm("allowthread").Int()
		group.Allowpost = r.GetForm("allowpost").Int()
		group.Allowattach = r.GetForm("allowattach").Int()
		group.Allowdown = r.GetForm("allowdown").Int()
		group.Allowtop = r.GetForm("allowtop").Int()
		group.Allowupdate = r.GetForm("allowupdate").Int()
		group.Allowdelete = r.GetForm("allowdelete").Int()
		group.Allowmove = r.GetForm("allowmove").Int()
		group.Allowbanuser = r.GetForm("allowbanuser").Int()
		group.Allowdeleteuser = r.GetForm("allowdeleteuser").Int()
		group.Allowviewip = r.GetForm("allowviewip").Uint()
		if err = c.service.UpdateGroup(r.Context(), group); err == nil {
			r.Response.RedirectTo("/admin/groups")
			return
		}
		message = err.Error()
	}
	if err = r.Response.WriteTpl("admin/group_edit.html", gview.Params{
		"Title": "编辑用户组", "Group": group, "Error": message, "User": user,
	}); err != nil {
		c.fail(r, err)
	}
}

func (c *Controller) currentUser(r *ghttp.Request) view.User {
	uid := r.Session.MustGet(sessionUID).Uint()
	user, _ := c.service.User(r.Context(), uid)
	return user
}

func (c *Controller) requireAdmin(r *ghttp.Request) (user view.User, ok bool) {
	user = c.currentUser(r)
	if user.Uid == 0 {
		r.Response.RedirectTo("/login")
		return user, false
	}
	if user.Gid != 1 {
		r.Response.Status = http.StatusForbidden
		r.Response.Write("只有管理员可以访问这里")
		return user, false
	}
	return user, true
}

func (c *Controller) fail(r *ghttp.Request, err error) {
	r.Response.Status = http.StatusInternalServerError
	r.Response.Write("暂时无法完成请求：", err.Error())
}

func remoteIPv4(r *ghttp.Request) uint32 {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host).To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
