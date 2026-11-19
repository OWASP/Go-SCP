package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAllowsOnlyImages(t *testing.T) {
	dir := t.TempDir()
	files := map[string][]byte{
		"img.png":    []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 16)),
		"img.gif":    []byte("GIF89a" + strings.Repeat("\x00", 16)),
		"script.png": []byte("<html><script>alert(1)</script></html>"),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o600); err != nil {
			t.Fatalf("writing %s: %s", name, err)
		}
	}
	// A file outside the upload directory that a "../" name could reach.
	if err := os.WriteFile(
		filepath.Join(filepath.Dir(dir), "outside.png"),
		files["img.png"],
		0o600,
	); err != nil {
		t.Fatalf("writing outside.png: %s", err)
	}

	cases := map[string]struct {
		file       string
		wantOutput string
		wantErr    string
	}{
		"png": {file: "img.png", wantOutput: "image/png\n"},
		"gif": {file: "img.gif", wantOutput: "image/gif\n"},
		"html named .png": {
			file:    "script.png",
			wantErr: "file type not allowed: text/html",
		},
		"name escaping upload dir": {file: "../outside.png", wantErr: "path escapes from parent"},
		"file that does not exist": {file: "missing.png", wantErr: "no such file"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var stdout bytes.Buffer
			args := []string{"filetype", "-dir", dir, "-file", tc.file}

			err := run(args, &stdout, &bytes.Buffer{})

			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
			if got := stdout.String(); got != tc.wantOutput {
				t.Fatalf("expected output %q, got %q", tc.wantOutput, got)
			}
		})
	}
}
