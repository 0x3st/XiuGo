package bbs

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/0x3st/XiuGo/internal/dao"
	"github.com/0x3st/XiuGo/internal/model/do"
	"github.com/0x3st/XiuGo/internal/model/entity"
	"github.com/0x3st/XiuGo/internal/model/view"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Home(ctx context.Context, viewer view.User) (forums []view.ForumSummary, threads []view.ThreadSummary, stats view.Stats, err error) {
	if err = dao.BbsForum.Ctx(ctx).
		Fields("fid,name,brief,threads,todayposts,todaythreads,announcement,icon").
		OrderDesc(dao.BbsForum.Columns().Rank).
		Scan(&forums); err != nil {
		return nil, nil, stats, gerror.Wrap(err, "读取板块失败")
	}
	visibleForums := make([]view.ForumSummary, 0, len(forums))
	for _, forum := range forums {
		allowed, permissionErr := s.forumPermission(ctx, forum.Fid, viewer.Gid, "read")
		if permissionErr != nil {
			return nil, nil, stats, permissionErr
		}
		if allowed {
			visibleForums = append(visibleForums, forum)
		}
	}
	forums = visibleForums
	for i := range forums {
		forums[i].IconURL = forumIconURL(forums[i].Fid, forums[i].Icon)
	}
	if threads, err = s.listThreads(ctx, 0, 20, "lastpid"); err != nil {
		return nil, nil, stats, err
	}
	if stats.Threads, err = dao.BbsThread.Ctx(ctx).Count(); err != nil {
		return nil, nil, stats, gerror.Wrap(err, "统计主题失败")
	}
	if stats.Posts, err = dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Isfirst: 0}).Count(); err != nil {
		return nil, nil, stats, gerror.Wrap(err, "统计帖子失败")
	}
	if stats.Users, err = dao.BbsUser.Ctx(ctx).Count(); err != nil {
		return nil, nil, stats, gerror.Wrap(err, "统计用户失败")
	}
	if stats.Onlines, err = s.originalRuntimeOnlines(ctx); err != nil {
		return nil, nil, stats, err
	}
	return forums, threads, stats, nil
}

func (s *Service) Forum(ctx context.Context, fid uint, viewer view.User) (forum view.ForumSummary, threads []view.ThreadSummary, err error) {
	if err = dao.BbsForum.Ctx(ctx).
		Where(do.BbsForum{Fid: fid}).
		Scan(&forum); err != nil {
		return forum, nil, gerror.Wrap(err, "读取板块失败")
	}
	if forum.Fid == 0 {
		return forum, nil, gerror.New("板块不存在")
	}
	allowed, err := s.forumPermission(ctx, fid, viewer.Gid, "read")
	if err != nil {
		return forum, nil, err
	}
	if !allowed {
		return forum, nil, gerror.New("当前用户组没有查看此板块的权限")
	}
	forum.IconURL = forumIconURL(forum.Fid, forum.Icon)
	if threads, err = s.listThreads(ctx, fid, 50, "lastpid"); err != nil {
		return forum, nil, err
	}
	return forum, threads, nil
}

// ForumOrdered is like Forum but allows lastpid/tid sort matching forum.php.
func (s *Service) ForumOrdered(ctx context.Context, fid uint, viewer view.User, orderby string) (forum view.ForumSummary, threads []view.ThreadSummary, err error) {
	forum, _, err = s.Forum(ctx, fid, viewer)
	if err != nil {
		return forum, nil, err
	}
	if orderby != "tid" {
		orderby = "lastpid"
	}
	threads, err = s.listThreads(ctx, fid, 50, orderby)
	return forum, threads, err
}


func (s *Service) listThreads(ctx context.Context, fid uint, limit int, orderby string) (threads []view.ThreadSummary, err error) {
	if orderby != "tid" {
		orderby = "lastpid"
	}
	orderColumn := "bbs_thread.lastpid"
	if orderby == "tid" {
		orderColumn = "bbs_thread.tid"
	}
	model := dao.BbsThread.Ctx(ctx).
		LeftJoin("bbs_user u", "u.uid=bbs_thread.uid").
		LeftJoin("bbs_user lu", "lu.uid=bbs_thread.lastuid").
		LeftJoin("bbs_forum f", "f.fid=bbs_thread.fid").
		Fields("bbs_thread.tid,bbs_thread.uid,bbs_thread.fid,bbs_thread.subject,bbs_thread.create_date,bbs_thread.last_date,bbs_thread.lastuid,bbs_thread.views,bbs_thread.posts,bbs_thread.files,bbs_thread.closed,bbs_thread.top,u.username,u.avatar,lu.username AS last_username,f.name AS forum_name").
		OrderDesc(orderColumn).
		Limit(limit)
	if fid > 0 {
		model = model.Where("bbs_thread.fid", fid)
	}
	if err = model.Scan(&threads); err != nil {
		return nil, gerror.Wrap(err, "读取主题列表失败")
	}
	// Avatar comes from join as "avatar" on first user - map via format helper using Uid only default for now.
	for i := range threads {
		formatThreadSummary(&threads[i])
	}
	// Match thread_find_by_fid / index: pinned threads lead page 1 only when ordered by default lastpid.
	if orderby == "lastpid" {
		pinned, pinErr := s.listPinnedThreads(ctx, fid)
		if pinErr != nil {
			return nil, pinErr
		}
		return mergePinnedThreads(pinned, threads, limit), nil
	}
	return threads, nil
}

// listPinnedThreads mirrors thread_top_find(). fid=0 returns site-wide top=3 only;
// fid>0 returns site-wide top=3 plus forum top=1 for that board.
func (s *Service) listPinnedThreads(ctx context.Context, fid uint) ([]view.ThreadSummary, error) {
	var topRows []entity.BbsThreadTop
	model := dao.BbsThreadTop.Ctx(ctx).OrderDesc(dao.BbsThreadTop.Columns().Tid).Limit(200)
	if fid == 0 {
		model = model.Where(do.BbsThreadTop{Top: 3})
	} else {
		// Site-wide pins (top=3) plus this forum's board pins (top=1).
		model = model.Where("top=? OR (fid=? AND top=?)", 3, fid, 1)
	}
	if err := model.Scan(&topRows); err != nil {
		return nil, gerror.Wrap(err, "读取置顶索引失败")
	}
	if len(topRows) == 0 {
		return nil, nil
	}
	// Preserve PHP order: site-wide top=3 first (by tid desc), then forum top=1.
	siteIDs := make([]uint, 0)
	forumIDs := make([]uint, 0)
	for _, row := range topRows {
		if row.Top == 3 {
			siteIDs = append(siteIDs, row.Tid)
		} else if fid > 0 && row.Top == 1 && uint(row.Fid) == fid {
			forumIDs = append(forumIDs, row.Tid)
		}
	}
	orderedIDs := append(siteIDs, forumIDs...)
	if len(orderedIDs) == 0 {
		return nil, nil
	}
	var threads []view.ThreadSummary
	if err := dao.BbsThread.Ctx(ctx).
		LeftJoin("bbs_user u", "u.uid=bbs_thread.uid").
		LeftJoin("bbs_user lu", "lu.uid=bbs_thread.lastuid").
		LeftJoin("bbs_forum f", "f.fid=bbs_thread.fid").
		Fields("bbs_thread.tid,bbs_thread.uid,bbs_thread.fid,bbs_thread.subject,bbs_thread.create_date,bbs_thread.last_date,bbs_thread.lastuid,bbs_thread.views,bbs_thread.posts,bbs_thread.files,bbs_thread.closed,bbs_thread.top,u.username,u.avatar,lu.username AS last_username,f.name AS forum_name").
		WhereIn("bbs_thread.tid", orderedIDs).
		Scan(&threads); err != nil {
		return nil, gerror.Wrap(err, "读取置顶主题失败")
	}
	byTID := make(map[uint]view.ThreadSummary, len(threads))
	for _, thread := range threads {
		formatThreadSummary(&thread)
		byTID[thread.Tid] = thread
	}
	result := make([]view.ThreadSummary, 0, len(orderedIDs))
	for _, tid := range orderedIDs {
		if thread, ok := byTID[tid]; ok {
			result = append(result, thread)
		}
	}
	return result, nil
}

