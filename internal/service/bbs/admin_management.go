package bbs

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/0x3st/XiuGo/internal/dao"
	"github.com/0x3st/XiuGo/internal/model/do"
	"github.com/0x3st/XiuGo/internal/model/entity"
	"github.com/0x3st/XiuGo/internal/model/view"
)

var (
	adminEmailPattern    = regexp.MustCompile(`^[\w\-.]+@[\w\-.]+(\.\w+)+$`)
	adminUsernamePattern = regexp.MustCompile(`^[\pL\pN_]+$`)
	systemGroupIDs       = map[uint]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 101: true}
)

func (s *Service) SyncAdminForums(ctx context.Context, submitted map[uint]view.AdminForum, icons map[uint]string) error {
	var existing []entity.BbsForum
	if err := dao.BbsForum.Ctx(ctx).Scan(&existing); err != nil {
		return gerror.Wrap(err, "读取板块列表失败")
	}
	existingByID := make(map[uint]entity.BbsForum, len(existing))
	for _, forum := range existing {
		existingByID[forum.Fid] = forum
	}
	ids := make([]uint, 0, len(submitted))
	for fid := range submitted {
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, fid := range ids {
		forum := submitted[fid]
		if _, exists := existingByID[fid]; exists {
			if _, err := dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: fid}).Data(do.BbsForum{
				Name: forum.Name, Rank: forum.Rank,
			}).Update(); err != nil {
				return gerror.Wrap(err, "更新板块失败")
			}
		} else {
			if _, err := dao.BbsForum.Ctx(ctx).Data(do.BbsForum{
				Fid: fid, Name: forum.Name, Rank: forum.Rank, CreateDate: uint(time.Now().Unix()),
			}).Insert(); err != nil {
				return gerror.Wrap(err, "创建板块失败")
			}
		}
		if dataURI := icons[fid]; dataURI != "" {
			if err := s.saveForumIcon(ctx, fid, dataURI); err != nil {
				return err
			}
		}
	}
	for _, forum := range existing {
		_, wasSubmitted := submitted[forum.Fid]
		if forum.Fid == 1 || wasSubmitted {
			continue
		}
		if err := s.deleteForumOriginal(ctx, forum.Fid); err != nil {
			return err
		}
	}
	return s.SyncPHPRuntime(ctx, nil)
}

func (s *Service) saveForumIcon(ctx context.Context, fid uint, dataURI string) error {
	comma := strings.IndexByte(dataURI, ',')
	if comma < 0 {
		return gerror.New("板块图标格式错误")
	}
	decoded, err := decodeBase64(dataURI[comma+1:])
	if err != nil {
		return gerror.Wrap(err, "解析板块图标失败")
	}
	directory := filepath.Join(s.phpRoot(ctx), "upload", "forum")
	if err = os.MkdirAll(directory, 0o755); err != nil {
		return gerror.Wrap(err, "创建板块图标目录失败")
	}
	if err = writeFileAtomic(filepath.Join(directory, strconv.FormatUint(uint64(fid), 10)+".png"), decoded); err != nil {
		return gerror.Wrap(err, "保存板块图标失败")
	}
	_, err = dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: fid}).Data(do.BbsForum{Icon: uint(time.Now().Unix())}).Update()
	return gerror.Wrap(err, "更新板块图标版本失败")
}

func (s *Service) deleteForumOriginal(ctx context.Context, fid uint) error {
	var threads []entity.BbsThread
	if err := dao.BbsThread.Ctx(ctx).Fields(dao.BbsThread.Columns().Tid).Where(do.BbsThread{Fid: fid}).Scan(&threads); err != nil {
		return gerror.Wrap(err, "读取板块主题失败")
	}
	for _, thread := range threads {
		if err := s.deleteThreadOriginal(ctx, thread.Tid); err != nil {
			return err
		}
	}
	if _, err := dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: fid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除板块失败")
	}
	if _, err := dao.BbsForumAccess.Ctx(ctx).Where(do.BbsForumAccess{Fid: fid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除板块权限失败")
	}
	return nil
}

