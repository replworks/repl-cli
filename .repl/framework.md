# framework.md

# Go CLI Framework

## Purpose

This document defines implementation constraints.

product.md defines product requirements.  
architecture.md defines system structure.  
framework.md defines implementation rules.

All implementation decisions must follow this document.

Do not replace technologies.  
Do not introduce alternative stacks.  
Do not redesign implementation conventions.

---

# Language

Use:

```text
Go >= 1.24
```

Do not use:

```text
Node.js
TypeScript
Python
PHP
Java
Rust
```

---

# Architecture Style

Use:

```text
Single Binary CLI
```

Requirements:

- standalone executable
- no backend service
- no database
- no web server
- no daemon process

---

# Command Framework

Use:

```text
github.com/spf13/cobra
```

All commands must be implemented through Cobra.

---

# Configuration

Configuration format:

```text
YAML
```

Location:

```text
~/.config/<application>/config.yaml
```

Configuration must be optional.  
Applications must start with sensible defaults.

---

# Secrets

Use environment variables for secrets.

Never:

- hardcode secrets
- commit secrets
- store secrets in repository files

GitHub App device flow tokens must be stored locally with restricted file permissions.

---

# External Integrations

External services must be isolated behind dedicated adapters.

Business logic must not depend directly on external SDKs or APIs.

Implementation must allow replacement of integrations without changing business logic.

---

# Directory Structure

Use:

```text
cmd/
    <application>/

internal/
```

Structure internal packages around responsibilities.

Avoid generic package names:

- utils
- helpers
- common
- shared
- misc

---

# Error Handling

Return explicit errors.

Do not silently recover from failures.

Error messages must be actionable.

---

# Logging

Use structured logging.

Do not log secrets, tokens, or credentials.

---

# Testing

Use Go standard testing package.

Requirements:

- unit tests for business logic
- deterministic tests
- isolated tests

Tests must not require:

- network access
- external services
- manual interaction

External dependencies must be mocked.

---

# Dependency Management

Minimize dependencies.

Before introducing a dependency:

1. Verify standard library is insufficient.
2. Verify necessity.
3. Verify active maintenance.

---

# Distribution

Distribution method:

```text
GitHub Releases
```

Produce binaries for:

- macOS
- Linux
- Windows

No source compilation required for installation.

---

# Coding Principles

Prefer:

- Simple
- Explicit
- Readable
- Deterministic

Avoid:

- Premature optimization
- Speculative abstraction
- Framework-like complexity
- Hidden behavior

---

# Framework Invariants

- Single Binary CLI
- Configuration is externalized
- Secrets are not stored in source code
- Business logic is isolated from external integrations
- Tests do not require external services
- Implementation favors simplicity over abstraction
