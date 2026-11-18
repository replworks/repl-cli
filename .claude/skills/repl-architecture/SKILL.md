---
name: repl-architecture
description: "Generate ARCHITECTURE.md defining components, responsibilities, and data flows. Use when the user asks to create or update ARCHITECTURE.md. (ARCHITECTURE.md 작성을 요청받으면 사용)"
---

Generate ARCHITECTURE.md.

Assume PRODUCT_SPEC.md and TECH_STACK.md already exist and are correct.
ARCHITECTURE.md derives from TECH_STACK.md. If TECH_STACK.md changes,
ARCHITECTURE.md is regenerated.

Read TECH_STACK.md to know which structure the stack already provides.
Do not restate or reference any technology, version, or convention from it.

If PRODUCT_SPEC.md and TECH_STACK.md conflict, or if a required component
cannot be reconciled with TECH_STACK.md, stop and report the conflict.

Do not create any other document.

The purpose of ARCHITECTURE.md is to define:

How the product works internally.
How responsibilities are separated.
How information flows through the system.
Which invariants must always remain true.

SCOPE BOUNDARY
- Define only what PRODUCT_SPEC.md requires and the stack's standard
  structure does not already provide.
- Anything the stack's standard structure already decides applies by default
  and is not defined here.
- Stay technology-agnostic: no technologies, languages, libraries,
  coding conventions, deployment, or folder structure.

The document must include:

- Purpose
- Core Concepts
- System Flow
- Components
- Component Responsibilities
- Responsibility Boundaries
- Data Flow
- Architectural Rules
- Failure Boundaries
- Non-Goals
- Architectural Invariants
- Resolved Decisions

Every component must have:

- Responsibilities
- Inputs
- Outputs
- Ownership boundaries

Before generating, identify:

- conflicting responsibilities
- ambiguous flows
- responsibilities PRODUCT_SPEC.md requires that the stack's standard
  structure does not provide

If resolving one requires interpreting PRODUCT_SPEC.md in a way
that is not explicitly stated there, ask. Otherwise resolve it yourself
and record it in "Resolved Decisions" (what was ambiguous, what you decided, why).

Non-Goals must contain only architectural non-goals
(structures or responsibilities this architecture deliberately does not provide).
Do not copy non-goals from PRODUCT_SPEC.md.

Prefer clear ownership over abstraction.
The document is intended for AI implementation, not human documentation.
