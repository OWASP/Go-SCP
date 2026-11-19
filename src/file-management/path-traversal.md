Path Traversal Prevention
=========================

Path traversal (directory traversal) attacks occur when user input is used to
construct file paths, allowing attackers to access files outside intended
directories. This is a critical vulnerability that can expose sensitive system
files.

The Vulnerability
-----------------

```go
// VULNERABLE - User can access any file on the system
func downloadHandler(w http.ResponseWriter, r *http.Request) {
    filename := r.URL.Query().Get("file")
    // Attacker requests: ?file=../../../etc/passwd
    path := filepath.Join("/uploads", filename)
    http.ServeFile(w, r, path) // Serves /etc/passwd!
}
```

The attack works because `filepath.Join` doesn't prevent `..` sequences from
escaping the base directory.

Go 1.24+: The os.Root Solution
------------------------------

Go 1.24 introduces `os.Root`, which provides traversal-resistant file operations:

```go
package main

import (
    "io"
    "net/http"
    "os"
)

func secureDownloadHandler(w http.ResponseWriter, r *http.Request) {
    filename := r.URL.Query().Get("file")

    // Open a root directory - all operations are confined here
    root, err := os.OpenRoot("/uploads")
    if err != nil {
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    defer root.Close()

    // Open file within root - traversal attempts are blocked
    f, err := root.Open(filename)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    defer f.Close()

    io.Copy(w, f)
}
```

### os.Root Operations

```go
root, _ := os.OpenRoot("/base/directory")
defer root.Close()

// Read operations
f, err := root.Open("path/to/file")           // Open for reading
info, err := root.Stat("path/to/file")        // Get file info
info, err := root.Lstat("path/to/file")       // Stat without following symlinks

// Write operations
f, err := root.Create("newfile.txt")          // Create file
f, err := root.OpenFile("file", os.O_WRONLY, 0644)
err := root.Mkdir("newdir", 0755)             // Create directory

// Delete operations
err := root.Remove("file.txt")                // Remove file or empty dir

// Nested roots
subroot, err := root.OpenRoot("subdir")       // Create sub-root
```

### Quick One-Off Operations

For simple cases, use `os.OpenInRoot`:

```go
// Opens file only if it's under /uploads
f, err := os.OpenInRoot("/uploads", userProvidedFilename)
if err != nil {
    // File doesn't exist or traversal attempted
}
defer f.Close()
```

Pre-Go 1.24 Solutions
---------------------

### Option 1: google/safeopen Package

```bash
go get github.com/google/safeopen
```

```go
import "github.com/google/safeopen"

func downloadHandler(w http.ResponseWriter, r *http.Request) {
    filename := r.URL.Query().Get("file")

    f, err := safeopen.OpenBeneath("/uploads", filename)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    defer f.Close()

    io.Copy(w, f)
}
```

### Option 2: Manual Validation with filepath.IsLocal

Go 1.20+ provides `filepath.IsLocal` to check if a path is "local" (doesn't
escape via `..` or absolute paths):

```go
func downloadHandler(w http.ResponseWriter, r *http.Request) {
    filename := r.URL.Query().Get("file")

    // Check if path is local (no .., no absolute paths)
    if !filepath.IsLocal(filename) {
        http.Error(w, "Invalid filename", http.StatusBadRequest)
        return
    }

    path := filepath.Join("/uploads", filename)
    http.ServeFile(w, r, path)
}
```

**Warning**: `filepath.IsLocal` alone doesn't protect against symlink attacks.

### Option 3: Manual Validation (All Go Versions)

```go
import (
    "path"
    "path/filepath"
    "strings"
)

func sanitizePath(baseDir, userPath string) (string, error) {
    // Clean the path (resolve . and ..)
    // Use path.Clean for URL paths, filepath.Clean for filesystem paths
    cleaned := path.Clean("/" + userPath)

    // Ensure it doesn't escape root
    if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/../") {
        return "", errors.New("invalid path")
    }

    // Remove leading slash and join with base
    cleaned = strings.TrimPrefix(cleaned, "/")
    fullPath := filepath.Join(baseDir, cleaned)

    // Verify the result is still under baseDir
    if !strings.HasPrefix(fullPath, filepath.Clean(baseDir)+string(os.PathSeparator)) {
        return "", errors.New("path escapes base directory")
    }

    return fullPath, nil
}
```

Zip Slip Attack
---------------

The "Zip Slip" vulnerability occurs when extracting archives containing files
with path traversal sequences in their names:

