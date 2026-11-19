// Command log-integrity checks that a log file still matches the SHA-256
// checksum recorded for it.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	err := run(os.Args, os.Stdout, os.Stderr)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// run compares the log's current hash with the stored checksum and reports
// tampering as an error.
func run(args []string, stdout, stderr io.Writer) (err error) {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	dir := flags.String("dir", "log", "directory holding the log and checksum files")
	logFile := flags.String("log", "log", "log file to check, inside -dir")
	checksumFile := flags.String("checksum", "checksum", "expected SHA-256 in hex, inside -dir")
	if err = flags.Parse(args[1:]); err != nil {
		return err
	}

	// Only read files inside the log directory, even if a name contains "..".
	root, err := os.OpenRoot(*dir)
	if err != nil {
		return fmt.Errorf("open log directory: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()

	// Get our known log checksum from the checksum file.
	stored, err := root.ReadFile(*checksumFile)
	if err != nil {
		return fmt.Errorf("read checksum: %w", err)
	}

	// Compute our current log's SHA-256 hash.
	b, err := ComputeSHA256(root, *logFile)
	if err != nil {
		return fmt.Errorf("hash log: %w", err)
	}

	// Compare our calculated hash with our stored hash.
	if hex.EncodeToString(b) != strings.TrimSpace(string(stored)) {
		return errors.New("file tampering detected")
	}
	_, _ = fmt.Fprintln(stdout, "Log integrity OK.")
	return nil
}

// ComputeSHA256 returns the SHA-256 hash of the file name in root.
func ComputeSHA256(root *os.Root, name string) ([]byte, error) {
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	_, copyErr := io.Copy(hash, file)
	if err = errors.Join(copyErr, file.Close()); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
