package bbs

import (
	"fmt"
	"html"
	"math"
	"strings"
)

// Pagination is the view data for original-style Bootstrap pagination.
type Pagination struct {
	HTML       string
	Page       int
	Pages      int
	Total      int
	PageSize   int
	HasPrev    bool
	HasNext    bool
	PrevPage   int
	NextPage   int
}

// buildPagination mirrors model/misc.func.php pagination() loosely:
// numbered links with prev/next for Bootstrap 4.
// urlPattern must contain "{page}" placeholder.
func buildPagination(urlPattern string, total, page, pageSize int) Pagination {
	if pageSize < 1 {
		pageSize = 20
	}
	if page < 1 {
		page = 1
	}
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		page = pages
	}
	result := Pagination{
		Page: page, Pages: pages, Total: total, PageSize: pageSize,
		HasPrev: page > 1, HasNext: page < pages,
		PrevPage: page - 1, NextPage: page + 1,
	}
	if total <= pageSize {
		result.HTML = ""
		return result
	}

	link := func(p int, label string, active, disabled bool) string {
		href := strings.ReplaceAll(urlPattern, "{page}", fmt.Sprintf("%d", p))
		class := "page-item"
		if active {
			class += " active"
		}
		if disabled {
			class += " disabled"
			return fmt.Sprintf(`<li class="%s"><span class="page-link">%s</span></li>`, class, html.EscapeString(label))
		}
		return fmt.Sprintf(`<li class="%s"><a class="page-link" href="%s">%s</a></li>`, class, html.EscapeString(href), html.EscapeString(label))
	}

	var b strings.Builder
	b.WriteString(`<ul class="pagination justify-content-center flex-wrap">`)
	if result.HasPrev {
		b.WriteString(link(result.PrevPage, "«", false, false))
	} else {
		b.WriteString(link(1, "«", false, true))
	}

	// Window of page numbers around current (similar to many BBS UIs).
	start := page - 2
	if start < 1 {
		start = 1
	}
	end := start + 4
	if end > pages {
		end = pages
		start = end - 4
		if start < 1 {
			start = 1
		}
	}
	if start > 1 {
		b.WriteString(link(1, "1", false, false))
		if start > 2 {
			b.WriteString(`<li class="page-item disabled"><span class="page-link">…</span></li>`)
		}
	}
	for i := start; i <= end; i++ {
		b.WriteString(link(i, fmt.Sprintf("%d", i), i == page, false))
	}
	if end < pages {
		if end < pages-1 {
			b.WriteString(`<li class="page-item disabled"><span class="page-link">…</span></li>`)
		}
		b.WriteString(link(pages, fmt.Sprintf("%d", pages), false, false))
	}
	if result.HasNext {
		b.WriteString(link(result.NextPage, "»", false, false))
	} else {
		b.WriteString(link(pages, "»", false, true))
	}
	b.WriteString(`</ul>`)
	result.HTML = b.String()
	return result
}

func clampPage(page, pages int) int {
	if page < 1 {
		return 1
	}
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		return pages
	}
	return page
}