func (s *Service) SyncAdminGroups(ctx context.Context, submitted map[uint]view.GroupPermission) error {
	var existing []entity.BbsGroup
	if err := dao.BbsGroup.Ctx(ctx).Scan(&existing); err != nil {
		return gerror.Wrap(err, "读取用户组失败")
	}
	existingByID := make(map[uint]entity.BbsGroup, len(existing))
	for _, group := range existing {
		existingByID[group.Gid] = group
	}
	ids := make([]uint, 0, len(submitted))
	for gid := range submitted {
		ids = append(ids, gid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, gid := range ids {
		group := submitted[gid]
		data := do.BbsGroup{Name: group.Name, Creditsfrom: group.Creditsfrom, Creditsto: group.Creditsto}
		if _, exists := existingByID[gid]; exists {
			if _, err := dao.BbsGroup.Ctx(ctx).Where(do.BbsGroup{Gid: gid}).Data(data).Update(); err != nil {
				return gerror.Wrap(err, "更新用户组失败")
			}
			continue
		}
		data.Gid = gid
		if _, err := dao.BbsGroup.Ctx(ctx).Data(data).Insert(); err != nil {
			return gerror.Wrap(err, "创建用户组失败")
		}
		var forums []entity.BbsForum
		if err := dao.BbsForum.Ctx(ctx).Scan(&forums); err != nil {
			return gerror.Wrap(err, "读取板块权限失败")
		}
		for _, forum := range forums {
			if forum.Accesson == 0 {
				continue
			}
			if _, err := dao.BbsForumAccess.Ctx(ctx).Data(do.BbsForumAccess{Fid: forum.Fid, Gid: gid}).Insert(); err != nil {
				return gerror.Wrap(err, "填充新用户组板块权限失败")
			}
		}
	}
	for _, group := range existing {
		_, wasSubmitted := submitted[group.Gid]
		if systemGroupIDs[group.Gid] || wasSubmitted {
			continue
		}
		if _, err := dao.BbsGroup.Ctx(ctx).Where(do.BbsGroup{Gid: group.Gid}).Delete(); err != nil {
			return gerror.Wrap(err, "删除用户组失败")
		}
		if _, err := dao.BbsForumAccess.Ctx(ctx).Where(do.BbsForumAccess{Gid: group.Gid}).Delete(); err != nil {
			return gerror.Wrap(err, "删除用户组板块权限失败")
		}
	}
	return nil
}

func (s *Service) AdminGroup(ctx context.Context, gid uint) (group view.GroupPermission, err error) {
	if err = dao.BbsGroup.Ctx(ctx).Where(do.BbsGroup{Gid: gid}).Scan(&group); err != nil {
		return group, gerror.Wrap(err, "读取用户组失败")
	}
	if group.Name == "" {
		return group, gerror.New("用户组不存在")
	}
	return group, nil
}

func (s *Service) AdminModeratorNames(ctx context.Context, uids string) (string, error) {
	return s.resolveModeratorNames(ctx, uids)
}

func (s *Service) AdminUserList(ctx context.Context, searchType, keyword string, page int) (result view.AdminUserPage, err error) {
	const pageSize = 20
	allowed := map[string]bool{"uid": true, "username": true, "email": true, "gid": true, "create_ip": true}
	if page < 1 {
		page = 1
	}
	if !allowed[searchType] {
		searchType = "uid"
	}
	keyword = strings.TrimSpace(keyword)
	model := dao.BbsUser.Ctx(ctx)
	if keyword != "" {
		value := any(keyword)
		if searchType == "create_ip" {
			value = ipv4Long(keyword)
		}
		model = model.Where(searchType, value)
	}
	if result.Total, err = model.Count(); err != nil {
		return result, gerror.Wrap(err, "统计用户失败")
	}
	listModel := dao.BbsUser.Ctx(ctx).
		LeftJoin("bbs_group g", "g.gid=bbs_user.gid").
		Fields("bbs_user.uid,bbs_user.username,bbs_user.email,bbs_user.gid,bbs_user.threads,bbs_user.posts,bbs_user.create_date,bbs_user.create_ip,g.name AS group_name").
		OrderDesc("bbs_user.uid").Page(page, pageSize)
	if keyword != "" {
		value := any(keyword)
		if searchType == "create_ip" {
			value = ipv4Long(keyword)
		}
		listModel = listModel.Where("bbs_user."+searchType, value)
	}
	if err = listModel.Scan(&result.Users); err != nil {
		return result, gerror.Wrap(err, "读取用户列表失败")
	}
	for index := range result.Users {
		result.Users[index].CreateTime = formatAdminDate(result.Users[index].CreateDate)
		result.Users[index].CreateIP = longIPv4(result.Users[index].CreateIp)
	}
	result.SearchType = searchType
	result.Keyword = keyword
	result.Page = page
	result.Pages = (result.Total + pageSize - 1) / pageSize
	if result.Pages < 1 {
		result.Pages = 1
	}
	return result, nil
}

func (s *Service) AdminUser(ctx context.Context, uid uint) (user entity.BbsUser, err error) {
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return user, gerror.Wrap(err, "读取用户失败")
	}
	if user.Uid == 0 {
		return user, gerror.New("用户不存在")
	}
	return user, nil
}

func (s *Service) CreateAdminUser(ctx context.Context, email, username, password string, gid, createIP uint) error {
	if err := validateAdminUser(email, username, true); err != nil {
		return err
	}
	if err := s.ensureAdminUserUnique(ctx, 0, email, username, true); err != nil {
		return err
	}
	salt, err := xiunoSalt()
	if err != nil {
		return gerror.Wrap(err, "生成用户密码盐失败")
	}
	serverPassword, err := xiunoPassword(password, salt)
	if err != nil {
		return err
	}
	if _, err = dao.BbsUser.Ctx(ctx).Data(do.BbsUser{
		Email: email, Username: username, Password: serverPassword, Salt: salt,
		Gid: gid, CreateIp: createIP, CreateDate: uint(time.Now().Unix()),
	}).Insert(); err != nil {
		return gerror.Wrap(err, "创建用户失败")
	}
	return s.SyncPHPRuntime(ctx, map[string]int{"todayusers": 1})
}

func (s *Service) UpdateAdminUser(ctx context.Context, uid uint, email, username, password string, gid uint) error {
	old, err := s.AdminUser(ctx, uid)
	if err != nil {
		return err
	}
	if err = validateAdminUser(email, username, false); err != nil {
		return err
	}
	if err = s.ensureAdminUserUnique(ctx, uid, email, username, false); err != nil {
		return err
	}
	updates := do.BbsUser{}
	changed := false
	if old.Email != email {
		updates.Email, changed = email, true
	}
	if old.Username != username {
		updates.Username, changed = username, true
	}
	if old.Gid != gid {
		updates.Gid, changed = gid, true
	}
	if password != "" {
		salt, saltErr := xiunoSalt()
		if saltErr != nil {
			return gerror.Wrap(saltErr, "生成用户密码盐失败")
		}
		serverPassword, passwordErr := xiunoPassword(password, salt)
		if passwordErr != nil {
			return passwordErr
		}
		updates.Salt = salt
		updates.Password = serverPassword
		changed = true
	}
	if !changed {
		return gerror.New("数据没有变化")
	}
	if _, err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Data(updates).Update(); err != nil {
		return gerror.Wrap(err, "更新用户失败")
	}
	return nil
}

func (s *Service) DeleteAdminUser(ctx context.Context, uid uint) error {
	user, err := s.AdminUser(ctx, uid)
	if err != nil {
		return err
	}
	if user.Gid == 1 {
		return gerror.New("管理员不能被删除")
	}
	var authoredThreads []entity.BbsThread
	if err = dao.BbsThread.Ctx(ctx).
		Fields(dao.BbsThread.Columns().Tid).
		Where(do.BbsThread{Uid: uid}).
		OrderDesc(dao.BbsThread.Columns().Tid).
		Scan(&authoredThreads); err != nil {
		return gerror.Wrap(err, "读取用户主题失败")
	}
	for _, thread := range authoredThreads {
		if err = s.deleteThreadOriginal(ctx, thread.Tid); err != nil {
			return err
		}
	}
	var replies []entity.BbsPost
	if err = dao.BbsPost.Ctx(ctx).
		Where(do.BbsPost{Uid: uid, Isfirst: 0}).
		OrderDesc(dao.BbsPost.Columns().Pid).
		Scan(&replies); err != nil {
		return gerror.Wrap(err, "读取用户回帖失败")
	}
	for _, reply := range replies {
		if err = s.deleteReplyOriginal(ctx, reply); err != nil {
			return err
		}
	}
	if _, err = dao.BbsMythread.Ctx(ctx).Where(do.BbsMythread{Uid: uid}).Delete(); err != nil {
		return gerror.Wrap(err, "清理用户主题索引失败")
	}
	if _, err = dao.BbsMypost.Ctx(ctx).Where(do.BbsMypost{Uid: uid}).Delete(); err != nil {
		return gerror.Wrap(err, "清理用户帖子索引失败")
	}
	if err = s.deleteAttachmentsByUser(ctx, uid); err != nil {
		return err
	}
	if _, err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除用户失败")
	}
	return s.SyncPHPRuntime(ctx, nil)
}

