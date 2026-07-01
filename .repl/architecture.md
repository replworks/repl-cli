# architecture.md

# REPL CLI System Architecture

> Internal architecture of the REPL execution system.

Status: Active

---

# 1. System Purpose

REPL CLI is a local execution system that acts strictly as a **Runtime Controller**.  
It manages, inspects, and resets the lifecycle state of the local runtime environment.

It does NOT execute AI, does NOT coordinate multi-step workflows, and does NOT manage external integrations directly.  
It is a deterministic state controller for external AI execution tracking.

It is responsible for deterministic state transitions only.

---

# 2. System Overview

```text
User (CLI Commands)
  ↓
REPL CLI (Cobra Layer)
  ↓
Runtime Manager (includes validation)
  ↓
.repl/runtime/* (Local State Assets)
```

---

# 3. Core Responsibilities

## REPL CLI

Responsible for:

- Parsing CLI commands
- Delegating execution to runtime manager
- Orchestrating runtime lifecycle operations

Not responsible for:

- AI execution
- Workflow orchestration
- External integrations
- Business logic

---

## Runtime Manager

Responsible for:

- Managing runtime state
- Executing deterministic state transitions
- Applying validated runtime updates
- Performing validation of AI-provided execution results

Includes:

- Validation step (strict schema enforcement)

Not responsible for:

- AI execution
- Prompt generation
- External system communication

---

# 4. Runtime Model

All runtime state is stored in:

```text
.repl/runtime/
```

## Runtime Assets

- execution-state.json
- task-progress.json
- execution-log.json

---

## Runtime Rules

- Runtime Manager is the only component allowed to modify runtime state
- CLI cannot directly mutate runtime state
- AI cannot access or modify runtime state
- All state transitions must be deterministic

---

# 5. Execution Model

## Execution Flow

```text
User Command
  ↓
REPL CLI
  ↓
Runtime Manager
  ↓
Validation Step
  ↓
State Transition
  ↓
.repl/runtime/*
```

---

# 6. Determinism Model

The system must guarantee:

- Identical inputs produce identical state transitions
- Runtime behavior is fully reproducible
- No hidden or implicit state mutations exist

---

# 7. System Constraints

- No AI execution inside CLI
- No workflow engine abstraction
- No external integrations handled by core system
- No background daemon processes
- No distributed runtime system

---

# 8. Error Handling Model

- All errors must be explicit
- All invalid state transitions must be rejected
- Validation failures result in blocked execution state
- No silent recovery is allowed

---

# 9. Isolation Rules

- CLI is orchestration only
- Runtime Manager is state authority
- Runtime files are the only source of truth
- AI is fully external and stateless
- Validation is mandatory before state mutation

---

# 10. Architectural Invariants

The system must always ensure:

- Runtime is the single source of truth
- State transitions are deterministic
- AI is external and stateless
- CLI does not own state
- No implicit behavior exists
- All mutations are explicitly validated
