# repl-cli

Command-line program for REPL Works projects.

`repl` installs the canonical AGENTS.md into a project directory, reports task
progress read from TASKS.md, verifies a project against a fixed set of rules,
and replaces a project's AGENTS.md with the canonical AGENTS.md.

## Install

```sh
brew install replworks/tap/repl
```

Or download a prebuilt archive from
[GitHub Releases](https://github.com/replworks/repl-cli/releases) for darwin,
linux, and windows (amd64, arm64). Linux users can also install the `.deb`
package from the [replworks APT repository](https://apt.repl.net).

## Usage

```sh
repl init      # create AGENTS.md and .replworks/
repl status    # report task progress from .replworks/TASKS.md
repl check     # verify the project against the project rules
repl sync      # replace AGENTS.md with the canonical AGENTS.md
```

## Build

```sh
make build      # go build -o repl ./cmd/repl
make check      # gofmt check, go vet, go test
```

## License

MIT