func (s *Service) CreateAdminThreadQueue(ctx context.Context) (uint, error) {
	queueID := uint(time.Now().Unix())
	if _, err := dao.BbsQueue.Ctx(ctx).Where(do.BbsQueue{Queueid: queueID}).Delete(); err != nil {
		return 0, gerror.Wrap(err, "初始化主题队列失败")
	}
	return queueID, nil
}

func (s *Service) ScanAdminThreads(ctx context.Context, queueID uint, scan view.AdminThreadScan) ([]uint, error) {
	if scan.Page < 1 {
		scan.Page = 1
	}
	if scan.Page == 1 {
		if _, err := dao.BbsQueue.Ctx(ctx).Where(do.BbsQueue{Queueid: queueID}).Delete(); err != nil {
			return nil, gerror.Wrap(err, "清空主题队列失败")
		}
	}
	model := dao.BbsThread.Ctx(ctx).OrderDesc(dao.BbsThread.Columns().Lastpid).Page(scan.Page, 100)
	if scan.Fid != 0 {
		model = model.Where(do.BbsThread{Fid: scan.Fid})
	}
	var threads []entity.BbsThread
	if err := model.Scan(&threads); err != nil {
		return nil, gerror.Wrap(err, "扫描主题失败")
	}
	tids := make([]uint, 0, len(threads))
	for _, thread := range threads {
		if scan.CreateDateStart != 0 && thread.CreateDate < scan.CreateDateStart ||
			scan.CreateDateEnd != 0 && thread.CreateDate > scan.CreateDateEnd ||
			scan.Uid != 0 && thread.Uid != scan.Uid || scan.UserIP != 0 && thread.Userip != scan.UserIP ||
			scan.Keyword != "" && !strings.Contains(strings.ToLower(thread.Subject), strings.ToLower(scan.Keyword)) {
			continue
		}
		tids = append(tids, thread.Tid)
		_, err := dao.BbsQueue.Ctx(ctx).Data(do.BbsQueue{Queueid: queueID, V: thread.Tid, Expiry: uint(time.Now().Unix() + 86400)}).Insert()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, gerror.Wrap(err, "写入主题队列失败")
		}
	}
	return tids, nil
}