```go
// VULNERABLE - Zip Slip attack
func extractZipInsecure(zipPath, destDir string) error {
    r, _ := zip.OpenReader(zipPath)
    defer r.Close()

    for _, f := range r.File {
        // Attacker's file: "../../../etc/cron.d/malicious"
        path := filepath.Join(destDir, f.Name) // DANGEROUS!
        // ... extract to path
    }
    return nil
}
```

### Secure Archive Extraction (Go 1.24+)

```go
func extractZipSecure(zipPath, destDir string) error {
    r, err := zip.OpenReader(zipPath)
    if err != nil {
        return err
    }
    defer r.Close()

    root, err := os.OpenRoot(destDir)
    if err != nil {
        return err
    }
    defer root.Close()

    for _, f := range r.File {
        if f.FileInfo().IsDir() {
            root.Mkdir(f.Name, f.Mode())
            continue
        }

        // Safe - os.Root prevents traversal
        outFile, err := root.Create(f.Name)
        if err != nil {
            return err
        }

        rc, err := f.Open()
        if err != nil {
            outFile.Close()
            return err
        }

        io.Copy(outFile, rc)
        outFile.Close()
        rc.Close()
    }

    return nil
}
```

### Secure Archive Extraction (Pre-Go 1.24)

```go
func extractZipSecure(zipPath, destDir string) error {
    r, _ := zip.OpenReader(zipPath)
    defer r.Close()

    destDir = filepath.Clean(destDir) + string(os.PathSeparator)

    for _, f := range r.File {
        // Validate each file path
        filePath := filepath.Join(destDir, f.Name)

        // Check for Zip Slip
        if !strings.HasPrefix(filePath, destDir) {
            return fmt.Errorf("illegal file path: %s", f.Name)
        }

        // Also check with filepath.IsLocal (Go 1.20+)
        if !filepath.IsLocal(f.Name) {
            return fmt.Errorf("illegal file path: %s", f.Name)
        }

        // ... proceed with extraction
    }
    return nil
}
```

Symlink Attacks
---------------

Even with path validation, symlinks can redirect file operations:

```go
// Attack scenario:
// 1. Attacker creates /uploads/link -> /etc
// 2. Attacker requests ?file=link/passwd
// 3. Server serves /etc/passwd

// PROTECTION: Use Lstat to check for symlinks
func isSymlink(path string) bool {
    info, err := os.Lstat(path) // Lstat doesn't follow symlinks
    if err != nil {
        return false
    }
    return info.Mode()&os.ModeSymlink != 0
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
    // ... path validation ...

    // Check for symlinks in the path
    if isSymlink(fullPath) {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }

    http.ServeFile(w, r, fullPath)
}
```

**Note**: `os.Root` (Go 1.24+) handles symlinks safely by default.

File Upload Security
--------------------

Combine path traversal prevention with upload security:

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // Limit upload size
    r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Upload failed", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Sanitize filename - don't trust the client
    filename := filepath.Base(header.Filename) // Remove path components
    if filename == "." || filename == ".." {
        http.Error(w, "Invalid filename", http.StatusBadRequest)
        return
    }

    // Generate safe filename (recommended)
    ext := filepath.Ext(filename)
    safeFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

    // Use os.Root for safe storage (Go 1.24+)
    root, _ := os.OpenRoot("/uploads")
    defer root.Close()

    dst, err := root.Create(safeFilename)
    if err != nil {
        http.Error(w, "Storage error", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    io.Copy(dst, file)
    w.WriteHeader(http.StatusCreated)
}
```

Best Practices Summary
----------------------

| Approach | Go Version | Symlink Safe | TOCTOU Safe |
|----------|------------|--------------|-------------|
| `os.Root` / `os.OpenInRoot` | 1.24+ | Yes | Yes |
| `google/safeopen` | Any | Yes | Yes |
| `filepath.IsLocal` + validation | 1.20+ | No | No |
| Manual validation | Any | No | No |

**Recommendations:**

1. **Use `os.Root`** (Go 1.24+) for all file operations with untrusted paths
2. **Use `google/safeopen`** for pre-1.24 codebases
3. **Never trust user input** for filenames - generate safe names when possible
4. **Map user input to IDs** instead of using filenames directly when feasible
5. **Validate file types** by content, not just extension

References
----------

- [Go os.Root Blog Post][1]
- [OWASP Path Traversal][2]
- [google/safeopen Package][3]

[1]: https://go.dev/blog/osroot
[2]: https://owasp.org/www-community/attacks/Path_Traversal
[3]: https://github.com/google/safeopen
