# product.md

# REPL CLI (Runtime Control System)

## 1. Product Overview

REPL CLI is a local command-line system that manages and observes a deterministic runtime state for external AI-driven task execution.

It does not execute AI.

It does not store long-term project state.

It acts only as a runtime controller for task lifecycle tracking and validation.

---

## 2. Core Purpose

REPL CLI provides:

- Task lifecycle initialization and management
- Runtime state inspection
- Runtime reset and recovery
- Controlled application of external AI-generated execution results

All state is local and stored under `.repl/runtime/*`.

---

## 3. Core Commands (MVP Scope)

### Project Commands

- `repl init`
  - Initialize a REPL project in current directory

- `repl doctor`
  - Validate runtime environment and configuration integrity

- `repl reset`
  - Reset local runtime state to initial clean state

---

### Execution Commands

- `repl runtime start`
  - Generate execution context for AI consumption
  - Output TASK execution prompt (stdout)

- `repl runtime status`
  - Display current TASK state and progress

- `repl runtime apply`
  - Apply AI-generated execution result to runtime state (via stdin)

- `repl runtime stop`
  - Stop runtime execution session (no state mutation)

---

## 4. Runtime Execution Model

REPL CLI does not execute AI.

It operates as a bridge between:

```text
User / AI Output
        ↓
repl runtime apply
        ↓
Validated runtime state update
        ↓
.repl/runtime/*
```

AI output is treated as external and stateless.

---

## 5. Runtime Apply Contract (MVP)

The following JSON structure must be used as input for:

```text
repl runtime apply
```

### 5.1 JSON Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "REPL Runtime Apply Schema (MVP)",
  "type": "object",
  "required": ["action", "taskId", "status"],
  "additionalProperties": false,

  "properties": {
    "action": {
      "type": "string",
      "enum": ["update_runtime"],
      "description": "Must be exactly 'update_runtime'"
    },

    "taskId": {
      "type": "string",
      "description": "Target TASK identifier matching TASKS.md (e.g., TASK_1)"
    },

    "status": {
      "type": "string",
      "enum": ["done", "blocked"],
      "description": "Final result status of the task"
    },

    "reason": {
      "type": "string",
      "description": "Required only when status is 'blocked'"
    },

    "events": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Simple execution milestone list"
    }
  },

  "allOf": [
    {
      "if": {
        "properties": {
          "status": { "const": "blocked" }
        }
      },
      "then": {
        "required": ["reason"]
      }
    }
  ]
}
```

---

## 6. Runtime State Model

All runtime state is stored locally under:

```text
.repl/runtime/
```

This includes:

- task progress state
- execution logs
- runtime session metadata

REPL CLI does not persist any external or cloud state.

---

## 7. Execution Flow

### 7.1 Runtime Start Flow

```text
repl runtime start
    ↓
Load TASKS.md
    ↓
Load .repl context
    ↓
Generate AI prompt
    ↓
Output to stdout
```

---

### 7.2 Runtime Apply Flow

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

### 7.3 Runtime Status Flow

```text
repl runtime status
    ↓
Read .repl/runtime/*
    ↓
Display TASK state summary
```

---

## 8. Multi-Organization Support

REPL CLI supports multiple GitHub organizations per workspace context.

However:

- No cross-organization state mixing
- Runtime state is always isolated per workspace

---

## 9. Data Ownership Rules

- `.repl/runtime/*` → REPL CLI owns local execution state
- GitHub → external reference only (no state authority)
- AI → stateless external processor

---

## 10. Design Constraints

- No backend service
- No daemon process
- No persistent remote synchronization
- No AI execution inside CLI
- No hidden state mutation
- No implicit runtime behavior

---

## 11. Non-Goals

REPL CLI does NOT:

- Execute AI models
- Replace GitHub
- Manage distributed systems
- Store project-level truth
- Provide collaboration features

---

## 12. Product Principle

> REPL CLI is a deterministic runtime controller that transforms external AI output into validated local state transitions.