func (s *Service) OperateAdminThreadQueue(ctx context.Context, queueID, adminUID uint, operation string) ([]uint, error) {
	var rows []entity.BbsQueue
	if err := dao.BbsQueue.Ctx(ctx).Where(do.BbsQueue{Queueid: queueID}).Limit(101).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "读取主题队列失败")
	}
	tids := make([]uint, 0, len(rows))
	for _, row := range rows {
		tid := uint(row.V)
		switch operation {
		case "delete":
			if err := s.deleteThreadOriginal(ctx, tid); err != nil {
				return tids, err
			}
		case "close":
			if _, err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Data(do.BbsThread{Closed: 1}).Update(); err != nil {
				return tids, gerror.Wrap(err, "关闭主题失败")
			}
		case "open":
			if _, err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Data(do.BbsThread{Closed: 0}).Update(); err != nil {
				return tids, gerror.Wrap(err, "打开主题失败")
			}
		default:
			return tids, gerror.New("未知主题操作")
		}
		if _, err := dao.BbsQueue.Ctx(ctx).Where(do.BbsQueue{Queueid: queueID, V: row.V}).Delete(); err != nil {
			return tids, gerror.Wrap(err, "弹出主题队列失败")
		}
		tids = append(tids, tid)
	}
	if operation == "delete" {
		if err := s.SyncPHPRuntime(ctx, nil); err != nil {
			return tids, err
		}
	}
	_ = adminUID
	return tids, nil
}

