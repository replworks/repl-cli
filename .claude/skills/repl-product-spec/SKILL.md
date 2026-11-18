---
name: repl-product-spec
description: "Generate PRODUCT_SPEC.md from product information. Use when the user asks to create or update PRODUCT_SPEC.md. (PRODUCT_SPEC.md 작성을 요청받으면 사용)"
---

Generate PRODUCT_SPEC.md.

INPUT
- The product information provided with this request.
  Product discovery and idea validation are complete and final.
  It is the only source of product information.
- If wireframes or mockups are provided, translate their structure into text.
  MUST NOT reference or embed image files.
- Output the contents of PRODUCT_SPEC.md only.

STOP CONDITIONS
Default is to write the document. Ask only when no document can be
written without choosing between materially different products.

ASK only if the gap meets one of these:
- Two statements in the input contradict each other and neither is
  more recent or more specific.
- A feature, screen, or input is referenced by the input but its
  purpose cannot be determined from the input.
- The input leaves open a choice that changes which screens or
  commands exist.

RESOLVE YOURSELF (never ask) everything else, including:
- States, empty states, loading states, error behavior, validation,
  confirmation behavior, navigation, and element arrangement that
  follow from the stated behavior.
- Platform, performance, data retention, accessibility: write "None stated."
- Behavior with several reasonable outcomes: choose the minimal one
  that adds no new feature, screen, or input.
- Wording, copy, and visual details: human-owned, never ask.

Rules:
- If ASK items exist: output ONLY one numbered list, maximum 10
  questions, ordered by impact on the document. Nothing else.
- Each question MUST be answerable in one line and MUST list the
  candidate answers.
- Mark exactly one candidate as "Recommended" on every question.
  The recommended candidate MUST be the minimal option that adds
  no new feature, screen, or input beyond what the input states.
- If no candidate is clearly more minimal, mark the one that keeps
  the product closest to the input's stated behavior.
- I may reply with only the question numbers that differ from the
  recommendation. Every unanswered question MUST be treated as
  accepting its recommended candidate.
- After I answer, output the document. MUST NOT ask a second round.

PURPOSE
The purpose of PRODUCT_SPEC.md is to define:

What the product is.
What the product does.
How it works from the end user's perspective, including its UI and UX.
What the user can do.

SCOPE BOUNDARY
- Specify only what the end user can observe and do.
- Specify only what is to be built now. Features that were discussed but
  deferred to later MUST NOT appear in the document, except as
  Non-Goals if explicitly excluded.
- UI and UX are product behavior. MUST specify them (see USER INTERFACE RULES).
- How the product works internally (mechanism) belongs to ARCHITECTURE.md.
  MUST NOT describe implementation, architecture, technologies, tech stacks,
  programming languages, databases, internal APIs, internal components,
  or directory structures.
- Persistence is specified as observable behavior (what is kept across sessions),
  never as storage technology.
- Human-owned. MUST NOT specify:
  wording, copy, translations, images,
  visual styling (colors, typography, spacing, visual polish).
  Specify which content slots exist and what each is for, not their content.
- Business outcomes (revenue, user counts, conversion, growth, market targets)
  are out of scope. MUST NOT include them anywhere in the document.
- MUST NOT add features that were not discussed.

USER INTERFACE RULES
For every surface the user interacts with:
- Screens: give each a stable ID (SCR-001, ...), its purpose, and the elements on it.
- Elements: for each, its purpose and what the user can do with it.
- Arrangement: relative placement, grouping, and hierarchy only.
- States: every state that applies (initial, empty, loading, populated, error,
  disabled) and what is shown in each.
- Navigation: for each trigger, the destination.
- Interactions: for each user action, the observable result.
- Platform or window-size differences, only if stated in the input.
For non-graphical products (e.g. command-line):
- Every command, option, input, and the structure of output that users
  or other programs rely on, and exit behavior.
  MUST NOT specify exact output wording unless programs depend on it.

REQUIREMENT RULES
- Give every Functional Requirement a stable ID (FR-001, FR-002, ...).
- Each requirement is one externally observable behavior that can be
  verified pass/fail. If it cannot be written objectively, write the
  nearest objectively verifiable behavior.
- Product Goals: only what the product enables the end user to do.
- Error Conditions specify the observable behavior for each condition.
- Acceptance Criteria reference requirement and screen IDs and are pass/fail verifiable.
- Non-Goals: only what the input explicitly excludes.
- External Systems: list each external system the product depends on
  and the behavior relied on.
  Mark with [UNVERIFIED] the heading of any section whose behavior depends on
  an external system whose behavior has not been observed.
- Constraints: only observable non-functional requirements stated in the input
  (supported platforms or environments, performance, offline behavior,
  data retention, accessibility). If none are stated, write "None stated."

The document must include, in this order:

- Purpose
- Problem
- Product Goals
- Users
- Inputs
- Outputs
- Constraints
- External Systems
- User Interface
- Functional Requirements
- User Flows
- Error Conditions
- Non-Goals
- Acceptance Criteria

FORMAT
LANGUAGE
- Write the entire document in English, regardless of the input language.
  Translate all input content into English.
- Exception: proper nouns and identifiers stay as given.

READER
- The only reader is a coding AI agent. There is no human reader.
- Optimize for unambiguous, direct implementation and test derivation.

STYLE
- MUST NOT include background, rationale, motivation, narrative,
  introductions, summaries, transitions, or commentary.
- One behavior or fact per bullet. MUST NOT combine multiple behaviors
  in one bullet.
- Use only "MUST" and "MUST NOT" for requirements.
  MUST NOT use "should", "may", "can", "ideally", "preferably".
- MUST NOT use vague terms: "etc.", "and so on", "appropriate", "as needed",
  "fast", "user-friendly", "properly", "similar".
  Enumerate every item explicitly. If the input does not list the items,
  enumerate only those stated in the input.
- Use one term per concept. Define it once in the section where it
  first appears and reuse it exactly. MUST NOT use synonyms.
- Reference other items by ID only (FR-xxx, SCR-xxx, EL-xxx, ERR-xxx).
  MUST NOT refer to them by description.

STRUCTURE
- Flat bullets. Maximum nesting depth: 2.
- Per screen, use this fixed schema in this order:
  ID / Purpose / Elements (EL-xxx: purpose, user actions) /
  States / Navigation (trigger -> destination) /
  Interactions (action -> observable result)
- Use "trigger -> result" lines for navigation and interactions.
- Give IDs to every element (EL-), error condition (ERR-),
  and acceptance criterion (AC-).
- No emphasis styling, no emoji, no ASCII art.
