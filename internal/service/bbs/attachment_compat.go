package bbs

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/0x3st/XiuGo/internal/dao"
	"github.com/0x3st/XiuGo/internal/model/do"
	"github.com/0x3st/XiuGo/internal/model/entity"
	"github.com/0x3st/XiuGo/internal/model/view"
)

const originalAttachmentMaxBytes = 20_480_000

var originalAttachmentTypes = map[string][]string{
	"video":   {"av", "wmv", "wav", "wma", "avi", "rm", "rmvb", "mp4"},
	"music":   {"mp3", "mp4"},
	"exe":     {"exe", "bin"},
	"flash":   {"swf", "fla", "as"},
	"image":   {"gif", "jpg", "jpeg", "png", "bmp"},
	"office":  {"doc", "xls", "ppt", "docx", "xlsx", "pptx"},
	"pdf":     {"pdf"},
	"text":    {"c", "cpp", "cc", "txt"},
	"zip":     {"tar", "zip", "gz", "rar", "7z", "bz"},
	"book":    {"chm"},
	"torrent": {"bt", "torrent"},
	"font":    {"ttf", "font", "fon"},
}

var originalAttachmentAllowed = func() map[string]bool {
	result := make(map[string]bool)
	for _, extensions := range originalAttachmentTypes {
		for _, extension := range extensions {
			result[extension] = true
		}
	}
	return result
}()

// CreatePendingAttachment mirrors route/attach.php?action=create. The binary
// is written to upload/tmp and the caller stores the returned metadata in the
// current HTTP session until a post is created or updated.
func (s *Service) CreatePendingAttachment(
	ctx context.Context, uid uint, name, encodedData string, width, height uint, isimage int,
) (pending view.PendingAttachment, err error) {
	if uid == 0 {
		return pending, gerror.New("请先登录")
	}
	var (
		user  entity.BbsUser
		group entity.BbsGroup
	)
	if err = dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return pending, gerror.Wrap(err, "读取上传用户失败")
	}
	if user.Uid == 0 {
		return pending, gerror.New("用户不存在")
	}
	if err = dao.BbsGroup.Ctx(ctx).Where(do.BbsGroup{Gid: user.Gid}).Scan(&group); err != nil {
		return pending, gerror.Wrap(err, "读取上传权限失败")
	}
	if user.Gid != 1 && group.Allowattach == 0 {
		return pending, gerror.New("您无权上传")
	}
	data, err := decodeAttachmentData(encodedData)
	if err != nil {
		return pending, err
	}
	if len(data) == 0 {
		return pending, gerror.New("附件数据为空")
	}
	if len(data) > originalAttachmentMaxBytes {
		return pending, gerror.New("附件不能超过 20M")
	}
	extension := originalAttachmentExtension(name)
	storedExtension := extension
	if !originalAttachmentAllowed[extension] {
		storedExtension = "_" + extension
	}
	randomName, err := originalAttachmentRandomName(15)
	if err != nil {
		return pending, gerror.Wrap(err, "生成附件名称失败")
	}
	filename := strconv.FormatUint(uint64(uid), 10) + "_" + randomName + "." + storedExtension
	tmpRoot := filepath.Join(s.uploadRoot(ctx), "tmp")
	if err = os.MkdirAll(tmpRoot, 0o777); err != nil {
		return pending, gerror.Wrap(err, "创建附件临时目录失败")
	}
	path := filepath.Join(tmpRoot, filename)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return pending, gerror.Wrap(err, "创建附件临时文件失败")
	}
	if _, err = file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return pending, gerror.Wrap(err, "写入附件失败")
	}
	if err = file.Close(); err != nil {
		_ = os.Remove(path)
		return pending, gerror.Wrap(err, "保存附件失败")
	}
	return view.PendingAttachment{
		URL:         "/upload/tmp/" + filename,
		Path:        path,
		Orgfilename: name,
		Filetype:    originalAttachmentType(extension),
		Filesize:    uint(len(data)),
		Width:       width,
		Height:      height,
		Isimage:     isimage,
		Downloads:   0,
	}, nil
}

