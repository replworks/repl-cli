# TECH_STACK.md

## 1. Languages / runtimes / frameworks / libraries

- Go >= 1.26. go.mod MUST declare the Go version required by the dependencies.
- Module path MUST be `github.com/replworks/repl-cli`.
- Dependency manager: Go modules. go.mod and go.sum MUST be committed.
- github.com/spf13/cobra v1.x. MUST implement every command.
- The canonical AGENTS.md document and helper script MUST be retrieved as unmodified response text over HTTPS with the standard-library `net/http` client, from the URLs defined in PRODUCT_SPEC.md.
- github.com/aymanbagabas/go-udiff v0.x. MUST produce line-by-line diffs.
- Go standard library MUST be used for: net/http, log/slog, os/exec, bufio, regexp, testing, net/http/httptest.
- `git` executable, invoked through os/exec. MUST read TASKS.md from the base branch.
- Dev tools: gofmt, go vet, golangci-lint v2.x, govulncheck, GoReleaser v2.x.
- CI system: GitHub Actions.
- License: MIT.
- MUST NOT add any third-party module other than cobra, golang.org/x/net/html, and go-udiff.
- A new dependency MUST NOT be added unless the standard library is insufficient, the dependency is necessary, and it is actively maintained.

## 2. Framework- or tool-mandated paths and file names only

- `go.mod`, `go.sum`: repository root.
- `cmd/repl/`: main package.
- `internal/`: all non-entry-point packages.
- `*_test.go`: next to the file under test.
- `.golangci.yml`: repository root.
- `.goreleaser.yml`: repository root.
- `.github/workflows/`: GitHub Actions workflows.
- `LICENSE`: repository root.

## 3. Coding conventions specific to the chosen stack

- Code MUST be formatted with gofmt.
- Code MUST pass go vet and golangci-lint with no findings.
- Package names MUST NOT be `utils`, `helpers`, `common`, `shared`, or `misc`.
- Functions MUST return errors explicitly.
- Errors MUST be wrapped with `fmt.Errorf` and `%w`.
- Error messages MUST be actionable.
- MUST NOT silently recover from failures.
- MUST NOT use panic for control flow.
- Command output MUST be written through `cmd.OutOrStdout()` and `cmd.ErrOrStderr()`.
- Terminal color output MUST use ANSI escape sequences from the Go standard library only, and MUST be disabled for non-terminal writers.
- log/slog MUST be used for structured diagnostics.
- MUST NOT log secrets, tokens, or credentials.
- Cobra default `help` command, `--help` flag, and `completion` command MUST NOT be disabled.
- MUST NOT add speculative abstractions.

## 4. Testing rules

- Tests MUST use the Go standard `testing` package only.
- MUST NOT use assertion libraries, mock generators, or mocking frameworks.
- Test files MUST be named `*_test.go`. Test functions MUST be named `TestXxx`.
- Unit tests MUST cover business logic.
- Tests MUST be deterministic and isolated.
- The default suite (`go test ./...`) MUST NOT require network access, external services, or manual interaction.
- In the default suite, repl.net MUST be replaced by `httptest.Server`.
- In the default suite, external dependencies MUST be replaced by hand-written fakes.
- Filesystem and git repositories used by tests MUST be created in `t.TempDir()`.
- Live E2E tests MUST carry the build tag `e2e`.
- Live E2E tests MUST run only through `go test -tags e2e ./...`.
- Live E2E tests MAY require network access.
- Coverage is not used. MUST NOT define a coverage threshold.

## 5. Configuration rules

- None

## 6. Build, packaging, and release rules

- Build command: `go build ./cmd/repl`.
- Builds MUST use `CGO_ENABLED=0`.
- Binary name MUST be `repl`.
- Verification commands, all MUST pass:
  - `gofmt -l .` (MUST print nothing)
  - `go vet ./...`
  - `golangci-lint run`
  - `go test ./...`
- GitHub Actions MUST run the verification commands on ubuntu and macOS.
- GitHub Actions MUST read the Go version from go.mod.
- GitHub Actions MUST install the newest patch release of the major.minor version read from go.mod.
- GitHub Actions MUST run `govulncheck ./...`.
- Releases MUST be created from git tags `vMAJOR.MINOR.PATCH`.
- GoReleaser MUST build binaries for darwin and linux, each on amd64 and arm64.
- GoReleaser MUST publish archives and a checksum file to GitHub Releases.
- GoReleaser MUST generate `.deb` packages for linux.
- GoReleaser MUST generate the Homebrew cask.
- Distribution channels:
  - GitHub Releases: darwin and linux binaries.
  - APT repository `apt.repl.net` (GitHub Pages of `replworks/apt`): linux `.deb` packages.
  - Homebrew tap `replworks/homebrew-tap`: cask.
- Installation MUST NOT require source compilation.
- `LICENSE` MUST contain the MIT license.

## 7. Security rules tied to the chosen stack

- HTTP requests MUST use HTTPS only.
- MUST NOT disable TLS verification (`InsecureSkipVerify`).
- `http.Client` MUST set a timeout.
- Content fetched from repl.net MUST be treated as untrusted data.
- `git` MUST be invoked through os/exec with an explicit argument list.
- MUST NOT invoke `git` through a shell.
- User-supplied values (including the `--base` value) MUST NOT be interpreted as git options.
- MUST NOT hardcode secrets.
- MUST NOT commit secrets.
- `govulncheck` MUST report no known vulnerabilities.

## 8. Non-goals (technologies and tools explicitly excluded)

- Languages: MUST NOT use Node.js, TypeScript, Python, PHP, Java, or Rust.
- Architecture: MUST NOT add a backend service, database, web server, or daemon process.
- Libraries: MUST NOT use go-git, goldmark, testify, gomock, or a YAML/config library.
- Platform: MUST NOT use cgo.
- Auth and config: MUST NOT implement GitHub App device flow, token storage, or a configuration file.
- Distribution: MUST NOT use distribution channels other than GitHub Releases, APT, and Homebrew.
- Tooling: MUST NOT use a coverage tool.
