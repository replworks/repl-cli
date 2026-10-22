# repl-cli

Command-line program for REPL Works projects.

`repl` installs the canonical AGENTS.md into a project directory, reports task
progress read from TASKS.md, verifies a project against a fixed set of rules,
and replaces a project's AGENTS.md with the canonical AGENTS.md.

## Install

```sh
brew install replworks/tap/repl-cli
```

Or download a prebuilt archive from
[GitHub Releases](https://github.com/replworks/repl-cli/releases) for darwin,
and linux (amd64, arm64). Linux users can also install the `.deb`
package from the [replworks APT repository](https://apt.repl.net).

## Usage

```sh
repl init      # create AGENTS.md and the .replworks documents
repl status    # report task progress from .replworks/TASKS.md
repl check     # verify the project against the project rules
repl sync      # replace AGENTS.md with the canonical AGENTS.md
```

## Build

```sh
make build      # go build -o repl ./cmd/repl
make check      # gofmt check, go vet, go test
```

## Development environment

Requirements:

- Go 1.26 or newer. `go.mod` declares the required version.
- `git` on `PATH`. `repl check` reads `.replworks/TASKS.md` from the base
  branch, and the test suite creates real repositories.
- Dev tools: `gofmt`, `go vet`, `golangci-lint` v2.x, `govulncheck`, and
  `goreleaser` v2.x for releases.

Commands:

```sh
make build          # go build -o repl ./cmd/repl
make test           # go test ./...
make vet            # go vet ./...
make lint           # golangci-lint run ./...
make fmt            # go fmt ./...
make check          # gofmt check, go vet, go test ./...
```

Run the local build from a scratch directory outside this repository, so a
command never changes the repository's own documents. `make build` writes the
`repl` binary to the repository root (it is excluded from version control):

```sh
make build
scratch_dir=$(mktemp -d)
cd "$scratch_dir"
/path/to/repl-cli/repl init
/path/to/repl-cli/repl sync
```

`status` and `check` read `.replworks/TASKS.md`, so create it in the scratch
directory when you exercise those commands.

The default test suite needs no network access, no external service, and no
manual interaction, but it does run the local `git` executable. Tests that
contact the canonical source and real repositories carry the `e2e` build tag
and run only on demand:

```sh
go test -tags e2e ./...   # requires network access to www.repl.net
```

Vulnerability scan:

```sh
go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
```

Release, from a `vMAJOR.MINOR.PATCH` tag:

```sh
make release-check   # goreleaser check
make snapshot        # goreleaser release --snapshot --clean
```

## License

MIT
