# REPL CLI 🚀

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.24-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Build](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com)

**A deterministic runtime controller for external AI-driven task execution**

---

## What is REPL CLI? 🤔

REPL CLI is a **local command-line system** that manages and observes deterministic runtime state for external AI-driven task execution. It acts as a bridge between AI systems and local project state, providing:

- ✅ **Deterministic state management** — Same input always produces the same output
- ✅ **Task lifecycle tracking** — Initialize, execute, and validate AI tasks
- ✅ **Local-only architecture** — No backend, no cloud, no daemon
- ✅ **Single binary distribution** — No dependencies, no installation hassle

---

## Why REPL CLI? 💡

### The Problem

When working with AI agents (like Claude, GPT-4, or custom LLMs), you need a way to:

- Track task execution state reliably
- Validate AI outputs before applying them
- Maintain deterministic project state
- Avoid hidden state mutations

### The Solution

REPL CLI provides a **deterministic runtime controller** that:

1. Manages local runtime state in `.repl/runtime/`
2. Validates AI-generated execution results against strict schemas
3. Ensures reproducible, auditable state transitions
4. Keeps everything local — no external dependencies

---

## Features ⚡

### 🎯 Core Commands

| Command | Description |
|---------|-------------|
| `repl init` | Initialize REPL project runtime environment |
| `repl doctor` | Validate system integrity and configuration |
| `repl reset` | Reset runtime to clean initial state |
| `repl runtime start` | Start execution session for AI input |
| `repl runtime stop` | Stop execution session (no state mutation) |
| `repl runtime apply` | Apply AI execution results (via JSON stdin) |
| `repl runtime status` | Display current runtime state and progress |

### 🔒 Key Principles

- **Deterministic**: Identical inputs → identical outputs
- **Stateless AI**: AI is external and has no access to runtime state
- **Validated**: All state transitions are explicitly validated
- **Local**: No backend, no cloud sync, no daemon processes
- **Simple**: Single binary, zero configuration required

---

## Installation 📦

### Prerequisites

- Go >= 1.24

### Install from Source

```bash
git clone https://github.com/replworks/repl-cli.git
cd repl-cli
go build -o repl ./cmd/repl
sudo mv repl /usr/local/bin/
```

### Verify Installation

```bash
repl --version
# Output: repl version 0.1.0
```

---

## Quick Start 🚀

### 1. Initialize Project

```bash
repl init
```

**Output:**

```bash
REPL project initialized successfully
Created: .repl/
Created: .repl/runtime/
Initialized: execution-state.json
Initialized: task-progress.json
Initialized: execution-log.json
Copied: .repl/agent.md
Copied: AGENTS.md
```

> If `AGENTS.md` already exists in the project root, the template content (excluding the `# AGENTS.md` heading) is **prepended** to the existing file instead of overwriting it.

### 2. Start Runtime Session

```bash
repl runtime start
```

**Output:**

```bash
Starting REPL runtime execution session...
Runtime execution session activated
System is ready for AI execution input
```

### 3. Apply AI Execution Result

```bash
echo '{
  "action": "update_runtime",
  "taskId": "TASK_1",
  "status": "done",
  "events": ["step1", "step2", "step3"]
}' | repl runtime apply
```

**Output:**

```bash
Applying AI execution result...
Task TASK_1 marked as: done
Task completed successfully
```

### 4. Check Status

```bash
repl runtime status
```

**Output:**

```bash
REPL Runtime Status
===================

Session Status:
  State: active
  Current Task: none

Task Progress:
  TASK_1: done

Execution Log (last 5 entries):
  - Runtime session started
  - Runtime apply executed for task: TASK_1
  - Task status updated to: done
  - Events: [step1 step2 step3]

Runtime state is readable and consistent
```

### 5. Stop Session

```bash
repl runtime stop
```

---

## Architecture 🏗️

```bash
User Command
    ↓
REPL CLI (Cobra Layer)
    ↓
Runtime Manager (includes validation)
    ↓
.repl/runtime/* (Local State Assets)
```

### Runtime State Files

All state is stored locally under `.repl/runtime/`:

- **execution-state.json** — Session status and current task
- **task-progress.json** — Task completion status
- **execution-log.json** — Execution history and audit trail

### Execution Flow

```text
repl runtime start
    ↓
Load tasks.md
    ↓
Load .repl context
    ↓
Generate AI prompt
    ↓
Output to stdout
```

```text
AI Output (JSON)
    ↓
repl runtime apply (stdin)
    ↓
Validate schema
    ↓
Update .repl/runtime/*
    ↓
Mark TASK status (DONE / BLOCKED)
```

---

## JSON Schema for `repl runtime apply` 📋

The `repl runtime apply` command accepts JSON via stdin with the following schema:

```json
{
  "action": "update_runtime",
  "taskId": "TASK_1",
  "status": "done",
  "reason": "optional reason if blocked",
  "events": ["event1", "event2"]
}
```

