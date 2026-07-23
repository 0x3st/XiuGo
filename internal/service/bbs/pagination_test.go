package bbs

import "testing"

func TestBuildPaginationEmptyWhenOnePage(t *testing.T) {
	p := buildPagination("/?page={page}", 10, 1, 20)
	if p.HTML != "" {
		t.Fatalf("expected empty html, got %q", p.HTML)
	}
	if p.Pages != 1 {
		t.Fatalf("pages=%d", p.Pages)
	}
}

func TestBuildPaginationMultiplePages(t *testing.T) {
	p := buildPagination("/forum/1?page={page}", 45, 2, 20)
	if p.Pages != 3 {
		t.Fatalf("pages=%d want 3", p.Pages)
	}
	if p.HTML == "" || !containsAll(p.HTML, []string{"page-item", "page=2", "page=1", "page=3"}) {
		t.Fatalf("unexpected html: %s", p.HTML)
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
