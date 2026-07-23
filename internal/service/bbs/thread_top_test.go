package bbs

import (
	"testing"

	"github.com/0x3st/XiuGo/internal/model/view"
)

func TestMergePinnedThreadsPrefersSiteThenForumAndDedupes(t *testing.T) {
	pinned := []view.ThreadSummary{
		{Tid: 30, Top: 3, Subject: "site"},
		{Tid: 10, Top: 1, Subject: "forum"},
	}
	regular := []view.ThreadSummary{
		{Tid: 10, Top: 1, Subject: "dup"},
		{Tid: 5, Top: 0, Subject: "normal"},
		{Tid: 4, Top: 0, Subject: "older"},
	}
	got := mergePinnedThreads(pinned, regular, 10)
	if len(got) != 4 {
		t.Fatalf("len=%d want 4: %+v", len(got), got)
	}
	if got[0].Tid != 30 || got[1].Tid != 10 || got[2].Tid != 5 || got[3].Tid != 4 {
		t.Fatalf("order wrong: %+v", got)
	}
}

func TestMergePinnedThreadsRespectsLimit(t *testing.T) {
	pinned := []view.ThreadSummary{{Tid: 3}, {Tid: 2}}
	regular := []view.ThreadSummary{{Tid: 1}}
	got := mergePinnedThreads(pinned, regular, 2)
	if len(got) != 2 || got[0].Tid != 3 || got[1].Tid != 2 {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestThreadTopClass(t *testing.T) {
	if threadTopClass(1) != "top_1" || threadTopClass(3) != "top_3" || threadTopClass(0) != "" {
		t.Fatal("top class mapping wrong")
	}
}