func mergePinnedThreads(pinned, regular []view.ThreadSummary, limit int) []view.ThreadSummary {
	seen := make(map[uint]bool, len(pinned))
	result := make([]view.ThreadSummary, 0, limit)
	for _, thread := range pinned {
		if seen[thread.Tid] {
			continue
		}
		seen[thread.Tid] = true
		result = append(result, thread)
		if limit > 0 && len(result) >= limit {
			return result
		}
	}
	for _, thread := range regular {
		if seen[thread.Tid] {
			continue
		}
		seen[thread.Tid] = true
		result = append(result, thread)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

func threadTopClass(top int) string {
	switch top {
	case 1:
		return "top_1"
	case 3:
		return "top_3"
	default:
		return ""
	}
}

func (s *Service) Thread(ctx context.Context, tid uint, viewer view.User) (page view.ThreadPage, err error) {
	if err = dao.BbsThread.Ctx(ctx).
		LeftJoin("bbs_user u", "u.uid=bbs_thread.uid").
		LeftJoin("bbs_forum f", "f.fid=bbs_thread.fid").
		Fields("bbs_thread.tid,bbs_thread.uid,bbs_thread.fid,bbs_thread.firstpid,bbs_thread.subject,bbs_thread.create_date,bbs_thread.views,bbs_thread.posts,bbs_thread.closed,bbs_thread.top,u.username,f.name AS forum_name").
		Where("bbs_thread.tid", tid).
		Scan(&page.Thread); err != nil {
		return page, gerror.Wrap(err, "读取主题失败")
	}
	if page.Thread.Tid == 0 {
		return page, gerror.New("主题不存在")
	}
	allowed, err := s.forumPermission(ctx, uint(page.Thread.Fid), viewer.Gid, "read")
	if err != nil {
		return page, err
	}
	if !allowed {
		return page, gerror.New("当前用户组没有查看此主题的权限")
	}
	page.Thread.CreateTime = humanDate(page.Thread.CreateDate)
	page.Thread.TopClass = threadTopClass(page.Thread.Top)
	if err = dao.BbsUser.Ctx(ctx).
		Fields("uid,username,avatar,threads,posts").
		Where(do.BbsUser{Uid: page.Thread.Uid}).
		Scan(&page.Author); err != nil {
		return page, gerror.Wrap(err, "读取主题作者失败")
	}
	page.Author.AvatarURL = avatarURL(page.Author.Uid, page.Author.Avatar)
	var posts []view.PostView
	if err = dao.BbsPost.Ctx(ctx).
		LeftJoin("bbs_user u", "u.uid=bbs_post.uid").
		Fields("bbs_post.pid,bbs_post.tid,bbs_post.uid,bbs_post.isfirst,bbs_post.create_date,bbs_post.doctype,bbs_post.quotepid,bbs_post.message,bbs_post.message_fmt,u.username,u.avatar").
		Where("bbs_post.tid", tid).
		OrderAsc("bbs_post.pid").
		Scan(&posts); err != nil {
		return page, gerror.Wrap(err, "读取回复失败")
	}
	postIDs := make([]uint, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.Pid)
	}
	attachments, err := s.AttachmentsForPosts(ctx, postIDs)
	if err != nil {
		return page, err
	}
	var (
		canPost      bool
		canUpdateMod bool
		canDeleteMod bool
		canTopMod    bool
	)
	// allowpost is group/forum permission and is checked even for guests (quote icons).
	if canPost, err = s.forumPermission(ctx, uint(page.Thread.Fid), viewer.Gid, "post"); err != nil {
		return page, err
	}
	if viewer.Uid != 0 {
		viewerRecord := entity.BbsUser{Uid: viewer.Uid, Gid: viewer.Gid}
		if canUpdateMod, err = s.canModerate(ctx, viewerRecord, uint(page.Thread.Fid), "update"); err != nil {
			return page, err
		}
		if canDeleteMod, err = s.canModerate(ctx, viewerRecord, uint(page.Thread.Fid), "delete"); err != nil {
			return page, err
		}
		if canTopMod, err = s.canModerate(ctx, viewerRecord, uint(page.Thread.Fid), "top"); err != nil {
			return page, err
		}
		page.CanTop = canTopMod
	}
	page.CanReply = viewer.Uid != 0 && canPost && canReplyToThread(page.Thread.Closed, viewer.Gid)
	if viewer.Uid != 0 && !page.CanReply {
		if page.Thread.Closed != 0 && !canReplyToThread(page.Thread.Closed, viewer.Gid) {
			page.ReplyNotice = "主题已经关闭，不能回复"
		} else {
			page.ReplyNotice = "当前用户组没有回复权限"
		}
	}
	for i := range posts {
		posts[i].Files = attachments[posts[i].Pid]
		posts[i].Floor = uint(i + 1)
		posts[i].CreateTime = humanDate(posts[i].CreateDate)
		posts[i].AvatarURL = avatarURL(posts[i].Uid, posts[i].Avatar)
		if posts[i].MessageFmt == "" {
			posts[i].MessageFmt = formatPostMessage(posts[i].Message, posts[i].Doctype, 0)
		}
		// Original post_list.inc.htm shows quote when $allowpost (group/forum), even for guests.
		posts[i].CanQuote = canPost && posts[i].Isfirst == 0
		canUpdate := canPost && (viewer.Uid == posts[i].Uid || canUpdateMod)
		canDelete := canPost && (viewer.Uid == posts[i].Uid || canDeleteMod)
		if page.Thread.Closed != 0 && !canUpdateMod {
			canUpdate = false
		}
		if page.Thread.Closed != 0 && !canDeleteMod {
			canDelete = false
		}
		posts[i].CanEdit = canUpdate
		posts[i].CanDelete = canDelete
		if posts[i].Pid == page.Thread.Firstpid {
			page.First = posts[i]
		} else {
			page.Replies = append(page.Replies, posts[i])
		}
	}
	if page.First.Pid == 0 {
		return page, gerror.New("主题首帖数据不完整")
	}
	page.NextFloor = page.Thread.Posts + 2
	_, _ = dao.BbsThread.Ctx(ctx).
		Where(do.BbsThread{Tid: tid}).
		Data(do.BbsThread{Views: gdb.Raw("views + 1")}).
		Update()
	return page, nil
}

func (s *Service) Authenticate(ctx context.Context, account, password string, userIP uint32) (user view.User, err error) {
	var record entity.BbsUser
	account = strings.TrimSpace(account)
	if account == "" || password == "" {
		return user, gerror.New("请输入账号和密码")
	}
	if _, err = normalizeClientPasswordHash(password); err != nil {
		return user, err
	}
	if err = dao.BbsUser.Ctx(ctx).
		Where(dao.BbsUser.Columns().Email, account).
		WhereOr(dao.BbsUser.Columns().Username, account).
		Scan(&record); err != nil {
		return user, gerror.Wrap(err, "登录查询失败")
	}
	if record.Uid == 0 {
		return user, gerror.New("账号或密码不正确")
	}
	expectedPassword, passwordErr := xiunoPassword(password, record.Salt)
	if passwordErr != nil {
		return user, passwordErr
	}
	if record.Password != expectedPassword {
		return user, gerror.New("账号或密码不正确")
	}
	if _, err = dao.BbsUser.Ctx(ctx).
		Where(do.BbsUser{Uid: record.Uid}).
		Data(do.BbsUser{
			LoginIp: userIP, LoginDate: uint(time.Now().Unix()), Logins: gdb.Raw("logins + 1"),
		}).Update(); err != nil {
		return user, gerror.Wrap(err, "更新登录信息失败")
	}
	return view.User{
		Uid: record.Uid, Username: record.Username, Gid: record.Gid,
		Avatar: record.Avatar, AvatarURL: avatarURL(record.Uid, record.Avatar),
	}, nil
}

func (s *Service) User(ctx context.Context, uid uint) (user view.User, err error) {
	if uid == 0 {
		return user, nil
	}
	var record entity.BbsUser
	if err = dao.BbsUser.Ctx(ctx).
		Fields(dao.BbsUser.Columns().Uid, dao.BbsUser.Columns().Username, dao.BbsUser.Columns().Gid, dao.BbsUser.Columns().Avatar).
		Where(do.BbsUser{Uid: uid}).
		Scan(&record); err != nil {
		return user, gerror.Wrap(err, "读取用户失败")
	}
	return view.User{
		Uid: record.Uid, Username: record.Username, Gid: record.Gid,
		Avatar: record.Avatar, AvatarURL: avatarURL(record.Uid, record.Avatar),
	}, nil
}

func (s *Service) CreateThread(
	ctx context.Context, fid, uid uint, subject, message string, doctype int, userIP uint32,
	pending []view.PendingAttachment,
) (tid uint, err error) {
	var (
		forum entity.BbsForum
		user  entity.BbsUser
	)
	if uid == 0 {
		return 0, gerror.New("请先登录")
	}
	if subject == "" || message == "" {
		return 0, gerror.New("标题和正文不能为空")
	}
	subject = phpHTMLSpecialChars(subject)
	if len([]rune(subject)) > 128 {
		return 0, gerror.New("标题不能超过 128 个字")
	}
	if len([]rune(message)) > 2_028_000 {
		return 0, gerror.New("正文内容过长")
	}
	if !validDoctype(doctype) {
		return 0, gerror.New("不支持的内容格式")
	}
	if err = dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: fid}).Scan(&forum); err != nil {
		return 0, gerror.Wrap(err, "读取板块失败")
	}
	if forum.Fid == 0 {
		return 0, gerror.New("板块不存在")
	}
	if err = s.requireForumPermission(ctx, fid, uid, "thread"); err != nil {
		return 0, err
	}
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return 0, gerror.Wrap(err, "读取发帖用户失败")
	}
	messageFmt := formatPostMessage(message, doctype, user.Gid)
	now := uint(time.Now().Unix())
	threadID, err := dao.BbsThread.Ctx(ctx).Data(do.BbsThread{
		Fid: fid, Uid: uid, Userip: userIP, Subject: subject,
		CreateDate: now, LastDate: now, Lastuid: uid,
	}).InsertAndGetId()
	if err != nil {
		return 0, gerror.Wrap(err, "创建主题失败")
	}
	postID, err := dao.BbsPost.Ctx(ctx).Data(do.BbsPost{
		Tid: uint(threadID), Uid: uid, Isfirst: 1, CreateDate: now,
		Userip: userIP, Doctype: doctype, Message: message, MessageFmt: messageFmt,
	}).InsertAndGetId()
	if err != nil {
		return 0, gerror.Wrap(err, "保存主题正文失败")
	}
	if _, err = dao.BbsThread.Ctx(ctx).
		Where(do.BbsThread{Tid: threadID}).
		Data(do.BbsThread{Firstpid: postID, Lastpid: postID}).
		Update(); err != nil {
		return 0, gerror.Wrap(err, "更新主题索引失败")
	}
	if _, err = dao.BbsMythread.Ctx(ctx).Data(do.BbsMythread{
		Uid: uid, Tid: uint(threadID),
	}).Insert(); err != nil {
		return 0, gerror.Wrap(err, "写入用户主题索引失败")
	}
	if _, err = dao.BbsMypost.Ctx(ctx).Data(do.BbsMypost{
		Uid: uid, Tid: uint(threadID), Pid: uint(postID),
	}).Insert(); err != nil {
		return 0, gerror.Wrap(err, "写入用户帖子索引失败")
	}
	if err = s.AssociatePendingAttachments(ctx, uint(postID), uid, pending); err != nil {
		return 0, err
	}
	if _, err = dao.BbsForum.Ctx(ctx).
		Where(do.BbsForum{Fid: fid}).
		Data(do.BbsForum{Threads: gdb.Raw("threads + 1"), Todaythreads: gdb.Raw("todaythreads + 1")}).
		Update(); err != nil {
		return 0, gerror.Wrap(err, "更新板块计数失败")
	}
	if _, err = dao.BbsUser.Ctx(ctx).
		Where(do.BbsUser{Uid: uid}).
		Data(do.BbsUser{Threads: gdb.Raw("threads + 1")}).
		Update(); err != nil {
		return 0, gerror.Wrap(err, "更新用户计数失败")
	}
	if err = s.SyncPHPRuntime(ctx, map[string]int{"todaythreads": 1}); err != nil {
		return 0, err
	}
	return uint(threadID), nil
}

