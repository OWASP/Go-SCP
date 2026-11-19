# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the OWASP Go Secure Coding Practices Guide - a GitBook-based documentation project providing security guidance for Go web application developers. The book follows the OWASP Secure Coding Practices Quick Reference Guide.

## Build Commands

```bash
# Build PDF, ePub and DOCX into dist/ with pandoc in Docker
./build-books.sh

# Build only some formats, or into another directory inside the repository
FORMATS=epub OUT_DIR=_book/test ./build-books.sh

# Check the Go examples (one module rooted at the repository root)
go vet ./... && go test ./...

# Legacy GitBook toolchain: live preview, and Mobi
npm i && node_modules/.bin/gitbook install && npm run serve
```

The pandoc build fails if a link between chapters cannot be resolved
(`fix-internal-links.lua`). Code blocks inside a Markdown blockquote break the
legacy GitBook build if they contain `{{ }}`.

## Project Structure

- `src/` - All book content in Markdown format
- `src/SUMMARY.md` - Book table of contents (defines chapter order)
- `src/README.md` - Book introduction
- `dist/` - Generated output files (PDF, ePub, Mobi, DOCX)
- `build-books.sh`, `Dockerfile.pandoc`, `book-metadata.yaml`, `book-cover.tex`,
  `fix-internal-links.lua`, `wrap-wide-tables.lua` - pandoc build
- `book.json`, `package.json`, `Dockerfile` - legacy GitBook build
- `go.mod` and `.go` files in `src/` subdirectories - runnable examples
  referenced by chapters, each with tests; keep the code quoted in a chapter in
  sync with its example file
- `.github/workflows/build.yml` - CI: Go checks and book build

## Content Organization

Book chapters follow OWASP security categories:
- Input Validation, Output Encoding
- Authentication/Password Management, Session Management
- Access Control, Cryptographic Practices
- Error Handling/Logging, Data Protection
- Communication Security, System Configuration
- Database Security, File Management, Memory Management
- General Coding Practices (CSRF, Regex)

## Git Flow

This project uses git-flow branching model:
- Features: `git flow feature start <name>` → PR to `develop`
- Hotfixes: `git flow hotfix start` → applies to both `develop` and `master`
- Releases are handled by project owner from `develop` to `master`