func (s *Service) AdminThreadQueue(ctx context.Context, queueID uint, page int) ([]view.AdminThread, int, error) {
	if page < 1 {
		page = 1
	}
	total, err := dao.BbsQueue.Ctx(ctx).Where(do.BbsQueue{Queueid: queueID}).Count()
	if err != nil {
		return nil, 0, gerror.Wrap(err, "统计主题队列失败")
	}
	var rows []entity.BbsQueue
	if err = dao.BbsQueue.Ctx(ctx).Where(do.BbsQueue{Queueid: queueID}).Page(page, 100).Scan(&rows); err != nil {
		return nil, total, gerror.Wrap(err, "读取主题队列失败")
	}
	if len(rows) == 0 {
		return []view.AdminThread{}, total, nil
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.V)
	}
	var threads []view.AdminThread
	if err = dao.BbsThread.Ctx(ctx).
		LeftJoin("bbs_user u", "u.uid=bbs_thread.uid").
		LeftJoin("bbs_forum f", "f.fid=bbs_thread.fid").
		Fields("bbs_thread.tid,bbs_thread.fid,bbs_thread.subject,bbs_thread.create_date,bbs_thread.views,bbs_thread.posts,bbs_thread.closed,u.username,f.name AS forum_name").
		WhereIn("bbs_thread.tid", ids).OrderDesc("bbs_thread.lastpid").Scan(&threads); err != nil {
		return nil, total, gerror.Wrap(err, "读取队列主题失败")
	}
	for index := range threads {
		threads[index].CreateTime = formatUnix(threads[index].CreateDate)
	}
	return threads, total, nil
}

// ClearOriginalCache rebuilds Go-shared runtime storage and clears temporary
// upload files. It no longer targets PHP compile caches for dual-runtime sync.
func (s *Service) ClearOriginalCache(ctx context.Context, clearCache, clearTmp bool) error {
	if clearCache {
		// GoFrame forbids DELETE without WHERE; match all rows explicitly.
		if _, err := dao.BbsCache.Ctx(ctx).Where("1=1").Delete(); err != nil {
			return gerror.Wrap(err, "清空运行数据缓存失败")
		}
		// Go still stores bbs_runtime in bbs_cache; rebuild after truncate.
		if err := s.SyncPHPRuntime(ctx, nil); err != nil {
			return err
		}
	}
	if !clearTmp {
		return nil
	}
	// Temporary attachment staging used by Go uploads (and legacy PHP upload/tmp).
	roots := []string{
		filepath.Join(s.uploadRoot(ctx), "tmp"),
	}
	for _, path := range roots {
		entries, err := os.ReadDir(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return gerror.Wrap(err, "读取临时目录失败")
		}
		for _, entry := range entries {
			if entry.Name() == ".gitkeep" {
				continue
			}
			if err = os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
				return gerror.Wrap(err, "清理临时目录失败")
			}
		}
	}
	return nil
}

func (s *Service) AdminPlugins(ctx context.Context) ([]view.AdminPlugin, error) {
	_ = ctx
	return []view.AdminPlugin{}, nil
}

func (s *Service) SetAdminPluginState(ctx context.Context, dir, action string) error {
	_ = ctx
	_ = dir
	_ = action
	return gerror.New("XiuGo 不支持原版 PHP 插件")
}

