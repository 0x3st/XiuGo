package bbs

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestXiunoPasswordUsesOriginalClientHashProtocol(t *testing.T) {
	clientHash := fmt.Sprintf("%x", md5.Sum([]byte("123.com")))
	want := fmt.Sprintf("%x", md5.Sum([]byte(clientHash+"test-salt")))
	got, err := xiunoPassword(clientHash, "test-salt")
	if err != nil {
		t.Fatalf("xiunoPassword returned error: %v", err)
	}
	if got != want {
		t.Fatalf("xiunoPassword = %q; want %q", got, want)
	}
}

func TestXiunoPasswordRejectsCleartext(t *testing.T) {
	if _, err := xiunoPassword("123.com", "test-salt"); err == nil {
		t.Fatal("xiunoPassword accepted a cleartext password")
	}
}