func (s *Service) Reply(
	ctx context.Context, tid, uid uint, message string, doctype int, quotePID uint, userIP uint32,
	pending []view.PendingAttachment,
) (pid uint, err error) {
	var (
		thread entity.BbsThread
		user   entity.BbsUser
	)
	if uid == 0 {
		return 0, gerror.New("请先登录")
	}
	if err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Scan(&thread); err != nil {
		return 0, gerror.Wrap(err, "读取主题失败")
	}
	if thread.Tid == 0 {
		return 0, gerror.New("主题不存在")
	}
	if err = s.requireForumPermission(ctx, uint(thread.Fid), uid, "post"); err != nil {
		return 0, err
	}
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return 0, gerror.Wrap(err, "读取回复用户失败")
	}
	if !canReplyToThread(thread.Closed, user.Gid) {
		return 0, gerror.New("主题已经关闭，不能回复")
	}
	if message == "" {
		return 0, gerror.New("回复内容不能为空")
	}
	if len([]rune(message)) > 2_028_000 {
		return 0, gerror.New("回复内容过长")
	}
	if !validDoctype(doctype) {
		return 0, gerror.New("不支持的内容格式")
	}
	validatedQuotePID, quoteHTML, err := s.replyQuote(ctx, tid, quotePID)
	if err != nil {
		return 0, err
	}
	messageFmt := quoteHTML + formatPostMessage(message, doctype, user.Gid)

	now := uint(time.Now().Unix())
	postID, err := dao.BbsPost.Ctx(ctx).Data(do.BbsPost{
		Tid: tid, Uid: uid, Isfirst: 0, CreateDate: now,
		Userip: userIP, Doctype: doctype, Quotepid: int(validatedQuotePID), Message: message, MessageFmt: messageFmt,
	}).InsertAndGetId()
	if err != nil {
		return 0, gerror.Wrap(err, "保存回复失败")
	}
	if _, err = dao.BbsMypost.Ctx(ctx).Data(do.BbsMypost{
		Uid: uid, Tid: tid, Pid: uint(postID),
	}).Insert(); err != nil {
		return 0, gerror.Wrap(err, "写入用户帖子索引失败")
	}
	if err = s.AssociatePendingAttachments(ctx, uint(postID), uid, pending); err != nil {
		return 0, err
	}
	if _, err = dao.BbsThread.Ctx(ctx).
		Where(do.BbsThread{Tid: tid}).
		Data(do.BbsThread{
			Posts: gdb.Raw("posts + 1"), LastDate: now, Lastuid: uid, Lastpid: uint(postID),
		}).Update(); err != nil {
		return 0, gerror.Wrap(err, "更新主题回复索引失败")
	}
	if _, err = dao.BbsForum.Ctx(ctx).
		Where(do.BbsForum{Fid: thread.Fid}).
		Data(do.BbsForum{Todayposts: gdb.Raw("todayposts + 1")}).Update(); err != nil {
		return 0, gerror.Wrap(err, "更新板块计数失败")
	}
	if _, err = dao.BbsUser.Ctx(ctx).
		Where(do.BbsUser{Uid: uid}).
		Data(do.BbsUser{Posts: gdb.Raw("posts + 1")}).Update(); err != nil {
		return 0, gerror.Wrap(err, "更新用户计数失败")
	}
	if err = s.SyncPHPRuntime(ctx, map[string]int{"todayposts": 1}); err != nil {
		return 0, err
	}
	return uint(postID), nil
}

func (s *Service) replyQuote(ctx context.Context, tid, quotePID uint) (uint, string, error) {
	if quotePID == 0 {
		return 0, "", nil
	}
	var (
		post entity.BbsPost
		user entity.BbsUser
	)
	if err := dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: quotePID}).Scan(&post); err != nil {
		return 0, "", gerror.Wrap(err, "读取引用帖子失败")
	}
	if post.Pid == 0 || post.Tid != tid {
		return 0, "", nil
	}
	if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: post.Uid}).Scan(&user); err != nil {
		return 0, "", gerror.Wrap(err, "读取引用用户失败")
	}
	quote := fmt.Sprintf(`<blockquote class="blockquote">
		<a href="/user/%d" class="text-small text-muted user">
			<img class="avatar-1" src="%s">
			%s
		</a>
		%s
		</blockquote>`, user.Uid, avatarURL(user.Uid, user.Avatar), user.Username, quoteBrief(post.Message, 100))
	return quotePID, quote, nil
}

