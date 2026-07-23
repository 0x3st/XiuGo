package bbs

import (
	"strings"
	"testing"
	"time"
)

func TestXiunoTextToHTML(t *testing.T) {
	t.Parallel()
	input := "a b\t<c>\r\n\"q\"\n"
	want := "a&nbsp; b &nbsp; &nbsp; &nbsp; &nbsp;&lt;c&gt;<br>&quot;q&quot;<br>"
	if got := xiunoTextToHTML(input); got != want {
		t.Fatalf("xiunoTextToHTML() = %q, want %q", got, want)
	}
}

func TestFormatPostMessageHTMLPermission(t *testing.T) {
	t.Parallel()
	message := `<strong onclick="bad()">正文</strong>`
	if got := formatPostMessage(message, 0, 1); got != message {
		t.Fatalf("administrator HTML should be preserved: %q", got)
	}
	if got := formatPostMessage(message, 0, 101); strings.Contains(got, "<strong") {
		t.Fatalf("ordinary user HTML should be escaped: %q", got)
	}
}

func TestQuoteBrief(t *testing.T) {
	t.Parallel()
	if got, want := quoteBrief("<b>abc</b>&", 100), "abc&amp;"; got != want {
		t.Fatalf("quoteBrief() = %q, want %q", got, want)
	}
	long := strings.Repeat("好", 101)
	if got := quoteBrief(long, 100); got != strings.Repeat("好", 100)+" ... " {
		t.Fatalf("quoteBrief() did not apply Xiuno's 100-character suffix: %q", got)
	}
}

func TestHumanDateAt(t *testing.T) {
	t.Parallel()
	now := time.Unix(2_000_000_000, 0)
	cases := []struct {
		seconds int64
		want    string
	}{
		{30, "30秒前"},
		{120, "2分钟前"},
		{7_200, "2小时前"},
		{172_800, "2天前"},
	}
	for _, tc := range cases {
		got := humanDateAt(uint(now.Unix()-tc.seconds), now)
		if got != tc.want {
			t.Fatalf("humanDateAt(%d seconds) = %q, want %q", tc.seconds, got, tc.want)
		}
	}
}

func TestCanReplyToClosedThread(t *testing.T) {
	t.Parallel()
	if !canReplyToThread(1, 1) || !canReplyToThread(1, 5) {
		t.Fatal("Xiuno administrator groups 1-5 should be able to reply to a closed thread")
	}
	if canReplyToThread(1, 0) || canReplyToThread(1, 101) {
		t.Fatal("guests and ordinary groups should not be able to reply to a closed thread")
	}
	if !canReplyToThread(0, 101) {
		t.Fatal("an open thread should allow an otherwise-authorized ordinary group")
	}
}
