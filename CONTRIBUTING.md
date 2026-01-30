# Contributing

Thanks for helping improve Cooperations.

## Prerequisites

- Go 1.23+

## Setup

```
copy .env.example .env
go test ./...
```

## Development Guidelines

- Keep the project local-first (no managed cloud dependencies).
- Prefer simple, readable Go code over complex abstractions.
- Add tests for bug fixes and non-trivial features.
- Avoid committing secrets or local artifacts.

## Pull Requests

1. Create a focused change set.
2. Ensure `go test ./...` passes.
3. Describe the problem, approach, and any tradeoffs.

## License

By contributing, you agree that your contributions are licensed under the GNU Affero General Public License v3.0.
