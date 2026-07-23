package bbs

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeAttachmentPathPreservesOriginalSubdirectory(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "xiuno-upload")
	got, err := safeAttachmentPath(root, "202607/example.png")
	if err != nil {
		t.Fatalf("safeAttachmentPath returned error: %v", err)
	}
	want := filepath.Join(root, "attach", "202607", "example.png")
	if got != want {
		t.Fatalf("safeAttachmentPath = %q; want %q", got, want)
	}
}

func TestSafeAttachmentPathRejectsTraversal(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "xiuno-upload")
	for _, filename := range []string{"../conf/conf.php", "202607/../../conf.php", "/tmp/file"} {
		if _, err := safeAttachmentPath(root, filename); err == nil {
			t.Fatalf("safeAttachmentPath accepted unsafe filename %q", filename)
		}
	}
}

func TestSafePendingAttachmentPathRejectsTraversal(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "xiuno-upload")
	valid := filepath.Join(root, "tmp", "1_example.txt")
	if got, err := safePendingAttachmentPath(root, valid); err != nil || got != valid {
		t.Fatalf("safePendingAttachmentPath(%q) = %q, %v", valid, got, err)
	}
	for _, path := range []string{
		filepath.Join(root, "attach", "example.txt"),
		filepath.Join(root, "tmp"),
		filepath.Join(root, "tmp", "..", "conf.php"),
	} {
		if _, err := safePendingAttachmentPath(root, path); err == nil {
			t.Fatalf("safePendingAttachmentPath accepted unsafe path %q", path)
		}
	}
}

func TestDecodeAttachmentDataURI(t *testing.T) {
	want := []byte("Xiuno attachment")
	encoded := "data:text/plain;base64," + base64.StdEncoding.EncodeToString(want)
	got, err := decodeAttachmentData(encoded)
	if err != nil {
		t.Fatalf("decodeAttachmentData returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("decodeAttachmentData = %q; want %q", got, want)
	}
}

func TestOriginalAttachmentTypeCompatibility(t *testing.T) {
	tests := map[string]struct {
		extension string
		stored    string
		filetype  string
	}{
		"image":      {extension: "jpg", stored: "jpg", filetype: "image"},
		"office":     {extension: "docx", stored: "docx", filetype: "office"},
		"disallowed": {extension: "html", stored: "_html", filetype: "other"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			extension := originalAttachmentExtension("sample." + test.extension)
			stored := extension
			if !originalAttachmentAllowed[extension] {
				stored = "_" + extension
			}
			if stored != test.stored || originalAttachmentType(extension) != test.filetype {
				t.Fatalf("extension=%q stored=%q type=%q", extension, stored, originalAttachmentType(extension))
			}
		})
	}
}

func TestMoveAttachmentFile(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "source.txt")
	destination := filepath.Join(directory, "destination.txt")
	if err := os.WriteFile(source, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := moveAttachmentFile(source, destination); err != nil {
		t.Fatalf("moveAttachmentFile returned error: %v", err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
	if data, err := os.ReadFile(destination); err != nil || string(data) != "payload" {
		t.Fatalf("destination=%q, %v", data, err)
	}
}