func (s *Service) PostForEdit(ctx context.Context, pid, uid uint) (result view.PostEdit, err error) {
	var (
		post   entity.BbsPost
		thread entity.BbsThread
	)
	if err = dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: pid}).Scan(&post); err != nil {
		return result, gerror.Wrap(err, "读取帖子失败")
	}
	if post.Pid == 0 {
		return result, gerror.New("帖子不存在")
	}
	if err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: post.Tid}).Scan(&thread); err != nil {
		return result, gerror.Wrap(err, "读取主题失败")
	}
	moderator, err := s.canPostAction(ctx, uid, post.Uid, uint(thread.Fid), "update")
	if err != nil {
		return result, err
	}
	if thread.Closed != 0 && !moderator {
		return result, gerror.New("主题已经关闭，不能编辑")
	}
	message := post.Message
	if message == "" {
		message = post.MessageFmt
	}
	files, err := s.AttachmentsForPost(ctx, post.Pid)
	if err != nil {
		return result, err
	}
	return view.PostEdit{
		Pid: post.Pid, Tid: post.Tid, Fid: thread.Fid, Subject: thread.Subject,
		Message: phpHTMLSpecialChars(message), Isfirst: post.Isfirst,
		Doctype: post.Doctype, Quotepid: post.Quotepid, Files: files,
	}, nil
}

func (s *Service) UpdatePost(
	ctx context.Context, pid, uid uint, subject, message string, doctype int,
	pending []view.PendingAttachment,
) (tid uint, err error) {
	var (
		post   entity.BbsPost
		thread entity.BbsThread
		user   entity.BbsUser
	)
	if message == "" {
		return 0, gerror.New("帖子内容不能为空")
	}
	if len([]rune(message)) > 2_048_000 {
		return 0, gerror.New("帖子内容过长")
	}
	if !validDoctype(doctype) {
		return 0, gerror.New("不支持的内容格式")
	}
	if err = dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: pid}).Scan(&post); err != nil {
		return 0, gerror.Wrap(err, "读取帖子失败")
	}
	if post.Pid == 0 {
		return 0, gerror.New("帖子不存在")
	}
	if err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: post.Tid}).Scan(&thread); err != nil {
		return 0, gerror.Wrap(err, "读取主题失败")
	}
	moderator, err := s.canPostAction(ctx, uid, post.Uid, uint(thread.Fid), "update")
	if err != nil {
		return 0, err
	}
	if thread.Closed != 0 && !moderator {
		return 0, gerror.New("主题已经关闭，不能编辑")
	}
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return 0, gerror.Wrap(err, "读取编辑用户失败")
	}
	if post.Isfirst != 0 {
		if subject == "" {
			return 0, gerror.New("主题标题不能为空")
		}
		subject = phpHTMLSpecialChars(subject)
		if len([]rune(subject)) > 80 {
			return 0, gerror.New("主题标题不能超过 80 个字")
		}
		if _, err = dao.BbsThread.Ctx(ctx).
			Where(do.BbsThread{Tid: post.Tid}).
			Data(do.BbsThread{Subject: subject}).Update(); err != nil {
			return 0, gerror.Wrap(err, "更新主题标题失败")
		}
	}
	messageFmt := formatPostMessage(message, doctype, user.Gid)
	if _, err = dao.BbsPost.Ctx(ctx).
		Where(do.BbsPost{Pid: pid}).
		Data(do.BbsPost{Doctype: doctype, Message: message, MessageFmt: messageFmt}).Update(); err != nil {
		return 0, gerror.Wrap(err, "更新帖子失败")
	}
	if err = s.AssociatePendingAttachments(ctx, pid, uid, pending); err != nil {
		return 0, err
	}
	return post.Tid, nil
}

func (s *Service) DeletePost(ctx context.Context, pid, uid uint) (tid uint, err error) {
	var (
		post   entity.BbsPost
		thread entity.BbsThread
	)
	if err = dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: pid}).Scan(&post); err != nil {
		return 0, gerror.Wrap(err, "读取帖子失败")
	}
	if post.Pid == 0 {
		return 0, gerror.New("帖子不存在")
	}
	if err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: post.Tid}).Scan(&thread); err != nil {
		return 0, gerror.Wrap(err, "读取主题失败")
	}
	moderator, err := s.canPostAction(ctx, uid, post.Uid, uint(thread.Fid), "delete")
	if err != nil {
		return 0, err
	}
	if thread.Closed != 0 && !moderator {
		return 0, gerror.New("主题已经关闭，不能删除")
	}
	if post.Isfirst != 0 {
		if err = s.DeleteThread(ctx, post.Tid, uid); err != nil {
			return 0, err
		}
		return 0, nil
	}
	if err = s.deleteReplyOriginal(ctx, post); err != nil {
		return 0, err
	}
	if err = s.SyncPHPRuntime(ctx, nil); err != nil {
		return 0, err
	}
	return post.Tid, nil
}

func (s *Service) Profile(ctx context.Context, uid uint) (profile view.UserProfile, err error) {
	if uid == 0 {
		return profile, gerror.New("用户不存在")
	}
	var record entity.BbsUser
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&record); err != nil {
		return profile, gerror.Wrap(err, "读取用户资料失败")
	}
	if record.Uid == 0 {
		return profile, gerror.New("用户不存在")
	}
	var group entity.BbsGroup
	_ = dao.BbsGroup.Ctx(ctx).Where(do.BbsGroup{Gid: record.Gid}).Scan(&group)
	profile = view.UserProfile{
		Uid: record.Uid, Gid: record.Gid, Username: record.Username, Email: record.Email,
		GroupName: group.Name, Avatar: record.Avatar, AvatarURL: avatarURL(record.Uid, record.Avatar),
		Threads: record.Threads, Posts: record.Posts,
		CreateDate: record.CreateDate, LoginDate: record.LoginDate,
		CreateTime: formatUnixDate(record.CreateDate), LoginTime: formatUnixDate(record.LoginDate),
	}
	return profile, nil
}


// UpdateAvatar mirrors route/my.php action=avatar POST: base64 image to upload/avatar/<dir>/<uid>.png.
func (s *Service) UpdateAvatar(ctx context.Context, uid uint, encodedData string) (url string, err error) {
	if uid == 0 {
		return "", gerror.New("请先登录")
	}
	data, err := decodeAttachmentData(encodedData)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", gerror.New("数据为空")
	}
	if len(data) > 2_048_000 {
		return "", gerror.New("文件尺寸太大，不能超过 2M")
	}
	dir := fmt.Sprintf("%09d", uid)[:3]
	root := filepath.Join(s.uploadRoot(ctx), "avatar", dir)
	if err = os.MkdirAll(root, 0o777); err != nil {
		return "", gerror.Wrap(err, "创建头像目录失败")
	}
	path := filepath.Join(root, fmt.Sprintf("%d.png", uid))
	if err = os.WriteFile(path, data, 0o644); err != nil {
		return "", gerror.Wrap(err, "写入头像失败")
	}
	now := uint(time.Now().Unix())
	if _, err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Data(do.BbsUser{Avatar: now}).Update(); err != nil {
		return "", gerror.Wrap(err, "更新头像时间戳失败")
	}
	return fmt.Sprintf("/upload/avatar/%s/%d.png", dir, uid), nil
}


// AnnotateThreadModFlags sets AllowTop per thread for original list checkboxes.
func (s *Service) AnnotateThreadModFlags(ctx context.Context, threads []view.ThreadSummary, viewer view.User) (haveAllowTop bool, err error) {
	if viewer.Uid == 0 {
		return false, nil
	}
	// Admins/super-mods always see list mod UI (matches forum_access_mod gid 1/2).
	if viewer.Gid == 1 || viewer.Gid == 2 {
		for i := range threads {
			threads[i].AllowTop = true
		}
		return len(threads) > 0, nil
	}
	user := entity.BbsUser{Uid: viewer.Uid, Gid: viewer.Gid}
	for i := range threads {
		allowed, modErr := s.canModerate(ctx, user, uint(threads[i].Fid), "top")
		if modErr != nil {
			return false, modErr
		}
		threads[i].AllowTop = allowed
		if allowed {
			haveAllowTop = true
		}
	}
	return haveAllowTop, nil
}