func (s *Service) DeletePendingAttachment(ctx context.Context, pending view.PendingAttachment) error {
	path, err := safePendingAttachmentPath(s.uploadRoot(ctx), pending.Path)
	if err != nil {
		return err
	}
	if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
		return gerror.Wrap(err, "删除临时附件失败")
	}
	return nil
}

// AssociatePendingAttachments mirrors attach_assoc_post(): tmp files move to
// upload/attach/<date>, bbs_attach rows are created, tmp URLs in the post are
// rewritten and the original images/files counters are refreshed.
func (s *Service) AssociatePendingAttachments(
	ctx context.Context, pid, uid uint, pending []view.PendingAttachment,
) error {
	if len(pending) == 0 {
		return nil
	}
	var post entity.BbsPost
	if err := dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: pid}).Scan(&post); err != nil {
		return gerror.Wrap(err, "读取附件所属帖子失败")
	}
	if post.Pid == 0 {
		return gerror.New("附件所属帖子不存在")
	}
	day := time.Now().Format(s.originalAttachmentDateLayout(ctx))
	destinationRoot := filepath.Join(s.uploadRoot(ctx), "attach", day)
	if err := os.MkdirAll(destinationRoot, 0o777); err != nil {
		return gerror.Wrap(err, "创建附件目录失败")
	}
	message := post.Message
	messageFmt := post.MessageFmt
	for _, item := range pending {
		source, err := safePendingAttachmentPath(s.uploadRoot(ctx), item.Path)
		if err != nil {
			return err
		}
		filename := filepath.Base(source)
		destination := filepath.Join(destinationRoot, filename)
		if err = moveAttachmentFile(source, destination); err != nil {
			return gerror.Wrap(err, "移动附件失败")
		}
		storedFilename := filepath.ToSlash(filepath.Join(day, filename))
		if _, err = dao.BbsAttach.Ctx(ctx).Data(do.BbsAttach{
			Tid: post.Tid, Pid: post.Pid, Uid: uid,
			Filesize: item.Filesize, Width: item.Width, Height: item.Height,
			Filename: storedFilename, Orgfilename: item.Orgfilename, Filetype: item.Filetype,
			CreateDate: uint(time.Now().Unix()), Comment: "", Downloads: 0, Isimage: item.Isimage,
		}).Insert(); err != nil {
			return gerror.Wrap(err, "保存附件记录失败")
		}
		destinationURL := "/upload/attach/" + storedFilename
		message = strings.ReplaceAll(message, item.URL, destinationURL)
		messageFmt = strings.ReplaceAll(messageFmt, item.URL, destinationURL)
	}
	if message != post.Message || messageFmt != post.MessageFmt {
		if _, err := dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: post.Pid}).Data(do.BbsPost{
			Message: message, MessageFmt: messageFmt,
		}).Update(); err != nil {
			return gerror.Wrap(err, "更新帖子附件地址失败")
		}
	}
	return s.refreshAttachmentCounts(ctx, post)
}

func (s *Service) AttachmentsForPosts(ctx context.Context, pids []uint) (map[uint][]view.Attachment, error) {
	result := make(map[uint][]view.Attachment)
	if len(pids) == 0 {
		return result, nil
	}
	var attachments []entity.BbsAttach
	if err := dao.BbsAttach.Ctx(ctx).
		WhereIn(dao.BbsAttach.Columns().Pid, pids).
		Where(do.BbsAttach{Isimage: 0}).
		OrderAsc(dao.BbsAttach.Columns().Aid).
		Scan(&attachments); err != nil {
		return nil, gerror.Wrap(err, "读取帖子附件失败")
	}
	for _, attachment := range attachments {
		item := attachmentView(attachment)
		result[uint(attachment.Pid)] = append(result[uint(attachment.Pid)], item)
	}
	return result, nil
}

func (s *Service) AttachmentsForPost(ctx context.Context, pid uint) ([]view.Attachment, error) {
	items, err := s.AttachmentsForPosts(ctx, []uint{pid})
	if err != nil {
		return nil, err
	}
	return items[pid], nil
}

