# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the OWASP Go Secure Coding Practices Guide - a GitBook-based documentation project providing security guidance for Go web application developers. The book follows the OWASP Secure Coding Practices Quick Reference Guide.

## Build Commands

```bash
# Install dependencies and build all formats (PDF, ePub, Mobi, DOCX)
npm i && node_modules/.bin/gitbook install && npm run build

# Live preview with hot reload
npm run serve

# Build individual formats
npm run build:pdf
npm run build:epub
npm run build:mobi
npm run build:docx

# Build using Docker (alternative)
docker-compose run -u node:node --rm build
```

## Project Structure

- `src/` - All book content in Markdown format
- `src/SUMMARY.md` - Book table of contents (defines chapter order)
- `src/README.md` - Book introduction
- `dist/` - Generated output files (PDF, ePub, Mobi)
- `book.json` - GitBook configuration
- `.go` files in `src/` subdirectories - Example code snippets referenced by chapters

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