// RenderPostListHTML returns HTML for one or more posts, matching post_list.inc.htm.
func (s *Service) PostViewForPID(ctx context.Context, pid uint, viewer view.User) (post view.PostView, err error) {
	var record entity.BbsPost
	if err = dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: pid}).Scan(&record); err != nil {
		return post, gerror.Wrap(err, "读取帖子失败")
	}
	if record.Pid == 0 {
		return post, gerror.New("帖子不存在")
	}
	var user entity.BbsUser
	_ = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: record.Uid}).Scan(&user)
	var thread entity.BbsThread
	_ = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: record.Tid}).Scan(&thread)
	files, _ := s.AttachmentsForPost(ctx, record.Pid)
	canPost, _ := s.forumPermission(ctx, uint(thread.Fid), viewer.Gid, "post")
	canUpdateMod, _ := s.canModerate(ctx, entity.BbsUser{Uid: viewer.Uid, Gid: viewer.Gid}, uint(thread.Fid), "update")
	canDeleteMod, _ := s.canModerate(ctx, entity.BbsUser{Uid: viewer.Uid, Gid: viewer.Gid}, uint(thread.Fid), "delete")
	// Xiuno: thread.posts counts replies only. New reply floor = posts+1 after increment
	// (posts already includes this reply). First post floor = 1.
	floor := uint(thread.Posts) + 1
	if record.Isfirst != 0 {
		floor = 1
	}
	if floor < 2 && record.Isfirst == 0 {
		floor = 2
	}
	messageFmt := record.MessageFmt
	if messageFmt == "" {
		messageFmt = formatPostMessage(record.Message, record.Doctype, user.Gid)
	}
	post = view.PostView{
		Pid: record.Pid, Tid: record.Tid, Uid: record.Uid, Isfirst: record.Isfirst,
		Doctype: record.Doctype, Quotepid: record.Quotepid, Floor: floor,
		Username: user.Username, Avatar: user.Avatar, AvatarURL: avatarURL(user.Uid, user.Avatar),
		CreateDate: record.CreateDate, CreateTime: humanDate(record.CreateDate),
		Message: record.Message, MessageFmt: messageFmt, Files: files,
		CanQuote: canPost && record.Isfirst == 0,
		CanEdit: viewer.Uid != 0 && canPost && (viewer.Uid == record.Uid || canUpdateMod),
		CanDelete: viewer.Uid != 0 && canPost && (viewer.Uid == record.Uid || canDeleteMod),
	}
	if thread.Closed != 0 && !canUpdateMod {
		post.CanEdit = false
	}
	if thread.Closed != 0 && !canDeleteMod {
		post.CanDelete = false
	}
	return post, nil
}

func (s *Service) ForumsAllowThread(ctx context.Context, viewer view.User) ([]view.ForumSummary, error) {
	forums, _, _, err := s.Home(ctx, viewer)
	if err != nil {
		return nil, err
	}
	result := make([]view.ForumSummary, 0, len(forums))
	for _, forum := range forums {
		allowed, permErr := s.forumPermission(ctx, forum.Fid, viewer.Gid, "thread")
		if permErr != nil {
			return nil, permErr
		}
		if allowed {
			result = append(result, forum)
		}
	}
	return result, nil
}

func (s *Service) UserThreads(ctx context.Context, uid uint) (threads []view.ThreadSummary, err error) {
	model := dao.BbsThread.Ctx(ctx).
		LeftJoin("bbs_mythread mt", "mt.tid=bbs_thread.tid").
		LeftJoin("bbs_user u", "u.uid=bbs_thread.uid").
		LeftJoin("bbs_user lu", "lu.uid=bbs_thread.lastuid").
		LeftJoin("bbs_forum f", "f.fid=bbs_thread.fid").
		Fields("bbs_thread.tid,bbs_thread.uid,bbs_thread.fid,bbs_thread.subject,bbs_thread.create_date,bbs_thread.last_date,bbs_thread.lastuid,bbs_thread.views,bbs_thread.posts,bbs_thread.files,bbs_thread.closed,bbs_thread.top,u.username,u.avatar,lu.username AS last_username,f.name AS forum_name").
		Where("mt.uid", uid).
		OrderDesc("bbs_thread.tid").
		Limit(100)
	if err = model.Scan(&threads); err != nil {
		return nil, gerror.Wrap(err, "读取用户主题失败")
	}
	for i := range threads {
		formatThreadSummary(&threads[i])
	}
	return threads, nil
}

func (s *Service) ChangePassword(ctx context.Context, uid uint, oldPassword, newPassword, repeat string) (err error) {
	var user entity.BbsUser
	if newPassword != repeat {
		return gerror.New("两次输入的新密码不一致")
	}
	if _, err = normalizeClientPasswordHash(oldPassword); err != nil {
		return err
	}
	if _, err = normalizeClientPasswordHash(newPassword); err != nil {
		return err
	}
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取用户失败")
	}
	if user.Uid == 0 {
		return gerror.New("用户不存在")
	}
	oldServerPassword, passwordErr := xiunoPassword(oldPassword, user.Salt)
	if passwordErr != nil {
		return passwordErr
	}
	if user.Password != oldServerPassword {
		return gerror.New("旧密码不正确")
	}
	newServerPassword, passwordErr := xiunoPassword(newPassword, user.Salt)
	if passwordErr != nil {
		return passwordErr
	}
	if _, err = dao.BbsUser.Ctx(ctx).
		Where(do.BbsUser{Uid: uid}).
		Data(do.BbsUser{Password: newServerPassword}).Update(); err != nil {
		return gerror.Wrap(err, "修改密码失败")
	}
	return nil
}

func (s *Service) Stats(ctx context.Context) (stats view.Stats, err error) {
	if stats.Threads, err = dao.BbsThread.Ctx(ctx).Count(); err != nil {
		return stats, gerror.Wrap(err, "统计主题失败")
	}
	if stats.Posts, err = dao.BbsPost.Ctx(ctx).Count(); err != nil {
		return stats, gerror.Wrap(err, "统计帖子失败")
	}
	if stats.Users, err = dao.BbsUser.Ctx(ctx).Count(); err != nil {
		return stats, gerror.Wrap(err, "统计用户失败")
	}
	if stats.Onlines, err = s.originalRuntimeOnlines(ctx); err != nil {
		return stats, err
	}
	return stats, nil
}

func (s *Service) AdminThreads(ctx context.Context, keyword string) (threads []view.AdminThread, err error) {
	model := dao.BbsThread.Ctx(ctx).
		LeftJoin("bbs_user u", "u.uid=bbs_thread.uid").
		LeftJoin("bbs_forum f", "f.fid=bbs_thread.fid").
		Fields("bbs_thread.tid,bbs_thread.fid,bbs_thread.subject,bbs_thread.create_date,bbs_thread.views,bbs_thread.posts,bbs_thread.closed,u.username,f.name AS forum_name").
		OrderDesc("bbs_thread.tid").
		Limit(100)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		model = model.WhereLike("bbs_thread.subject", "%"+keyword+"%")
	}
	if err = model.Scan(&threads); err != nil {
		return nil, gerror.Wrap(err, "读取后台主题列表失败")
	}
	for i := range threads {
		threads[i].CreateTime = formatUnix(threads[i].CreateDate)
	}
	return threads, nil
}

func (s *Service) AdminUsers(ctx context.Context, keyword string) (users []view.AdminUser, err error) {
	model := dao.BbsUser.Ctx(ctx).
		LeftJoin("bbs_group g", "g.gid=bbs_user.gid").
		Fields("bbs_user.uid,bbs_user.username,bbs_user.email,bbs_user.gid,bbs_user.threads,bbs_user.posts,bbs_user.create_date,g.name AS group_name").
		OrderAsc("bbs_user.uid").
		Limit(100)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where("bbs_user.username LIKE ? OR bbs_user.email LIKE ?", like, like)
	}
	if err = model.Scan(&users); err != nil {
		return nil, gerror.Wrap(err, "读取后台用户列表失败")
	}
	for i := range users {
		users[i].CreateTime = formatUnix(users[i].CreateDate)
	}
	return users, nil
}

func (s *Service) Groups(ctx context.Context) (groups []view.GroupOption, err error) {
	if err = dao.BbsGroup.Ctx(ctx).
		Fields(dao.BbsGroup.Columns().Gid, dao.BbsGroup.Columns().Name).
		OrderAsc(dao.BbsGroup.Columns().Gid).
		Scan(&groups); err != nil {
		return nil, gerror.Wrap(err, "读取用户组失败")
	}
	return groups, nil
}

