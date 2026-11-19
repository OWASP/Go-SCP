package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDetectsTampering(t *testing.T) {
	const logContent = "2026-11-19 user=alice action=login\n"
	const logName = "log"
	sum := sha256.Sum256([]byte(logContent))
	checksum := hex.EncodeToString(sum[:])

	cases := map[string]struct {
		log        string
		checksum   string
		logName    string
		wantOutput string
		wantErr    string
	}{
		"unchanged log": {
			log: logContent, checksum: checksum, logName: logName,
			wantOutput: "Log integrity OK.\n",
		},
		"checksum file ends with a newline": {
			log: logContent, checksum: checksum + "\n", logName: logName,
			wantOutput: "Log integrity OK.\n",
		},
		"edited log": {
			log: strings.Replace(
				logContent,
				"alice",
				"mallory",
				1,
			),
			checksum: checksum,
			logName:  "log",
			wantErr:  "file tampering detected",
		},
		"log name escaping the log directory": {
			log: logContent, checksum: checksum, logName: "../log",
			wantErr: "path escapes from parent",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, logName), []byte(tc.log), 0o600); err != nil {
				t.Fatalf("writing log: %s", err)
			}
			if err := os.WriteFile(
				filepath.Join(dir, "checksum"),
				[]byte(tc.checksum),
				0o600,
			); err != nil {
				t.Fatalf("writing checksum: %s", err)
			}
			var stdout bytes.Buffer
			args := []string{"log-integrity", "-dir", dir, "-log", tc.logName}

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
