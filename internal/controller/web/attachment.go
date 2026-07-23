package web

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/0x3st/XiuGo/internal/model/view"
)

const pendingAttachmentSession = "xiuno_go_tmp_files"

func (c *Controller) AttachCreate(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		c.writeXiunoMessage(r, -1, "请先登录")
		return
	}
	pending, err := c.service.CreatePendingAttachment(
		r.Context(), user.Uid, r.GetForm("name").String(), r.GetForm("data").String(),
		r.GetForm("width").Uint(), r.GetForm("height").Uint(), r.GetForm("is_image").Int(),
	)
	if err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	items := c.pendingAttachments(r)
	pending.ID = nextPendingAttachmentID(items)
	items = append(items, pending)
	if err = c.setPendingAttachments(r, items); err != nil {
		_ = c.service.DeletePendingAttachment(r.Context(), pending)
		c.writeXiunoMessage(r, -1, "保存附件会话失败")
		return
	}
	c.writeXiunoMessage(r, 0, map[string]any{
		"url": pending.URL, "orgfilename": pending.Orgfilename,
		"filetype": pending.Filetype, "filesize": pending.Filesize,
		"width": pending.Width, "height": pending.Height, "isimage": pending.Isimage,
		"downloads": pending.Downloads, "aid": pending.ID,
	})
}

func (c *Controller) AttachDelete(r *ghttp.Request) {
	user := c.currentUser(r)
	if user.Uid == 0 {
		c.writeXiunoMessage(r, -1, "请先登录")
		return
	}
	aid := r.GetRouter("aid").String()
	if strings.HasPrefix(aid, "_") {
		items := c.pendingAttachments(r)
		for index, item := range items {
			if item.ID != aid {
				continue
			}
			if err := c.service.DeletePendingAttachment(r.Context(), item); err != nil {
				c.writeXiunoMessage(r, -1, err.Error())
				return
			}
			items = append(items[:index], items[index+1:]...)
			if err := c.setPendingAttachments(r, items); err != nil {
				c.writeXiunoMessage(r, -1, "更新附件会话失败")
				return
			}
			c.writeXiunoMessage(r, 0, "删除成功")
			return
		}
		c.writeXiunoMessage(r, -1, "临时附件不存在")
		return
	}
	parsed, err := strconv.ParseUint(aid, 10, 64)
	if err != nil || parsed == 0 {
		c.writeXiunoMessage(r, -1, "附件编号不正确")
		return
	}
	if err = c.service.DeleteAttachment(r.Context(), uint(parsed), user.Uid); err != nil {
		c.writeXiunoMessage(r, -1, err.Error())
		return
	}
	c.writeXiunoMessage(r, 0, "删除成功")
}

func (c *Controller) AttachDownload(r *ghttp.Request) {
	attachment, path, err := c.service.AttachmentForDownload(
		r.Context(), r.GetRouter("aid").Uint(), c.currentUser(r),
	)
	if err != nil {
		c.fail(r, err)
		return
	}
	r.Response.ServeFileDownload(path, attachment.Orgfilename)
}

func (c *Controller) pendingAttachments(r *ghttp.Request) []view.PendingAttachment {
	raw := r.Session.MustGet(pendingAttachmentSession).String()
	if raw == "" {
		return nil
	}
	var items []view.PendingAttachment
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func (c *Controller) setPendingAttachments(r *ghttp.Request, items []view.PendingAttachment) error {
	if len(items) == 0 {
		r.Session.MustRemove(pendingAttachmentSession)
		return nil
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return err
	}
	r.Session.MustSet(pendingAttachmentSession, string(encoded))
	return nil
}

func (c *Controller) clearPendingAttachments(r *ghttp.Request) {
	r.Session.MustRemove(pendingAttachmentSession)
}

func nextPendingAttachmentID(items []view.PendingAttachment) string {
	maximum := -1
	for _, item := range items {
		value, err := strconv.Atoi(strings.TrimPrefix(item.ID, "_"))
		if err == nil && value > maximum {
			maximum = value
		}
	}
	return "_" + strconv.Itoa(maximum+1)
}
