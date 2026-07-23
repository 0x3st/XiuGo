package bbs

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

func formatPostMessage(message string, doctype int, gid uint) string {
	switch doctype {
	case 0:
		if gid == 1 {
			return message
		}
		// Xiuno 对普通用户的 HTML 使用白名单清洗。Go 版当前没有开放 HTML
		// 编辑器，因此对伪造的 doctype=0 请求采取更严格的完整转义。
		return phpHTMLSpecialChars(message)
	case 1:
		return xiunoTextToHTML(message)
	default:
		return phpHTMLSpecialChars(message)
	}
}

func xiunoTextToHTML(message string) string {
	formatted := phpHTMLSpecialChars(message)
	formatted = strings.ReplaceAll(formatted, " ", "&nbsp; ")
	formatted = strings.ReplaceAll(formatted, "\t", " &nbsp; &nbsp; &nbsp; &nbsp;")
	formatted = strings.ReplaceAll(formatted, "\r\n", "\n")
	return strings.ReplaceAll(formatted, "\n", "<br>")
}

// phpHTMLSpecialChars mirrors htmlspecialchars() with Xiuno's default ENT_COMPAT
// behavior: ampersands, double quotes and angle brackets are escaped, apostrophes are not.
func phpHTMLSpecialChars(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"\"", "&quot;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(value)
}

func quoteBrief(message string, maxLength int) string {
	message = htmlTagPattern.ReplaceAllString(message, "")
	message = phpHTMLSpecialChars(message)
	runes := []rune(message)
	if len(runes) <= maxLength {
		return message
	}
	return string(runes[:maxLength]) + " ... "
}

func avatarURL(uid, avatar uint) string {
	if uid == 0 || avatar == 0 {
		return "/view/img/avatar.png"
	}
	dir := fmt.Sprintf("%09d", uid)[:3]
	return fmt.Sprintf("/upload/avatar/%s/%d.png?%d", dir, uid, avatar)
}

func humanDate(timestamp uint) string {
	return humanDateAt(timestamp, time.Now())
}

func humanDateAt(timestamp uint, now time.Time) string {
	if timestamp == 0 {
		return "刚刚"
	}
	created := time.Unix(int64(timestamp), 0)
	seconds := int64(now.Sub(created).Seconds())
	switch {
	case seconds > 31_536_000:
		return created.Format("2006-1-2")
	case seconds > 2_592_000:
		return fmt.Sprintf("%d月前", seconds/2_592_000)
	case seconds > 86_400:
		return fmt.Sprintf("%d天前", seconds/86_400)
	case seconds > 3_600:
		return fmt.Sprintf("%d小时前", seconds/3_600)
	case seconds > 60:
		return fmt.Sprintf("%d分钟前", seconds/60)
	default:
		return fmt.Sprintf("%d秒前", seconds)
	}
}

func canReplyToThread(closed, gid uint) bool {
	return closed == 0 || (gid > 0 && gid <= 5)
}

func validDoctype(doctype int) bool {
	return doctype >= 0 && doctype <= 10
}

// highlightKeyword mirrors post_highlight_keyword(): case-insensitive wrap in <span class="red">.
func highlightKeyword(str, keyword string) string {
	keyword = strings.TrimSpace(keyword)
	if str == "" || keyword == "" {
		return str
	}
	lower := strings.ToLower(str)
	key := strings.ToLower(keyword)
	var b strings.Builder
	i := 0
	for {
		j := strings.Index(lower[i:], key)
		if j < 0 {
			b.WriteString(str[i:])
			break
		}
		j += i
		b.WriteString(str[i:j])
		b.WriteString(`<span class="red">`)
		b.WriteString(str[j : j+len(keyword)])
		b.WriteString(`</span>`)
		i = j + len(keyword)
	}
	return b.String()
}
