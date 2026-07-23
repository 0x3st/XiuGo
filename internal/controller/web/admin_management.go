package web

import (
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gview"

	"github.com/0x3st/XiuGo/internal/model/view"
)

const adminThreadQueueSession = "xiuno_go_thread_find_queueid"

var indexedFormPattern = regexp.MustCompile(`^([A-Za-z_]+)\[(\d+)]$`)

func (c *Controller) adminCompatForum(r *ghttp.Request, user view.User, action string, arguments []string) {
	if action == "getname" {
		names, err := c.service.AdminModeratorNames(r.Context(), firstArgument(arguments))
		if err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, names)
		return
	}
	if action == "update" {
		c.adminCompatForumUpdate(r, user, adminArgumentUint(arguments, 0))
		return
	}
	if action == "delete" {
		fid := adminArgumentUint(arguments, 0)
		if fid == 1 {
			c.writeXiunoMessage(r, -1, "Not allowed")
			return
		}
		if err := c.service.DeleteForum(r.Context(), fid); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, "删除成功")
		return
	}
	if r.Method == http.MethodPost {
		fidValues := indexedFormValues(r, "fid")
		nameValues := indexedFormValues(r, "name")
		rankValues := indexedFormValues(r, "rank")
		iconValues := indexedFormValues(r, "icon")
		forums := make(map[uint]view.AdminForum, len(fidValues))
		for fid := range fidValues {
			rank, _ := strconv.ParseUint(rankValues[fid], 10, 64)
			forums[fid] = view.AdminForum{Fid: fid, Name: nameValues[fid], Rank: uint(rank)}
		}
		if err := c.service.SyncAdminForums(r.Context(), forums, iconValues); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, "保存成功")
		return
	}
	forums, err := c.service.AdminForums(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	maxID := uint(0)
	for _, forum := range forums {
		if forum.Fid > maxID {
			maxID = forum.Fid
		}
	}
	c.writeAdminCompatTemplate(r, "admin_compat/forum_list.html", gview.Params{
		"Title": "版块管理", "Active": "forum", "User": user, "Forums": forums, "MaxID": maxID,
	})
}

func (c *Controller) adminCompatForumUpdate(r *ghttp.Request, user view.User, fid uint) {
	if r.Method == http.MethodPost {
		groups, err := c.service.AdminGroups(r.Context())
		if err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		read := indexedFormValues(r, "allowread")
		thread := indexedFormValues(r, "allowthread")
		post := indexedFormValues(r, "allowpost")
		attach := indexedFormValues(r, "allowattach")
		down := indexedFormValues(r, "allowdown")
		rules := make([]view.ForumAccessRule, 0, len(groups))
		for _, group := range groups {
			rules = append(rules, view.ForumAccessRule{
				Gid: group.Gid, Allowread: formFlag(read, group.Gid), Allowthread: formFlag(thread, group.Gid),
				Allowpost: formFlag(post, group.Gid), Allowattach: formFlag(attach, group.Gid), Allowdown: formFlag(down, group.Gid),
			})
		}
		if err = c.service.UpdateForum(
			r.Context(), fid, formString(r, "name"), formUint(r, "rank"), formString(r, "brief"),
			formString(r, "announcement"), formString(r, "modnames"), formUint(r, "accesson"), rules,
		); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, "编辑成功")
		return
	}
	forum, rules, err := c.service.ForumEditor(r.Context(), fid)
	if err != nil {
		c.fail(r, err)
		return
	}
	c.writeAdminCompatTemplate(r, "admin_compat/forum_update.html", gview.Params{
		"Title": "编辑版块", "Active": "forum", "User": user, "Forum": forum, "Rules": rules,
	})
}

func (c *Controller) adminCompatGroup(r *ghttp.Request, user view.User, action string, arguments []string) {
	if action == "update" {
		gid := adminArgumentUint(arguments, 0)
		if r.Method == http.MethodPost {
			group := view.GroupPermission{
				Gid: gid, Name: formString(r, "name"), Creditsfrom: formInt(r, "creditsfrom"), Creditsto: formInt(r, "creditsto"),
				Allowread: formInt(r, "allowread"), Allowthread: formInt(r, "allowthread"), Allowpost: formInt(r, "allowpost"),
				Allowattach: formInt(r, "allowattach"), Allowdown: formInt(r, "allowdown"), Allowtop: formInt(r, "allowtop"),
				Allowupdate: formInt(r, "allowupdate"), Allowdelete: formInt(r, "allowdelete"), Allowmove: formInt(r, "allowmove"),
				Allowbanuser: formInt(r, "allowbanuser"), Allowdeleteuser: formInt(r, "allowdeleteuser"), Allowviewip: formUint(r, "allowviewip"),
			}
			if err := c.service.UpdateGroup(r.Context(), group); err != nil {
				c.writeXiunoMessage(r, -1, err.Error())
				return
			}
			c.writeXiunoMessage(r, 0, "编辑成功")
			return
		}
		group, err := c.service.AdminGroup(r.Context(), gid)
		if err != nil {
			c.fail(r, err)
			return
		}
		c.writeAdminCompatTemplate(r, "admin_compat/group_update.html", gview.Params{
			"Title": "用户组管理", "Active": "user", "User": user, "Group": group,
		})
		return
	}
	if r.Method == http.MethodPost {
		gidValues := indexedFormValues(r, "_gid")
		nameValues := indexedFormValues(r, "name")
		fromValues := indexedFormValues(r, "creditsfrom")
		toValues := indexedFormValues(r, "creditsto")
		groups := make(map[uint]view.GroupPermission, len(gidValues))
		for gid := range gidValues {
			from, _ := strconv.Atoi(fromValues[gid])
			to, _ := strconv.Atoi(toValues[gid])
			groups[gid] = view.GroupPermission{Gid: gid, Name: nameValues[gid], Creditsfrom: from, Creditsto: to}
		}
		if err := c.service.SyncAdminGroups(r.Context(), groups); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, "保存成功")
		return
	}
	groups, err := c.service.AdminGroups(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	maxID := uint(0)
	for _, group := range groups {
		if group.Gid > maxID {
			maxID = group.Gid
		}
	}
	c.writeAdminCompatTemplate(r, "admin_compat/group_list.html", gview.Params{
		"Title": "用户组管理", "Active": "user", "User": user, "Groups": groups, "MaxID": maxID,
	})
}

