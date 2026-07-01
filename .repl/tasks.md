# TASKS.md

# REPL CLI Execution Tasks (MVP)

> This document defines executable tasks for the REPL CLI system.

Status: MVP

---

# Execution Rules

- Each TASK corresponds to a single CLI command.
- Each TASK must be independently executable.
- Each TASK must be independently verifiable.
- TASKS are executed in isolation.
- No TASK depends on partial completion of another TASK.

---

# Failure Policy

- If a TASK fails, execution stops immediately.
- No retry is allowed.
- Error must be printed explicitly.
- Failed TASK is marked as BLOCKED.
- System does not attempt automatic recovery.

---

# TASK LIST

---

# TASK_1 — repl init

## Description

Initialize REPL project runtime environment.

## Acceptance Criteria

- `.repl/` directory exists after execution
- `.repl/runtime/` directory exists
- Default runtime state is initialized
- Execution environment is ready for subsequent commands

---

# TASK_2 — repl doctor

## Description

Validate REPL system integrity.

## Acceptance Criteria

- All required REPL directories exist
- Runtime state is consistent
- No missing or corrupted runtime files
- CLI reports system health status explicitly

---

# TASK_3 — repl reset

## Description

Reset runtime environment to clean state.

## Acceptance Criteria

- `.repl/runtime/` state is cleared
- Execution logs are removed or reset
- System returns to initial state after reset
- No residual execution state remains

---

# TASK_4 — repl runtime start

## Description

Start execution session.

## Acceptance Criteria

- Execution session is activated
- Context loading is prepared
- Runtime state reflects "active session"
- System is ready for AI execution input

---

# TASK_5 — repl runtime stop

## Description

Stop execution session.

## Acceptance Criteria

- Execution session is terminated
- Runtime state reflects "inactive session"
- No active task remains in execution mode

---

# TASK_6 — repl runtime apply

## Description

Apply external AI execution result to runtime state.

## Acceptance Criteria

- Accepts structured JSON input via stdin
- Input is validated against runtime schema
- Runtime state is updated deterministically
- Task status is updated to "done" or "blocked" (lowercase)
- On failure, execution stops immediately with error output

---

# TASK_7 — repl runtime status

## Description

Display current runtime state.

## Acceptance Criteria

- Current active task is displayed
- Runtime session status is shown (active/inactive)
- Task progress is visible
- Runtime state is readable and consistent

---

# TASK EXECUTION MODEL

```text id="exec_model"
TASK SELECTED
  ↓
EXECUTE CLI COMMAND
  ↓
VALIDATE OUTPUT
  ↓
UPDATE RUNTIME STATE
  ↓
COMPLETE OR BLOCKED
```

---

# TASK INVARIANTS

The following must always remain true:

- Each CLI command maps to exactly one TASK
- TASK execution is deterministic
- TASK completion depends only on acceptance criteria
- Runtime state is the only source of truth
- No hidden or implicit state transitions exist