func (s *Service) DeleteAttachment(ctx context.Context, aid, uid uint) error {
	if uid == 0 {
		return gerror.New("请先登录")
	}
	var (
		attachment entity.BbsAttach
		thread     entity.BbsThread
		user       entity.BbsUser
		post       entity.BbsPost
	)
	if err := dao.BbsAttach.Ctx(ctx).Where(do.BbsAttach{Aid: aid}).Scan(&attachment); err != nil {
		return gerror.Wrap(err, "读取附件失败")
	}
	if attachment.Aid == 0 {
		return gerror.New("附件不存在")
	}
	if err := dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: attachment.Tid}).Scan(&thread); err != nil {
		return gerror.Wrap(err, "读取附件主题失败")
	}
	if thread.Tid == 0 {
		return gerror.New("附件主题不存在")
	}
	if err := dao.BbsUser.Ctx(ctx).Where(do.BbsUser{Uid: uid}).Scan(&user); err != nil {
		return gerror.Wrap(err, "读取附件用户失败")
	}
	if user.Uid == 0 {
		return gerror.New("用户不存在")
	}
	if uint(attachment.Uid) != uid {
		allowed, err := s.canModerate(ctx, user, uint(thread.Fid), "delete")
		if err != nil {
			return err
		}
		if !allowed {
			return gerror.New("没有权限删除这个附件")
		}
	}
	if err := dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: attachment.Pid}).Scan(&post); err != nil {
		return gerror.Wrap(err, "读取附件帖子失败")
	}
	if err := s.deleteAttachmentRows(ctx, []entity.BbsAttach{attachment}); err != nil {
		return err
	}
	if post.Pid != 0 {
		return s.refreshAttachmentCounts(ctx, post)
	}
	return nil
}

func (s *Service) AttachmentForDownload(
	ctx context.Context, aid uint, viewer view.User,
) (attachment view.Attachment, path string, err error) {
	var (
		record entity.BbsAttach
		thread entity.BbsThread
	)
	if err = dao.BbsAttach.Ctx(ctx).Where(do.BbsAttach{Aid: aid}).Scan(&record); err != nil {
		return attachment, "", gerror.Wrap(err, "读取附件失败")
	}
	if record.Aid == 0 {
		return attachment, "", gerror.New("附件不存在")
	}
	if err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: record.Tid}).Scan(&thread); err != nil {
		return attachment, "", gerror.Wrap(err, "读取附件主题失败")
	}
	if thread.Tid == 0 {
		return attachment, "", gerror.New("附件主题不存在")
	}
	allowed, err := s.forumPermission(ctx, uint(thread.Fid), viewer.Gid, "down")
	if err != nil {
		return attachment, "", err
	}
	if !allowed {
		return attachment, "", gerror.New("当前用户组没有下载附件的权限")
	}
	path, err = safeAttachmentPath(s.uploadRoot(ctx), record.Filename)
	if err != nil {
		return attachment, "", err
	}
	if info, statErr := os.Stat(path); statErr != nil || info.IsDir() {
		return attachment, "", gerror.New("附件文件不存在")
	}
	if _, err = dao.BbsAttach.Ctx(ctx).Where(do.BbsAttach{Aid: aid}).Data(do.BbsAttach{
		Downloads: gdb.Raw("downloads + 1"),
	}).Update(); err != nil {
		return attachment, "", gerror.Wrap(err, "更新附件下载次数失败")
	}
	record.Downloads++
	return attachmentView(record), path, nil
}

func (s *Service) refreshAttachmentCounts(ctx context.Context, post entity.BbsPost) error {
	images, err := dao.BbsAttach.Ctx(ctx).Where(do.BbsAttach{Pid: post.Pid, Isimage: 1}).Count()
	if err != nil {
		return gerror.Wrap(err, "统计图片附件失败")
	}
	files, err := dao.BbsAttach.Ctx(ctx).Where(do.BbsAttach{Pid: post.Pid, Isimage: 0}).Count()
	if err != nil {
		return gerror.Wrap(err, "统计文件附件失败")
	}
	if _, err = dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Pid: post.Pid}).Data(do.BbsPost{
		Images: images, Files: files,
	}).Update(); err != nil {
		return gerror.Wrap(err, "更新帖子附件计数失败")
	}
	if post.Isfirst != 0 {
		if _, err = dao.BbsThread.Ctx(ctx).Where(do.BbsThread{Tid: post.Tid}).Data(do.BbsThread{
			Images: images, Files: files,
		}).Update(); err != nil {
			return gerror.Wrap(err, "更新主题附件计数失败")
		}
	}
	return nil
}