func (c *Controller) adminCompatUser(r *ghttp.Request, user view.User, action string, arguments []string) {
	switch action {
	case "create":
		c.adminCompatUserEdit(r, user, 0)
		return
	case "update":
		c.adminCompatUserEdit(r, user, adminArgumentUint(arguments, 0))
		return
	case "delete":
		if r.Method != http.MethodPost {
			c.writeXiunoMessage(r, -1, "Method Error.")
			return
		}
		if err := c.service.DeleteAdminUser(r.Context(), formUint(r, "uid")); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, "删除成功")
		return
	}
	searchType, keyword, page := "uid", "", 1
	if len(arguments) > 0 && arguments[0] != "" {
		searchType = arguments[0]
	}
	if len(arguments) > 1 {
		keyword = arguments[1]
	}
	if len(arguments) > 2 {
		page, _ = strconv.Atoi(arguments[2])
	}
	result, err := c.service.AdminUserList(r.Context(), searchType, keyword, page)
	if err != nil {
		c.fail(r, err)
		return
	}
	pages := make([]int, result.Pages)
	for index := range pages {
		pages[index] = index + 1
	}
	c.writeAdminCompatTemplate(r, "admin_compat/user_list.html", gview.Params{
		"Title": "用户管理", "Active": "user", "User": user, "Result": result, "Pages": pages,
	})
}

func (c *Controller) adminCompatUserEdit(r *ghttp.Request, current view.User, uid uint) {
	if r.Method == http.MethodPost {
		var err error
		if uid == 0 {
			err = c.service.CreateAdminUser(
				r.Context(), formString(r, "email"), formString(r, "username"), formString(r, "password"),
				formUint(r, "_gid"), ipv4FromRequest(r),
			)
		} else {
			err = c.service.UpdateAdminUser(
				r.Context(), uid, formString(r, "email"), formString(r, "username"), formString(r, "password"), formUint(r, "_gid"),
			)
		}
		if err != nil {
			code := any(-1)
			if strings.Contains(err.Error(), "邮箱") {
				code = "email"
			} else if strings.Contains(err.Error(), "用户名") || strings.Contains(err.Error(), "用户已经") {
				code = "username"
			}
			c.writeXiunoMessage(r, code, err.Error())
			return
		}
		message := "更新成功"
		if uid == 0 {
			message = "创建成功"
		}
		c.writeXiunoMessage(r, 0, message)
		return
	}
	groups, err := c.service.AdminGroups(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	var edited any
	if uid != 0 {
		user, readErr := c.service.AdminUser(r.Context(), uid)
		if readErr != nil {
			c.fail(r, readErr)
			return
		}
		edited = user
	}
	template := "admin_compat/user_create.html"
	title := "创建用户"
	if uid != 0 {
		template, title = "admin_compat/user_update.html", "编辑用户"
	}
	c.writeAdminCompatTemplate(r, template, gview.Params{
		"Title": title, "Active": "user", "User": current, "Edited": edited, "Groups": groups, "UID": uid,
	})
}

func (c *Controller) adminCompatThread(r *ghttp.Request, user view.User, action string, arguments []string) {
	if action == "scan" {
		queueID := r.Session.MustGet(adminThreadQueueSession).Uint()
		if queueID == 0 {
			c.writeXiunoMessage(r, -1, "主题队列不存在")
			return
		}
		tids, err := c.service.ScanAdminThreads(r.Context(), queueID, view.AdminThreadScan{
			Fid: formUint(r, "fid"), CreateDateStart: parseAdminDate(formString(r, "create_date_start")),
			CreateDateEnd: parseAdminDate(formString(r, "create_date_end")), Uid: formUint(r, "uid"),
			UserIP: ipv4LongValue(formString(r, "userip")), Keyword: formString(r, "keyword"), Page: formInt(r, "page"),
		})
		if err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, tids)
		return
	}
	if action == "operation" {
		queueID := r.Session.MustGet(adminThreadQueueSession).Uint()
		if queueID == 0 {
			c.writeXiunoMessage(r, -1, "主题队列不存在")
			return
		}
		tids, err := c.service.OperateAdminThreadQueue(r.Context(), queueID, user.Uid, firstArgument(arguments))
		if err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, tids)
		return
	}
	if action == "found" {
		queueID := r.Session.MustGet(adminThreadQueueSession).Uint()
		if queueID == 0 {
			c.writeXiunoMessage(r, -1, "主题队列不存在")
			return
		}
		page := int(adminArgumentUint(arguments, 0))
		threads, total, err := c.service.AdminThreadQueue(r.Context(), queueID, page)
		if err != nil {
			c.fail(r, err)
			return
		}
		c.writeAdminCompatTemplate(r, "admin_compat/thread_found.html", gview.Params{
			"Title": "主题搜索结果", "Active": "thread", "User": user, "Threads": threads, "Total": total,
		})
		return
	}
	queueID, err := c.service.CreateAdminThreadQueue(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	r.Session.MustSet(adminThreadQueueSession, queueID)
	forums, err := c.service.AdminForums(r.Context())
	if err != nil {
		c.fail(r, err)
		return
	}
	dashboard, err := c.service.OriginalAdminDashboard(r.Context(), requestIP(r))
	if err != nil {
		c.fail(r, err)
		return
	}
	totalPages := (dashboard.Threads + 99) / 100
	if totalPages < 1 {
		totalPages = 1
	}
	c.writeAdminCompatTemplate(r, "admin_compat/thread_list.html", gview.Params{
		"Title": "主题管理", "Active": "thread", "User": user, "Forums": forums, "TotalPages": totalPages,
	})
}

