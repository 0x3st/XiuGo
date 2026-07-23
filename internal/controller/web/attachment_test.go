package web

import (
	"testing"

	"github.com/0x3st/XiuGo/internal/model/view"
)

func TestNextPendingAttachmentIDDoesNotReuseSparseIndex(t *testing.T) {
	items := []view.PendingAttachment{{ID: "_0"}, {ID: "_3"}, {ID: "invalid"}}
	if got := nextPendingAttachmentID(items); got != "_4" {
		t.Fatalf("nextPendingAttachmentID = %q; want _4", got)
	}
}