func (s *Service) AdminForums(ctx context.Context) (forums []view.AdminForum, err error) {
	if err = dao.BbsForum.Ctx(ctx).
		Fields("fid,name,`rank`,threads,todayposts,brief,announcement,accesson,moduids,icon").
		OrderDesc(dao.BbsForum.Columns().Rank).
		Scan(&forums); err != nil {
		return nil, gerror.Wrap(err, "读取后台板块列表失败")
	}
	for index := range forums {
		if forums[index].Icon != 0 {
			forums[index].IconURL = fmt.Sprintf("/upload/forum/%d.png?%d", forums[index].Fid, forums[index].Icon)
		} else {
			forums[index].IconURL = "/view/img/forum.png"
		}
	}
	return forums, nil
}

func (s *Service) ForumEditor(ctx context.Context, fid uint) (forum view.AdminForum, rules []view.ForumAccessRule, err error) {
	var (
		groups   []entity.BbsGroup
		accesses []entity.BbsForumAccess
	)
	if err = dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: fid}).Scan(&forum); err != nil {
		return forum, nil, gerror.Wrap(err, "读取板块失败")
	}
	if forum.Fid == 0 {
		return forum, nil, gerror.New("板块不存在")
	}
	if forum.Moderators, err = s.resolveModeratorNames(ctx, forum.Moduids); err != nil {
		return forum, nil, err
	}
	if err = dao.BbsGroup.Ctx(ctx).OrderAsc(dao.BbsGroup.Columns().Gid).Scan(&groups); err != nil {
		return forum, nil, gerror.Wrap(err, "读取用户组失败")
	}
	if err = dao.BbsForumAccess.Ctx(ctx).
		Where(do.BbsForumAccess{Fid: fid}).
		Scan(&accesses); err != nil {
		return forum, nil, gerror.Wrap(err, "读取板块访问规则失败")
	}
	accessByGID := make(map[uint]entity.BbsForumAccess, len(accesses))
	for _, access := range accesses {
		accessByGID[access.Gid] = access
	}
	for _, group := range groups {
		access, exists := accessByGID[group.Gid]
		if !exists && len(accesses) == 0 {
			access = entity.BbsForumAccess{
				Gid: group.Gid, Allowread: uint(group.Allowread), Allowthread: uint(group.Allowthread),
				Allowpost: uint(group.Allowpost), Allowattach: uint(group.Allowattach), Allowdown: uint(group.Allowdown),
			}
		}
		rules = append(rules, view.ForumAccessRule{
			Gid: group.Gid, Name: group.Name, Allowread: access.Allowread,
			Allowthread: access.Allowthread, Allowpost: access.Allowpost,
			Allowattach: access.Allowattach, Allowdown: access.Allowdown,
		})
	}
	return forum, rules, nil
}

func (s *Service) CreateForum(ctx context.Context, name string, rank uint, brief string) (fid uint, err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, gerror.New("板块名称不能为空")
	}
	if len([]rune(name)) > 16 {
		return 0, gerror.New("板块名称不能超过 16 个字")
	}
	forumID, err := dao.BbsForum.Ctx(ctx).Data(do.BbsForum{
		Name: name, Rank: rank, Brief: strings.TrimSpace(brief), CreateDate: uint(time.Now().Unix()),
	}).InsertAndGetId()
	if err != nil {
		return 0, gerror.Wrap(err, "创建板块失败")
	}
	return uint(forumID), nil
}

func (s *Service) UpdateForum(
	ctx context.Context, fid uint, name string, rank uint, brief, announcement, moderatorNames string,
	accesson uint, rules []view.ForumAccessRule,
) (err error) {
	var forum entity.BbsForum
	name = strings.TrimSpace(name)
	if name == "" {
		return gerror.New("板块名称不能为空")
	}
	if len([]rune(name)) > 16 {
		return gerror.New("板块名称不能超过 16 个字")
	}
	if err = dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: fid}).Scan(&forum); err != nil {
		return gerror.Wrap(err, "读取板块失败")
	}
	if forum.Fid == 0 {
		return gerror.New("板块不存在")
	}
	moduids, err := s.resolveModeratorUIDs(ctx, moderatorNames)
	if err != nil {
		return err
	}
	if accesson != 0 {
		accesson = 1
	}
	if _, err = dao.BbsForum.Ctx(ctx).
		Where(do.BbsForum{Fid: fid}).
		Data(do.BbsForum{
			Name: name, Rank: rank, Brief: brief, Announcement: announcement,
			Moduids: moduids, Accesson: accesson,
		}).Update(); err != nil {
		return gerror.Wrap(err, "更新板块失败")
	}
	if _, err = dao.BbsForumAccess.Ctx(ctx).Where(do.BbsForumAccess{Fid: fid}).Delete(); err != nil {
		return gerror.Wrap(err, "清理板块访问规则失败")
	}
	if accesson != 0 {
		for _, rule := range rules {
			if _, err = dao.BbsForumAccess.Ctx(ctx).Data(do.BbsForumAccess{
				Fid: fid, Gid: rule.Gid, Allowread: rule.Allowread, Allowthread: rule.Allowthread,
				Allowpost: rule.Allowpost, Allowattach: rule.Allowattach, Allowdown: rule.Allowdown,
			}).Insert(); err != nil {
				return gerror.Wrap(err, "写入板块访问规则失败")
			}
		}
	}
	return nil
}

func (s *Service) DeleteForum(ctx context.Context, fid uint) (err error) {
	if fid == 1 {
		return gerror.New("默认板块不能删除")
	}
	count, err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Fid: fid}).Count()
	if err != nil {
		return gerror.Wrap(err, "检查板块主题失败")
	}
	if count > 0 {
		return gerror.New("请先删除或移动板块中的主题")
	}
	if _, err = dao.BbsForumAccess.Ctx(ctx).Where(do.BbsForumAccess{Fid: fid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除板块访问规则失败")
	}
	if _, err = dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: fid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除板块失败")
	}
	return nil
}

func (s *Service) AdminGroups(ctx context.Context) (groups []view.GroupPermission, err error) {
	if err = dao.BbsGroup.Ctx(ctx).OrderAsc(dao.BbsGroup.Columns().Gid).Scan(&groups); err != nil {
		return nil, gerror.Wrap(err, "读取后台用户组失败")
	}
	return groups, nil
}

func (s *Service) UpdateGroup(ctx context.Context, group view.GroupPermission) (err error) {
	var current entity.BbsGroup
	if err = dao.BbsGroup.Ctx(ctx).Where(do.BbsGroup{Gid: group.Gid}).Scan(&current); err != nil {
		return gerror.Wrap(err, "读取用户组失败")
	}
	if current.Name == "" {
		return gerror.New("用户组不存在")
	}
	if group.Gid == 0 {
		group.Allowthread = 0
		group.Allowpost = 0
		group.Allowattach = 0
	}
	updates := do.BbsGroup{
		Name: group.Name, Creditsfrom: group.Creditsfrom, Creditsto: group.Creditsto,
		Allowread: group.Allowread, Allowthread: group.Allowthread, Allowpost: group.Allowpost,
		Allowattach: group.Allowattach, Allowdown: group.Allowdown,
	}
	if group.Gid >= 1 && group.Gid <= 5 {
		updates.Allowtop = group.Allowtop
		updates.Allowupdate = group.Allowupdate
		updates.Allowdelete = group.Allowdelete
		updates.Allowmove = group.Allowmove
		updates.Allowbanuser = group.Allowbanuser
		updates.Allowdeleteuser = group.Allowdeleteuser
		updates.Allowviewip = group.Allowviewip
	}
	if _, err = dao.BbsGroup.Ctx(ctx).
		Where(do.BbsGroup{Gid: group.Gid}).
		Data(updates).Update(); err != nil {
		return gerror.Wrap(err, "更新用户组失败")
	}
	return nil
}

func (s *Service) SetThreadClosed(ctx context.Context, tid, adminUID, closed uint) (err error) {
	var thread entity.BbsThread
	if closed != 0 {
		closed = 1
	}
	if err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Scan(&thread); err != nil {
		return gerror.Wrap(err, "读取主题失败")
	}
	if thread.Tid == 0 {
		return gerror.New("主题不存在")
	}
	if _, err = dao.BbsThread.Ctx(ctx).
		Where(do.BbsThread{Tid: tid}).
		Data(do.BbsThread{Closed: closed}).Update(); err != nil {
		return gerror.Wrap(err, "更新主题状态失败")
	}
	action := "open"
	if closed == 1 {
		action = "close"
	}
	return s.writeModlog(ctx, adminUID, tid, thread.Firstpid, thread.Subject, action)
}

