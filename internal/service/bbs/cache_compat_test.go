package bbs

import (
	"context"
	"testing"
	"time"
)

func TestPHPCacheKeysIncludeConfiguredPrefix(t *testing.T) {
	keys := New().phpCacheKeys(context.Background(), "runtime")
	want := map[string]bool{"runtime": true, "bbs_runtime": true}
	for _, key := range keys {
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("phpCacheKeys(runtime) = %#v; missing %#v", keys, want)
	}
}

func TestMaintenanceDueUsesOriginalStrictIntervals(t *testing.T) {
	const now = int64(100_000)
	if maintenanceDue(now-300, now, fiveMinuteInterval) {
		t.Fatal("exactly five minutes should not be due; PHP uses > 300")
	}
	if !maintenanceDue(now-301, now, fiveMinuteInterval) {
		t.Fatal("more than five minutes should be due")
	}
	if maintenanceDue(now-int64((24*time.Hour)/time.Second), now, dailyInterval) {
		t.Fatal("exactly one day should not be due; PHP uses > 86400")
	}
	if !maintenanceDue(0, now, dailyInterval) {
		t.Fatal("an uninitialized daily timestamp should be due")
	}
}
