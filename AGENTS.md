# Repository Guidelines

## Project Structure & Module Organization

This Go 1.25+ project provides a Zhihu CLI and full-screen reading TUI. `cmd/zhihu/main.go` is the executable entry point. Internal packages separate responsibilities:

- `internal/cli/`: command parsing, execution, and notification monitoring.
- `internal/client/`: HTTP API calls, comments, and read-request retries.
- `internal/auth/` and `internal/config/`: login, cookies, endpoints, and local configuration.
- `internal/display/`: terminal text output.
- `internal/feedtui/`: feed navigation, rendering, layout, comments, and terminal input.

Tests live beside implementation files as `*_test.go`. Platform-specific terminal code uses `_darwin.go`, `_linux.go`, and `_other.go` suffixes. There is no separate asset directory.

## Build, Test, and Development Commands

Run commands from the repository root:

- `go build -o zhihu ./cmd/zhihu`: build the local executable.
- `go run ./cmd/zhihu --help`: inspect commands without a separate build.
- `./zhihu feed --tui`: open the TUI after login.
- `go test ./...`: run all tests.
- `go test ./internal/feedtui -run TestName`: run a focused test.
- `golangci-lint run ./...`: run the documented lint checks.
- `gofmt -w path/to/changed.go`: format edited Go files before submitting.

After code changes pass verification, run `go install ./cmd/zhihu` from this fork's repository to update the locally installed `zhihu` before handing the work back. Verify the installed command with `zhihu --help` or the changed command's help. Use the current fork's source for installation. Documentation-only changes do not require reinstalling.

## Coding Style & Naming Conventions

Use standard Go formatting with tabs, lowercase package names, exported `PascalCase` identifiers, and unexported `camelCase` identifiers. Follow existing descriptive snake_case filenames.

Parse external input into trusted structures at system boundaries; internal functions should rely on established constraints. Propagate errors unless recovery is concrete, and expose programming errors rather than hiding them with defaults. Prefer direct implementations and introduce abstractions only for demonstrated reuse or a substantial, clearly named boundary.

## Testing Guidelines

Use Go's standard `testing` package and descriptive `TestBehavior` names. API tests use `net/http/httptest` with local handlers. Cover changed behavior and regressions through the actual implementation. For rendering changes, exercise narrow widths, Unicode text, and relevant navigation states. No numerical coverage threshold is configured; run the full suite before submitting.

## Commit & Pull Request Guidelines

Recent history predominantly uses `fix(scope): description` and `feat(scope): description`, with scopes such as `feed-tui`, `client`, and `notifications`. Follow that pattern and keep commits focused.

PRs should explain the problem, resulting behavior, and verification commands. Link relevant issues and include terminal screenshots for visible TUI changes.

## Security & Configuration

Cookies reside in `~/.zhihu-cli/cookies.json`; `ZHIHU_CLI_HOME` overrides the configuration directory. Keep credentials out of commits, fixtures, and logs, and preserve cookie-file permissions of `0600`.