func (s *Service) uploadRoot(ctx context.Context) string {
	root := g.Cfg().MustGet(ctx, "xiuno.uploadPath", "../xiuno-bbs/upload").String()
	if absolute, err := filepath.Abs(root); err == nil {
		return absolute
	}
	return root
}

func (s *Service) originalAttachmentDateLayout(ctx context.Context) string {
	rule := "Ym"
	if content, err := os.ReadFile(filepath.Join(s.phpRoot(ctx), "conf", "conf.php")); err == nil {
		if configured := phpConfigString(content, "attach_dir_save_rule"); configured != "" {
			rule = configured
		}
	}
	var layout strings.Builder
	for _, token := range rule {
		switch token {
		case 'Y':
			layout.WriteString("2006")
		case 'y':
			layout.WriteString("06")
		case 'm':
			layout.WriteString("01")
		case 'n':
			layout.WriteString("1")
		case 'd':
			layout.WriteString("02")
		case 'j':
			layout.WriteString("2")
		case 'H':
			layout.WriteString("15")
		case 'i':
			layout.WriteString("04")
		case 's':
			layout.WriteString("05")
		default:
			layout.WriteRune(token)
		}
	}
	if layout.Len() == 0 {
		return "200601"
	}
	return layout.String()
}

func attachmentView(record entity.BbsAttach) view.Attachment {
	return view.Attachment{
		Aid: record.Aid, Tid: record.Tid, Pid: record.Pid, Uid: record.Uid,
		Filesize: record.Filesize, Filename: record.Filename, Orgfilename: record.Orgfilename,
		Filetype: record.Filetype, Downloads: record.Downloads, Isimage: record.Isimage,
		URL: "/upload/attach/" + record.Filename,
	}
}

func decodeAttachmentData(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if comma := strings.IndexByte(value, ','); comma >= 0 {
		value = value[comma+1:]
	}
	if value == "" {
		return nil, gerror.New("附件数据为空")
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(value)
	}
	if err != nil {
		return nil, gerror.New("附件数据格式不正确")
	}
	return decoded, nil
}

func originalAttachmentExtension(name string) string {
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	if len(extension) > 7 {
		extension = extension[:7]
	}
	var cleaned strings.Builder
	for _, character := range extension {
		if character == '_' || unicode.IsLetter(character) || unicode.IsDigit(character) {
			cleaned.WriteRune(character)
		}
	}
	if cleaned.Len() == 0 {
		return "attach"
	}
	return cleaned.String()
}

func originalAttachmentType(extension string) string {
	for filetype, extensions := range originalAttachmentTypes {
		for _, candidate := range extensions {
			if candidate == extension {
				return filetype
			}
		}
	}
	return "other"
}

func originalAttachmentRandomName(length int) (string, error) {
	const alphabet = "23456789abcdefghijkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ"
	random := make([]byte, length)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	result := make([]byte, length)
	for index := range random {
		result[index] = alphabet[int(random[index])%len(alphabet)]
	}
	return string(result), nil
}

func safePendingAttachmentPath(uploadRoot, path string) (string, error) {
	root := filepath.Clean(filepath.Join(uploadRoot, "tmp"))
	target := filepath.Clean(path)
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, target)
	}
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", gerror.New("临时附件路径不安全")
	}
	return target, nil
}

func moveAttachmentFile(source, destination string) error {
	if err := os.Rename(source, destination); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrInvalid) {
		// Rename can fail across filesystems; copy below for every failure so the
		// behavior remains equivalent to Xiuno's xn_copy fallback.
	}
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err = io.Copy(destinationFile, sourceFile); err != nil {
		_ = destinationFile.Close()
		_ = os.Remove(destination)
		return err
	}
	if err = destinationFile.Close(); err != nil {
		_ = os.Remove(destination)
		return err
	}
	return os.Remove(source)
}