### Required Fields

- `action` — Must be exactly `"update_runtime"`
- `taskId` — Target task identifier (e.g., `"TASK_1"`)
- `status` — Must be `"done"` or `"blocked"`

### Conditional Fields

- `reason` — Required only when `status` is `"blocked"`
- `events` — Optional array of execution milestones

---

## Use Cases 🎯

### 1. AI Agent Task Management

Use REPL CLI to manage tasks executed by AI agents:

```bash
# Initialize project
repl init

# Start session
repl runtime start

# Let AI execute tasks and apply results
cat > ai_result.json <<EOF
{
  "action": "update_runtime",
  "taskId": "TASK_1",
  "status": "done",
  "events": ["analyzed", "implemented", "tested"]
}
EOF

cat ai_result.json | repl runtime apply

# Check status
repl runtime status

# Stop session
repl runtime stop
```

### 2. Deterministic Build Pipelines

Ensure reproducible builds by tracking execution state:

```bash
repl init
repl runtime start

# Execute build steps
echo '{"action":"update_runtime","taskId":"BUILD","status":"done"}' | repl runtime apply

repl runtime status
repl runtime stop
```

### 3. Multi-Organization Workflows

Manage isolated runtime state per workspace:

```bash
# Project A
cd project-a && repl init && repl runtime start

# Project B (separate state)
cd project-b && repl init && repl runtime start
```

---

## Design Principles 🎨

### No Backend Service

REPL CLI is a **single binary** with no server component.

### No Daemon Process

All operations are **command-line driven** — no background processes.

### No Cloud Sync

All state is **local only** under `.repl/runtime/`.

### No AI Execution

REPL CLI **does not execute AI** — it only manages state for external AI systems.

### Deterministic Behavior

Same input → same output. No hidden state, no surprises.

---

## Development 🛠️

### Project Structure

```text
repl-cli/
├── cmd/
│   └── repl/
│       ├── main.go
│       ├── init.go
│       ├── doctor.go
│       ├── reset.go
│       └── runtime.go
├── internal/
│   └── runtime/
│       └── manager.go
├── templates/
│   ├── .repl/
│   │   └── agent.md       ← copied to .repl/agent.md on init
│   └── AGENTS.md          ← copied (or prepended) to AGENTS.md on init
├── .repl/
│   ├── agent.md
│   ├── product.md
│   ├── framework.md
│   ├── architecture.md
│   └── tasks.md
├── go.mod
└── README.md
```

### Build

```bash
go build ./cmd/repl
```

### Test

```bash
# Initialize project
./repl init

# Validate system
./repl doctor

# Test runtime commands
./repl runtime start
./repl runtime status
./repl runtime stop
./repl reset
```

---

## Contributing 🤝

Contributions are welcome! Please follow these guidelines:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Code Standards

- Follow Go best practices
- Write unit tests for business logic
- Ensure deterministic behavior
- No external service dependencies in tests

---

## Roadmap 🗺️

- [ ] GitHub App integration for multi-org support
- [ ] Configuration file support (`~/.config/repl/config.yaml`)
- [ ] Structured logging with log levels
- [ ] Plugin system for custom validators
- [ ] Shell completion (bash, zsh, fish)
- [ ] Man pages
- [ ] Homebrew tap for macOS
- [ ] apt repository for Linux
- [ ] Chocolatey package for Windows

---

## FAQ ❓

**Q: Is REPL CLI an AI execution engine?**

A: No. REPL CLI does not execute AI. It manages runtime state for external AI systems.

**Q: Where is my data stored?**

A: All data is stored locally under `.repl/runtime/` in your project directory. No cloud sync.

**Q: Can I use this with any AI system?**

A: Yes! REPL CLI is AI-agnostic. It accepts JSON input via stdin, so any AI system can use it.

**Q: Is this production-ready?**

A: REPL CLI is currently in MVP stage. It's suitable for development and testing. Production use is coming soon.

**Q: Why not just use a database?**

A: REPL CLI is designed for simplicity and determinism. JSON files are human-readable, version-controllable, and require no additional infrastructure.

---

## License 📄

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments 🙏

- Built with [Cobra](https://github.com/spf13/cobra) — A CLI framework for Go
- Inspired by the need for deterministic AI task management
- Designed for simplicity, explicitness, and reliability

---

## Star History ⭐

If you find REPL CLI useful, please consider giving it a star! It helps others discover the project.

[![Star History Chart](https://api.star-history.com/svg?repos=replworks/repl-cli&type=Date)](https://star-history.com/#replworks/repl-cli&Date)

---

## Contact 📧

Have questions? Found a bug? Want to contribute?

- **Issues**: [GitHub Issues](https://github.com/replworks/repl-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/replworks/repl-cli/discussions)
- **Email**: <your.email@example.com>

---

**Made with ❤️ for the AI engineering community**
