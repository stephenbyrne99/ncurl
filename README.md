# ncurl – curl in English ⚡️

[![CI](https://github.com/stephenbyrne99/ncurl/actions/workflows/ci.yml/badge.svg)](https://github.com/stephenbyrne99/ncurl/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stephenbyrne99/ncurl.svg)](https://pkg.go.dev/github.com/stephenbyrne99/ncurl)

`ncurl` lets you describe an HTTP request in plain language. It asks Anthropic’s Claude to translate the description into a fully-specified request, executes it, and prints a JSON summary with status, headers and body.

---

## 💾 Installation

```bash
go install github.com/stephenbyrne99/ncurl/cmd/ncurl@latest
```

> **Prerequisite:** export `ANTHROPIC_API_KEY` in your shell before running `ncurl`.

I set this in my .zshrc so its always set :)

---

## 🚀 Quick Start

```bash
# Simple GET
ncurl "download https://httpbin.org/get"

# POST with JSON and a shorter timeout
ncurl -t 10 "POST to httpbin with a name field beind hello"

# Pipe prettified JSON through jq
ncurl "get goland github | jq '.body | fromjson | .stargazers_count'
```

---

## 🗂️ Project layout

```
ncurl/
├── cmd/
│   └── ncurl/          # CLI entry-point
├── internal/
│   ├── httpx/          # Request struct + executor
│   └── llm/            # Anthropic wrapper
├── go.mod
└── README.md
```

`httpx` and `llm` live in `internal/` so that they remain implementation details—only the top-level command is public.

---

## 🏗️ Continuous Integration

A simple **GitHub Actions** workflow (`.github/workflows/ci.yml`) runs on every push:

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Validate formatting
        run: |
          go vet ./...
          gofmt -s -d $(git ls-files '*.go') | tee /dev/stderr | (! read)
      - name: Run tests (none yet, placeholder)
        run: go test ./...
      - name: Build CLI
        run: go build ./cmd/ncurl
```

Add tests under `internal/...` as the project grows—GitHub will run them automatically.

---

## 📦 Releases

When you’re ready to ship binaries, drop a **GoReleaser** config and add this job to the workflow:

```yaml
  release:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main' && startsWith(github.event.head_commit.message, 'release:')
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

Each tag that starts with `release:` will build cross-platform binaries, attach them to the GitHub Release page, and update the `go install` path.

---

## 🤝 Contributing

1. Fork & clone, then run `go vet ./...` before opening a PR.
2. Keep commits small and descriptive.
3. All checks must pass before merge.

---

## 📝 License

MIT © 2025 Stephen Byrne