// SetThreadTop mirrors route/mod.php action=top and thread_top_change().
// top: 0 cancel, 1 forum pin, 3 site-wide pin (admin/super-mod only).
func (s *Service) SetThreadTop(ctx context.Context, tid, operatorUID uint, top int) error {
	if top != 0 && top != 1 && top != 3 {
		return gerror.New("置顶范围不正确")
	}
	var (
		thread entity.BbsThread
		user   entity.BbsUser
	)
	if operatorUID == 0 {
		return gerror.New("请先登录")
	}
	if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: operatorUID}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取操作用户失败")
	}
	if user.Uid == 0 {
		return gerror.New("用户不存在")
	}
	if err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Scan(&thread); err != nil {
		return gerror.Wrap(err, "读取主题失败")
	}
	if thread.Tid == 0 {
		return gerror.New("主题不存在")
	}
	allowed, err := s.canModerate(ctx, user, uint(thread.Fid), "top")
	if err != nil {
		return err
	}
	if !allowed {
		return gerror.New("没有权限置顶该主题")
	}
	if top == 3 && user.Gid != 1 && user.Gid != 2 {
		return gerror.New("只有管理员可以设置全站置顶")
	}
	if thread.Top == top {
		return nil
	}
	if _, err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Data(do.BbsThread{Top: top}).Update(); err != nil {
		return gerror.Wrap(err, "更新主题置顶状态失败")
	}
	// Original always db_replace even when top=0; keep the same row shape.
	// PHP uses db_replace('thread_top', ...).
	if _, err = dao.BbsThreadTop.Ctx(ctx).Data(do.BbsThreadTop{
		Fid: thread.Fid, Tid: tid, Top: top,
	}).Replace(); err != nil {
		return gerror.Wrap(err, "写入置顶索引失败")
	}
	return s.writeModlog(ctx, operatorUID, tid, thread.Firstpid, thread.Subject, "top")
}

// SetThreadsTop applies the original multi-select mod-top action.

// SetThreadsClosed mirrors mod.php action=close for multi tidarr.
func (s *Service) SetThreadsClosed(ctx context.Context, tids []uint, operatorUID uint, closed uint) error {
	if len(tids) == 0 {
		return gerror.New("请选择主题")
	}
	if closed != 0 {
		closed = 1
	}
	var user entity.BbsUser
	if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: operatorUID}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取操作用户失败")
	}
	for _, tid := range tids {
		if tid == 0 {
			continue
		}
		var thread entity.BbsThread
		if err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Scan(&thread); err != nil {
			return gerror.Wrap(err, "读取主题失败")
		}
		if thread.Tid == 0 {
			continue
		}
		allowed, err := s.canModerate(ctx, user, uint(thread.Fid), "top")
		if err != nil {
			return err
		}
		// Original close uses allowtop permission.
		if !allowed {
			continue
		}
		if err := s.SetThreadClosed(ctx, tid, operatorUID, closed); err != nil {
			return err
		}
	}
	return nil
}

// DeleteThreads mirrors mod.php action=delete for multi tidarr.
func (s *Service) DeleteThreads(ctx context.Context, tids []uint, operatorUID uint) error {
	if len(tids) == 0 {
		return gerror.New("请选择主题")
	}
	var user entity.BbsUser
	if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: operatorUID}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取操作用户失败")
	}
	for _, tid := range tids {
		if tid == 0 {
			continue
		}
		var thread entity.BbsThread
		if err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Scan(&thread); err != nil {
			return err
		}
		if thread.Tid == 0 {
			continue
		}
		allowed, err := s.canModerate(ctx, user, uint(thread.Fid), "delete")
		if err != nil {
			return err
		}
		if !allowed {
			continue
		}
		if err := s.DeleteThread(ctx, tid, operatorUID); err != nil {
			return err
		}
	}
	return nil
}

// MoveThreads mirrors mod.php action=move.
func (s *Service) MoveThreads(ctx context.Context, tids []uint, operatorUID, newFID uint) error {
	if len(tids) == 0 {
		return gerror.New("请选择主题")
	}
	if newFID == 0 {
		return gerror.New("版块不存在")
	}
	var forum entity.BbsForum
	if err := dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: newFID}).Scan(&forum); err != nil {
		return gerror.Wrap(err, "读取目标版块失败")
	}
	if forum.Fid == 0 {
		return gerror.New("版块不存在")
	}
	var user entity.BbsUser
	if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: operatorUID}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取操作用户失败")
	}
	for _, tid := range tids {
		if tid == 0 {
			continue
		}
		var thread entity.BbsThread
		if err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Scan(&thread); err != nil {
			return err
		}
		if thread.Tid == 0 || uint(thread.Fid) == newFID {
			continue
		}
		allowed, err := s.canModerate(ctx, user, uint(thread.Fid), "move")
		if err != nil {
			return err
		}
		if !allowed {
			continue
		}
		oldFID := thread.Fid
		if _, err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Data(do.BbsThread{Fid: int(newFID)}).Update(); err != nil {
			return gerror.Wrap(err, "移动主题失败")
		}
		// Keep thread_top.fid in sync when present.
		_, _ = dao.BbsThreadTop.Ctx(ctx).Where(do.BbsThreadTop{Tid: tid}).Data(do.BbsThreadTop{Fid: int(newFID)}).Update()
		// Adjust forum thread counters.
		_, _ = dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: oldFID}).Data(do.BbsForum{Threads: gdb.Raw("GREATEST(threads - 1, 0)")}).Update()
		_, _ = dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: newFID}).Data(do.BbsForum{Threads: gdb.Raw("threads + 1")}).Update()
		if err = s.writeModlog(ctx, operatorUID, tid, thread.Firstpid, thread.Subject, "move"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) SetThreadsTop(ctx context.Context, tids []uint, operatorUID uint, top int) error {
	if len(tids) == 0 {
		return gerror.New("请选择主题")
	}
	for _, tid := range tids {
		if tid == 0 {
			continue
		}
		if err := s.SetThreadTop(ctx, tid, operatorUID, top); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) DeleteThread(ctx context.Context, tid, adminUID uint) (err error) {
	var (
		thread entity.BbsThread
		posts  []entity.BbsPost
	)
	if err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Scan(&thread); err != nil {
		return gerror.Wrap(err, "读取主题失败")
	}
	if thread.Tid == 0 {
		return gerror.New("主题不存在")
	}
	if err = dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Tid: tid}).Scan(&posts); err != nil {
		return gerror.Wrap(err, "读取主题帖子失败")
	}
	if err = s.deleteAttachmentsByThread(ctx, tid); err != nil {
		return err
	}
	if _, err = dao.BbsMypost.Ctx(ctx).Where(do.BbsMypost{Tid: tid}).Delete(); err != nil {
		return gerror.Wrap(err, "清理用户帖子索引失败")
	}
	if _, err = dao.BbsMythread.Ctx(ctx).Where(do.BbsMythread{Tid: tid}).Delete(); err != nil {
		return gerror.Wrap(err, "清理用户主题索引失败")
	}
	if _, err = dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Tid: tid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除帖子失败")
	}
	if _, err = dao.BbsThreadTop.Ctx(ctx).Where(do.BbsThreadTop{Tid: tid}).Delete(); err != nil {
		return gerror.Wrap(err, "清理主题置顶索引失败")
	}
	if _, err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除主题失败")
	}
	if _, err = dao.BbsForum.Ctx(ctx).
		Where(do.BbsForum{Fid: thread.Fid}).
		Data(do.BbsForum{Threads: gdb.Raw("GREATEST(threads - 1, 0)")}).Update(); err != nil {
		return gerror.Wrap(err, "更新板块计数失败")
	}
	postCounts := make(map[uint]int)
	for _, post := range posts {
		if post.Isfirst == 0 && post.Uid != 0 {
			postCounts[post.Uid]++
		}
	}
	for uid, count := range postCounts {
		if _, err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Data(do.BbsUser{
			Posts: gdb.Raw(fmt.Sprintf("GREATEST(posts - %d, 0)", count)),
		}).Update(); err != nil {
			return gerror.Wrap(err, "更新用户计数失败")
		}
	}
	if thread.Uid != 0 {
		if _, err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: thread.Uid}).Data(do.BbsUser{
			Threads: gdb.Raw("GREATEST(threads - 1, 0)"),
		}).Update(); err != nil {
			return gerror.Wrap(err, "更新主题用户计数失败")
		}
	}
	if err = s.writeModlog(ctx, adminUID, tid, thread.Firstpid, thread.Subject, "delete"); err != nil {
		return err
	}
	return s.SyncPHPRuntime(ctx, nil)
}