func (c *Controller) adminCompatOther(r *ghttp.Request, user view.User, action string) {
	if action == "" {
		action = "cache"
	}
	if action != "cache" {
		c.writeXiunoMessage(r, -1, "功能不存在")
		return
	}
	if r.Method == http.MethodPost {
		if err := c.service.ClearOriginalCache(r.Context(), formUint(r, "clear_cache") != 0, formUint(r, "clear_tmp") != 0); err != nil {
			c.writeXiunoMessage(r, -1, err.Error())
			return
		}
		c.writeXiunoMessage(r, 0, "维护完成")
		return
	}
	c.writeAdminCompatTemplate(r, "admin_compat/other_cache.html", gview.Params{
		"Title": "其他", "Active": "other", "User": user,
	})
}

func (c *Controller) adminCompatPlugin(r *ghttp.Request, user view.User, action string, arguments []string) {
	_ = user
	_ = action
	_ = arguments
	c.writeXiunoMessage(r, -1, "XiuGo 不支持原版 PHP 插件")
}

func indexedFormValues(r *ghttp.Request, name string) map[uint]string {
	_ = r.Request.ParseForm()
	values := make(map[uint]string)
	for key, items := range r.Request.PostForm {
		match := indexedFormPattern.FindStringSubmatch(key)
		if len(match) != 3 || match[1] != name || len(items) == 0 {
			continue
		}
		index, err := strconv.ParseUint(match[2], 10, 64)
		if err == nil {
			values[uint(index)] = items[len(items)-1]
		}
	}
	return values
}

func flatFormValues(r *ghttp.Request, name string) []string {
	_ = r.Request.ParseForm()
	if values := r.Request.PostForm[name+"[]"]; len(values) > 0 {
		return values
	}
	return r.Request.PostForm[name]
}

func formString(r *ghttp.Request, name string) string {
	value := r.GetForm(name)
	if value == nil {
		return ""
	}
	return value.String()
}

func formUint(r *ghttp.Request, name string) uint {
	value := r.GetForm(name)
	if value == nil {
		return 0
	}
	return value.Uint()
}

func formInt(r *ghttp.Request, name string) int {
	value := r.GetForm(name)
	if value == nil {
		return 0
	}
	return value.Int()
}

func formFlag(values map[uint]string, key uint) uint {
	if values[key] == "" || values[key] == "0" {
		return 0
	}
	return 1
}

func firstArgument(arguments []string) string {
	if len(arguments) == 0 {
		return ""
	}
	return arguments[0]
}

func parseAdminDate(value string) uint {
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return 0
	}
	return uint(parsed.Unix())
}

func ipv4FromRequest(r *ghttp.Request) uint {
	return ipv4LongValue(requestIP(r))
}

func ipv4LongValue(value string) uint {
	parts := strings.Split(value, ".")
	if len(parts) != 4 {
		return 0
	}
	var result uint
	for _, part := range parts {
		number, err := strconv.ParseUint(part, 10, 8)
		if err != nil {
			return 0
		}
		result = result<<8 | uint(number)
	}
	return result
}

func sortedFormKeys(values map[uint]string) []uint {
	keys := make([]uint, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
