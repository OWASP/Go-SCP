// Command filetype detects a file's type from its content and accepts only
// allowed image types.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

// sniffLen is how much content http.DetectContentType considers; see
// https://pkg.go.dev/net/http#DetectContentType.
const sniffLen = 512

func main() {
	err := run(os.Args, os.Stdout, os.Stderr)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// run prints the detected type of the uploaded file, or returns an error if
// it is not an allowed type.
func run(args []string, stdout, stderr io.Writer) (err error) {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	dir := flags.String("dir", "uploads", "directory holding uploaded files")
	name := flags.String("file", "img.png", "uploaded file to check, inside -dir")
	if err = flags.Parse(args[1:]); err != nil {
		return err
	}

	// Only open files inside the upload directory, even if the name
	// contains "..".
	root, err := os.OpenRoot(*dir)
	if err != nil {
		return fmt.Errorf("open upload directory: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()

	filetype, err := detectType(root, *name)
	if err != nil {
		return err
	}

	// Validate the detected type against an allowlist.
	switch filetype {
	case "image/jpeg", "image/gif", "image/png":
		_, _ = fmt.Fprintln(stdout, filetype)
		return nil
	default:
		return fmt.Errorf("file type not allowed: %s", filetype)
	}
}

// detectType returns the MIME type of the file name in root, based on its
// first bytes rather than its name.
func detectType(root *os.Root, name string) (string, error) {
	file, err := root.Open(name)
	if err != nil {
		return "", err
	}

	// Write the start of our file to a buffer. A file shorter than
	// sniffLen is fine: use only the bytes that were read.
	buff := make([]byte, sniffLen)
	n, readErr := io.ReadFull(file, buff)
	if errors.Is(readErr, io.ErrUnexpectedEOF) {
		readErr = nil
	}
	if err = errors.Join(readErr, file.Close()); err != nil {
		return "", fmt.Errorf("read %s: %w", name, err)
	}
	return http.DetectContentType(buff[:n]), nil
}