func (s *Service) UpdateUserGroup(ctx context.Context, uid, gid, adminUID uint) (err error) {
	var (
		user  entity.BbsUser
		group entity.BbsGroup
	)
	if uid == 1 && gid != 1 {
		return gerror.New("不能移除创始管理员权限")
	}
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取用户失败")
	}
	if user.Uid == 0 {
		return gerror.New("用户不存在")
	}
	if err = dao.BbsGroup.Ctx(ctx).Where(do.BbsGroup{Gid: gid}).Scan(&group); err != nil {
		return gerror.Wrap(err, "读取用户组失败")
	}
	if group.Gid == 0 {
		return gerror.New("用户组不存在")
	}
	if _, err = dao.BbsUser.Ctx(ctx).
		Where(do.BbsUser{Uid: uid}).
		Data(do.BbsUser{Gid: gid}).Update(); err != nil {
		return gerror.Wrap(err, "更新用户组失败")
	}
	return s.writeModlog(ctx, adminUID, 0, 0, user.Username, "user-group")
}

func (s *Service) requireForumPermission(ctx context.Context, fid, uid uint, permission string) (err error) {
	var user entity.BbsUser
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取用户权限失败")
	}
	if user.Uid == 0 {
		return gerror.New("用户不存在")
	}
	allowed, err := s.forumPermission(ctx, fid, user.Gid, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return gerror.New("当前用户组没有此板块的操作权限")
	}
	return nil
}

func (s *Service) forumPermission(ctx context.Context, fid, gid uint, permission string) (allowed bool, err error) {
	var (
		forum  entity.BbsForum
		group  entity.BbsGroup
		access entity.BbsForumAccess
	)
	if err = dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: fid}).Scan(&forum); err != nil {
		return false, gerror.Wrap(err, "读取板块权限失败")
	}
	if forum.Fid == 0 {
		return false, gerror.New("板块不存在")
	}
	if err = dao.BbsGroup.Ctx(ctx).Where(do.BbsGroup{Gid: gid}).Scan(&group); err != nil {
		return false, gerror.Wrap(err, "读取用户组权限失败")
	}
	switch permission {
	case "read":
		allowed = group.Allowread != 0
	case "thread":
		allowed = group.Allowthread != 0
	case "post":
		allowed = group.Allowpost != 0
	case "attach":
		allowed = group.Allowattach != 0
	case "down":
		allowed = group.Allowdown != 0
	}
	if !allowed {
		return false, nil
	}
	if forum.Accesson == 0 {
		return true, nil
	}
	if err = dao.BbsForumAccess.Ctx(ctx).
		Where(do.BbsForumAccess{Fid: fid, Gid: gid}).
		Scan(&access); err != nil {
		return false, gerror.Wrap(err, "读取板块访问规则失败")
	}
	switch permission {
	case "read":
		allowed = access.Allowread != 0
	case "thread":
		allowed = access.Allowthread != 0
	case "post":
		allowed = access.Allowpost != 0
	case "attach":
		allowed = access.Allowattach != 0
	case "down":
		allowed = access.Allowdown != 0
	}
	return allowed, nil
}

func (s *Service) canPostAction(ctx context.Context, uid, ownerUID, fid uint, permission string) (moderator bool, err error) {
	var (
		user entity.BbsUser
	)
	if uid == 0 {
		return false, gerror.New("请先登录")
	}
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return false, gerror.Wrap(err, "读取用户权限失败")
	}
	if user.Uid == 0 {
		return false, gerror.New("用户不存在")
	}
	allowed, err := s.forumPermission(ctx, fid, user.Gid, "post")
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, gerror.New("当前用户组没有帖子操作权限")
	}
	if moderator, err = s.canModerate(ctx, user, fid, permission); err != nil {
		return false, err
	}
	if uid != ownerUID && !moderator {
		return false, gerror.New("没有权限操作这个帖子")
	}
	return moderator, nil
}

func (s *Service) canModerate(ctx context.Context, user entity.BbsUser, fid uint, permission string) (allowed bool, err error) {
	var (
		forum entity.BbsForum
		group entity.BbsGroup
	)
	if user.Gid == 1 || user.Gid == 2 {
		return true, nil
	}
	if user.Gid != 3 && user.Gid != 4 {
		return false, nil
	}
	if err = dao.BbsGroup.Ctx(ctx).Where(do.BbsGroup{Gid: user.Gid}).Scan(&group); err != nil {
		return false, gerror.Wrap(err, "读取版主用户组失败")
	}
	switch permission {
	case "update":
		allowed = group.Allowupdate != 0
	case "delete":
		allowed = group.Allowdelete != 0
	case "top":
		allowed = group.Allowtop != 0
	case "move":
		allowed = group.Allowmove != 0
	}
	if !allowed {
		return false, nil
	}
	if err = dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: fid}).Scan(&forum); err != nil {
		return false, gerror.Wrap(err, "读取版主板块失败")
	}
	targetUID := fmt.Sprintf("%d", user.Uid)
	for _, value := range strings.Split(forum.Moduids, ",") {
		if strings.TrimSpace(value) == targetUID {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) resolveModeratorUIDs(ctx context.Context, names string) (result string, err error) {
	var user entity.BbsUser
	uids := make([]string, 0)
	seen := make(map[uint]bool)
	for _, name := range strings.Split(names, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		user = entity.BbsUser{}
		if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Username: name}).Scan(&user); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", gerror.Wrap(err, "查询版主用户失败")
		}
		if user.Uid == 0 || user.Gid > 4 || seen[user.Uid] {
			continue
		}
		seen[user.Uid] = true
		uids = append(uids, strconv.FormatUint(uint64(user.Uid), 10))
	}
	return strings.Join(uids, ","), nil
}

func (s *Service) resolveModeratorNames(ctx context.Context, uids string) (result string, err error) {
	var user entity.BbsUser
	names := make([]string, 0)
	for _, uidString := range strings.Split(uids, ",") {
		uid, convertErr := strconv.ParseUint(strings.TrimSpace(uidString), 10, 64)
		if convertErr != nil || uid == 0 {
			continue
		}
		user = entity.BbsUser{}
		if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uint(uid)}).Scan(&user); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", gerror.Wrap(err, "查询版主名称失败")
		}
		if user.Uid != 0 {
			names = append(names, user.Username)
		}
	}
	return strings.Join(names, ","), nil
}


func (s *Service) writeModlog(ctx context.Context, uid, tid, pid uint, subject, action string) (err error) {
	if _, err = dao.BbsModlog.Ctx(ctx).Data(do.BbsModlog{
		Uid: uid, Tid: tid, Pid: pid, Subject: subject, Comment: "GoFrame 管理后台", Rmbs: 0,
		CreateDate: uint(time.Now().Unix()), Action: action,
	}).Insert(); err != nil {
		return gerror.Wrap(err, "写入管理日志失败")
	}
	return nil
}

func xiunoPassword(clientHash, salt string) (string, error) {
	normalized, err := normalizeClientPasswordHash(clientHash)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(normalized+salt))), nil
}

func normalizeClientPasswordHash(clientHash string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(clientHash))
	decoded, err := hex.DecodeString(normalized)
	if err != nil || len(decoded) != md5.Size {
		return "", gerror.New("密码提交格式不正确，请刷新页面后重试")
	}
	return normalized, nil
}

func formatThreadSummary(thread *view.ThreadSummary) {
	if thread == nil {
		return
	}
	thread.CreateTime = humanDate(thread.CreateDate)
	thread.LastTime = humanDate(thread.LastDate)
	thread.TopClass = threadTopClass(thread.Top)
	thread.AvatarURL = avatarURL(thread.Uid, thread.Avatar)
}

func forumIconURL(fid, icon uint) string {
	if icon == 0 {
		return "/view/img/forum.png"
	}
	return fmt.Sprintf("/upload/forum/%d.png?%d", fid, icon)
}

func formatUnixDate(timestamp uint) string {
	if timestamp == 0 {
		return ""
	}
	return time.Unix(int64(timestamp), 0).Format("2006-01-02")
}

func formatUnix(timestamp uint) string {
	if timestamp == 0 {
		return "刚刚"
	}
	return time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04")
}