func (s *Service) deleteThreadOriginal(ctx context.Context, tid uint) error {
	var (
		thread      entity.BbsThread
		replyCounts []struct {
			Uid   uint `orm:"uid"`
			Posts int  `orm:"posts"`
		}
	)
	if err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Scan(&thread); err != nil {
		return gerror.Wrap(err, "读取主题失败")
	}
	if thread.Tid == 0 {
		return nil
	}
	if err := dao.BbsPost.Ctx(ctx).
		Fields(dao.BbsPost.Columns().Uid+", COUNT(*) AS posts").
		Where(do.BbsPost{Tid: tid, Isfirst: 0}).
		WhereGT(dao.BbsPost.Columns().Uid, 0).
		Group(dao.BbsPost.Columns().Uid).
		Scan(&replyCounts); err != nil {
		return gerror.Wrap(err, "统计主题回复用户失败")
	}
	if err := s.deleteAttachmentsByThread(ctx, tid); err != nil {
		return err
	}
	for _, count := range replyCounts {
		if _, err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: count.Uid}).Data(do.BbsUser{
			Posts: gdb.Raw(fmt.Sprintf("GREATEST(posts - %d, 0)", count.Posts)),
		}).Update(); err != nil {
			return gerror.Wrap(err, "更新回帖用户统计失败")
		}
	}
	if _, err := dao.BbsMypost.Ctx(ctx).Where(do.BbsMypost{Tid: tid}).Delete(); err != nil {
		return gerror.Wrap(err, "清理用户帖子索引失败")
	}
	if _, err := dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Tid: tid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除主题帖子失败")
	}
	if _, err := dao.BbsThreadTop.Ctx(ctx).Where(do.BbsThreadTop{Tid: tid}).Delete(); err != nil {
		return gerror.Wrap(err, "清理主题置顶索引失败")
	}
	if thread.Uid != 0 {
		if _, err := dao.BbsMythread.Ctx(ctx).Where(do.BbsMythread{Uid: thread.Uid, Tid: tid}).Delete(); err != nil {
			return gerror.Wrap(err, "删除用户主题索引失败")
		}
	}
	if _, err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: tid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除主题失败")
	}
	if _, err := dao.BbsForum.Ctx(ctx).Where(do.BbsForum{Fid: thread.Fid}).Data(do.BbsForum{Threads: gdb.Raw("GREATEST(threads - 1, 0)")}).Update(); err != nil {
		return gerror.Wrap(err, "更新板块主题数失败")
	}
	if thread.Uid != 0 {
		if _, err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: thread.Uid}).Data(do.BbsUser{Threads: gdb.Raw("GREATEST(threads - 1, 0)")}).Update(); err != nil {
			return gerror.Wrap(err, "更新用户主题数失败")
		}
	}
	return nil
}

func (s *Service) deleteAttachmentsByThread(ctx context.Context, tid uint) error {
	var attachments []entity.BbsAttach
	if err := dao.BbsAttach.Ctx(ctx).Where(do.BbsAttach{Tid: tid}).Scan(&attachments); err != nil {
		return gerror.Wrap(err, "读取主题附件失败")
	}
	return s.deleteAttachmentRows(ctx, attachments)
}

func (s *Service) deleteAttachmentsByPost(ctx context.Context, pid uint) error {
	var attachments []entity.BbsAttach
	if err := dao.BbsAttach.Ctx(ctx).Where(do.BbsAttach{Pid: pid}).Scan(&attachments); err != nil {
		return gerror.Wrap(err, "读取帖子附件失败")
	}
	return s.deleteAttachmentRows(ctx, attachments)
}

func (s *Service) deleteAttachmentsByUser(ctx context.Context, uid uint) error {
	var attachments []entity.BbsAttach
	if err := dao.BbsAttach.Ctx(ctx).Where(do.BbsAttach{Uid: uid}).Scan(&attachments); err != nil {
		return gerror.Wrap(err, "读取用户附件失败")
	}
	return s.deleteAttachmentRows(ctx, attachments)
}

func (s *Service) deleteAttachmentRows(ctx context.Context, attachments []entity.BbsAttach) error {
	uploadPath := g.Cfg().MustGet(ctx, "xiuno.uploadPath", "../xiuno-bbs/upload").String()
	if absolute, err := filepath.Abs(uploadPath); err == nil {
		uploadPath = absolute
	}
	for _, attachment := range attachments {
		if attachment.Filename != "" {
			if path, pathErr := safeAttachmentPath(uploadPath, attachment.Filename); pathErr == nil {
				_ = os.Remove(path)
			}
		}
		if _, err := dao.BbsAttach.Ctx(ctx).Where(do.BbsAttach{Aid: attachment.Aid}).Delete(); err != nil {
			return gerror.Wrap(err, "删除附件记录失败")
		}
	}
	return nil
}

