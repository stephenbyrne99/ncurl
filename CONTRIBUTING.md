# Contributing to ncurl

We love your input! We want to make contributing to ncurl as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

### Pull Requests

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

### Testing

Before submitting your PR, make sure all tests pass:

```bash
go test ./...
```

Also, run the linter to ensure your code follows our coding standards:

```bash
go vet ./...
gofmt -s -d $(git ls-files '*.go')
```

## Any contributions you make will be under the MIT Software License

When you submit code changes, your submissions are understood to be under the same [MIT License](http://choosealicense.com/licenses/mit/) that covers the project. Feel free to contact the maintainers if that's a concern.

## Report bugs using GitHub Issues

We use GitHub issues to track public bugs. Report a bug by [opening a new issue](https://github.com/stephenbyrne99/ncurl/issues/new); it's that easy!

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License.