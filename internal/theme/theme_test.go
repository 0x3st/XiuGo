package theme

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitAndStylesheets(t *testing.T) {
	root := t.TempDir()
	// create hifini-like
	dir := filepath.Join(root, "hifini")
	_ = os.MkdirAll(filepath.Join(dir, "css"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "theme.json"), []byte(`{"name":"H","css":["css/theme.css"]}`), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "css", "theme.css"), []byte("body{}"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "default"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "default", "theme.json"), []byte(`{"name":"D"}`), 0o644)
	if err := Init(context.Background(), root, func(context.Context) (string, error) { return "hifini", nil }); err != nil {
		t.Fatal(err)
	}
	if Global().ActiveID() != "hifini" {
		t.Fatalf("active=%s", Global().ActiveID())
	}
	css := Global().Stylesheets()
	if len(css) != 1 || css[0] != "/themes/hifini/css/theme.css" {
		t.Fatalf("css=%v", css)
	}
}