func (s *Service) deleteReplyOriginal(ctx context.Context, post entity.BbsPost) error {
	if post.Pid == 0 || post.Isfirst != 0 {
		return gerror.New("只能按回复执行删除")
	}
	var (
		thread   entity.BbsThread
		lastPost entity.BbsPost
	)
	if err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: post.Tid}).Scan(&thread); err != nil {
		return gerror.Wrap(err, "读取回复主题失败")
	}
	if err := s.deleteAttachmentsByPost(ctx, post.Pid); err != nil {
		return err
	}
	if _, err := dao.BbsMypost.Ctx(ctx).Where(do.BbsMypost{Pid: post.Pid}).Delete(); err != nil {
		return gerror.Wrap(err, "清理用户帖子索引失败")
	}
	if _, err := dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: post.Pid}).Delete(); err != nil {
		return gerror.Wrap(err, "删除回复失败")
	}
	if post.Uid != 0 {
		if _, err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: post.Uid}).Data(do.BbsUser{
			Posts: gdb.Raw("GREATEST(posts - 1, 0)"),
		}).Update(); err != nil {
			return gerror.Wrap(err, "更新用户帖子数失败")
		}
	}
	if thread.Tid == 0 {
		return nil
	}
	if _, err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: post.Tid}).Data(do.BbsThread{
		Posts: gdb.Raw("GREATEST(posts - 1, 0)"),
	}).Update(); err != nil {
		return gerror.Wrap(err, "更新主题回复数失败")
	}
	if post.Pid != thread.Lastpid {
		return nil
	}
	if err := dao.BbsPost.Ctx(ctx).
		Where(do.BbsPost{Tid: post.Tid}).
		OrderDesc(dao.BbsPost.Columns().Pid).
		Limit(1).
		Scan(&lastPost); err != nil {
		return gerror.Wrap(err, "读取最后回复失败")
	}
	if _, err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: post.Tid}).Data(do.BbsThread{
		Lastpid: lastPost.Pid, Lastuid: lastPost.Uid, LastDate: lastPost.CreateDate,
	}).Update(); err != nil {
		return gerror.Wrap(err, "更新最后回复索引失败")
	}
	return nil
}

func safeAttachmentPath(uploadPath, filename string) (string, error) {
	root := filepath.Clean(filepath.Join(uploadPath, "attach"))
	relative := filepath.Clean(filepath.FromSlash(filename))
	if filepath.IsAbs(relative) || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", gerror.New("附件路径不安全")
	}
	target := filepath.Join(root, relative)
	withinRoot, err := filepath.Rel(root, target)
	if err != nil || withinRoot == ".." || strings.HasPrefix(withinRoot, ".."+string(filepath.Separator)) {
		return "", gerror.New("附件路径超出上传目录")
	}
	return target, nil
}

func (s *Service) ensureAdminUserUnique(ctx context.Context, uid uint, email, username string, creating bool) error {
	var user entity.BbsUser
	if email != "" || creating {
		if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Email: email}).Scan(&user); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return gerror.Wrap(err, "检查邮箱失败")
		}
		if user.Uid != 0 && user.Uid != uid {
			return gerror.New("邮箱已经被使用")
		}
	}
	if username != "" || creating {
		user = entity.BbsUser{}
		if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Username: username}).Scan(&user); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return gerror.Wrap(err, "检查用户名失败")
		}
		if user.Uid != 0 && user.Uid != uid {
			return gerror.New("用户已经存在")
		}
	}
	return nil
}

func validateAdminUser(email, username string, requireEmail bool) error {
	if requireEmail && email == "" {
		return gerror.New("请输入邮箱")
	}
	if email != "" && !adminEmailPattern.MatchString(email) {
		return gerror.New("邮箱格式不正确")
	}
	if utf8.RuneCountInString(username) > 16 {
		return gerror.New("用户名不能超过 16 个字符")
	}
	if username != "" && !adminUsernamePattern.MatchString(username) {
		return gerror.New("用户名格式不正确")
	}
	return nil
}

func xiunoSalt() (string, error) {
	const alphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
	buffer := make([]byte, 16)
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	for index := range buffer {
		buffer[index] = alphabet[int(random[index])%len(alphabet)]
	}
	return string(buffer), nil
}

func ipv4Long(value string) uint {
	ip := net.ParseIP(value).To4()
	if ip == nil {
		return 0
	}
	return uint(binary.BigEndian.Uint32(ip))
}

func longIPv4(value uint) string {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, uint32(value))
	return net.IP(buffer).String()
}

func formatAdminDate(value uint) string {
	if value == 0 {
		return "0000-00-00"
	}
	return time.Unix(int64(value), 0).Format("2006-01-02")
}

func adminDate(value string) uint {
	if value == "" {
		return 0
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return 0
	}
	return uint(parsed.Unix())
}

func decodeBase64(value string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(value)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func adminPageURL(searchType, keyword string, page int) string {
	return fmt.Sprintf("user-list-%s-%s-%d", searchType, keyword, page)
}
