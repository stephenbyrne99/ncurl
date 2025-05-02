# ncurl – curl in English ⚡️

`ncurl` lets you describe an HTTP request in plain language. It asks Anthropic's Claude to translate the description into a fully-specified request, executes it, and prints a JSON summary with status, headers and body.

## Installation

```bash
npm install -g ncurl
```

> **Prerequisite:** export `ANTHROPIC_API_KEY` in your shell before running `ncurl`.

## Usage

```bash
# Simple GET
ncurl "download https://httpbin.org/get"

# POST with JSON and a shorter timeout
ncurl -t 10 "POST to httpbin with a name field beind hello"

# Pipe prettified JSON through jq
ncurl "get goland github | jq '.body | fromjson | .stargazers_count'
```

## License

MIT © 2025 Stephen Byrne
