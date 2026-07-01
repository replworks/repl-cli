# .repl/agent.md

# Context Loading Order

AI must read files in this order:

1. .repl/product.md
2. .repl/framework.md
3. .repl/architecture.md
4. .repl/tasks.md

---

# Role

AI is only responsible for:

- executing tasks

AI is not responsible for:

- system design
- runtime state management
- execution orchestration
- architecture decisions

---

# Rules

AI must:

- CRITICAL: Never delete, modify, or tamper with any .md files under the .repl/ directory.
- follow tasks.md exactly
- follow product.md strictly
- follow framework.md strictly
- respect architecture.md constraints
- use runtime state only as reference
- treat all inputs as read-only context

---

# Output Contract

AI must output JSON compatible with:

```text
repl runtime apply
```

Required fields:

- taskId
- action
- status

Optional fields:

- reason (if status is blocked)
- events

AI must not deviate from the schema defined in product.md.

After completing or blocking any TASK, AI must produce a JSON payload that is directly compatible with `repl runtime apply` and must not output free-form text instead.

Example success payload:

```json
{
  "action": "update_runtime",
  "taskId": "TASK_1",
  "status": "done",
  "events": ["step1", "step2"]
}
```

Example blocked payload:

```json
{
  "action": "update_runtime",
  "taskId": "TASK_2",
  "status": "blocked",
  "reason": "dependency missing"
}
```

---

# Execution Behavior

For each TASK:

1. Load context in defined order
2. Understand TASK requirements
3. Generate solution
4. Validate against "framework" and "architecture"
5. Output runtime apply JSON

---

# Failure Rules

If TASK cannot be completed:

- status = blocked
- reason is required
- execution must stop immediately

No retry behavior is allowed.

---

# Determinism Rule

AI must behave deterministically:

- same input → same output
- no hidden memory
- no external state assumptions

---

# Core Principle

AI transforms:

TASK + context → valid runtime apply JSON
