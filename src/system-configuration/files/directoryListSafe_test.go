package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileServerHidesDirectoryContents(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.css"), []byte("body{}"), 0o600); err != nil {
		t.Fatalf("writing file: %s", err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, "secret.txt"),
		[]byte("password"),
		0o600,
	); err != nil {
		t.Fatalf("writing file: %s", err)
	}
	handler := http.FileServer(justFilesFilesystem{http.Dir(dir)})

	cases := map[string]struct {
		target      string
		wantBody    string
		notWantBody string
	}{
		"file is served":          {target: "/index.css", wantBody: "body{}"},
		"directory is not listed": {target: "/", notWantBody: "secret.txt"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(
				rec,
				httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.target, nil),
			)

			body, err := io.ReadAll(rec.Result().Body)
			if err != nil {
				t.Fatalf("reading body: %s", err)
			}
			if !strings.Contains(string(body), tc.wantBody) {
				t.Fatalf("expected body containing %q, got %q", tc.wantBody, body)
			}
			if tc.notWantBody != "" && strings.Contains(string(body), tc.notWantBody) {
				t.Fatalf("expected body without %q, got %q", tc.notWantBody, body)
			}
		})
	}
}
