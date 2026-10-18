# ARCHITECTURE.md

## 1. Purpose

- repl-cli is a set of short-lived, independent command invocations.
- Each invocation reads inputs, evaluates them, optionally writes allowed files, and reports results.
- This document defines:
  - how responsibilities are separated,
  - how information flows through the system,
  - which invariants always hold.
- Product behavior is defined by PRODUCT_SPEC.md. This document defines only the internal structure that realizes it.

## 2. Core Concepts

- **Invocation**: one execution of one command. No information survives between invocations.
- **Project Directory**: the directory in which the command runs. All file access is confined to it.
- **Authoritative Documents**: the five documents named in PRODUCT_SPEC.md.
- **Canonical Text**: the canonical AGENTS.md content, obtained from the single published source defined in PRODUCT_SPEC.md.
- **Task Ledger**: the ordered list of task entries parsed from one TASKS.md text.
- **Task Entry**: a task ID plus a completion state, as recognized by the Task Ledger Reader.
- **Mark**: one occurrence of the UNVERIFIED text within a heading line of an authoritative document.
- **Fact**: a piece of information obtained from outside a pure component. Every fact has exactly one state:
  - `Present(value)`
  - `Absent` (does not exist; not an error)
  - `Failed(reason)` (existence or content could not be determined; an execution error)
- **Violation**: a failed rule, evaluated on available facts. It has a rule identifier, a file path, and a message.
- **Execution Error**: the failure to obtain a fact or to perform an allowed write. It has a subject and a message.
- **Outcome**: the structured result a coordinator returns. It holds result data, ordered violations (check only), and ordered execution errors.
- **Capability**: a granted permission to use a component's operations. Workspace capabilities are `Read` and `Write`.

## 3. System Flow

### Invocation lifecycle

1. Command Interface receives the command line.
2. Command Interface rejects an unknown command or option, or handles assistance output itself. In both cases no coordinator runs.
3. Command Interface selects exactly one Coordinator and passes a Request.
4. The Coordinator gathers facts through the I/O components.
5. The Coordinator passes facts to pure components for analysis.
6. The Coordinator returns an Outcome. It never prints.
7. Command Interface renders the Outcome as output lines and computes the exit status.

### Dependency map

```text
Command Interface
 ├─ Init Coordinator   → Canonical Document Provider, Project Workspace (Read+Write)
 ├─ Status Coordinator → Project Workspace (Read), Task Ledger Reader
 ├─ Check Coordinator  → Project Workspace (Read), Canonical Document Provider,
 │                       Base Branch Reader, Task Ledger Reader,
 │                       Unverified Mark Scanner, Rule Evaluator
 └─ Sync Coordinator   → Project Workspace (Read+Write), Canonical Document Provider,
                         Difference Calculator

Task Ledger Reader and Unverified Mark Scanner → Document Line Classifier
```

- Dependencies point only downward in this map.
- No component calls a Coordinator or the Command Interface.

## 4. Components

### C1. Command Interface

- Responsibilities:
  - Recognize the command and its options.
  - Reject an unknown command or option before any coordinator runs.
  - Resolve the base branch name for check (the given value, otherwise the default defined in PRODUCT_SPEC.md).
  - For init and sync, prompt for and validate the selected published AGENTS.md document.
  - Produce assistance output (usage help and similar) itself, without a coordinator.
  - Render each Outcome into output lines.
  - Map each Outcome to an exit status.
- Inputs: command-line arguments and, for init, the user's document selection.
- Outputs: output lines, exit status, one Request to one Coordinator.
- Ownership boundaries:
  - Owns all printed text, line layout, line order, and the exit status.
  - Does not own violation content, error content, file access, network access, repository access, or parsing.

### C2. Init Coordinator

- Responsibilities:
  - Obtain the Canonical Text for the selected init document first.
  - Make no write when the Canonical Text is `Failed`.
  - Process items in this order: AGENTS.md, then the `.replworks/` directory.
  - Mark an item `skipped` when it is present. Otherwise create it (AGENTS.md with the Canonical Text, the directory empty).
  - Create nothing else.
  - Stop at the first write failure.
- Inputs: Request with the selected document source, Canonical Text fact, Workspace presence facts.
- Outputs: Outcome with one `(item path, created | skipped)` entry per processed item, and execution errors.
- Ownership boundaries:
  - Owns the item order and the created/skipped decision.
  - Does not perform I/O itself, print, or compare content.
  - Holds the Workspace `Write` capability.

### C3. Status Coordinator

- Responsibilities:
  - Read the TASKS.md fact from the Workspace.
  - Record an execution error when TASKS.md is `Absent` or `Failed`.
  - Pass present text to the Task Ledger Reader.
  - Report total, complete count, and first incomplete task ID (or none).
- Inputs: Request (no options), TASKS.md fact.
- Outputs: Outcome with `total`, `completeCount`, `firstIncompleteId | none`, and execution errors.
- Ownership boundaries:
  - Owns the selection of facts for status.
  - Holds only the Workspace `Read` capability.
  - Does not write, print, or compute ledger derivations itself.

### C4. Check Coordinator

- Responsibilities:
  - Gather facts in this fixed order, each at most once per invocation:
    1. Read each authoritative document in the order AGENTS.md, PRODUCT_SPEC.md, TECH_STACK.md, ARCHITECTURE.md, TASKS.md.
    2. Obtain the Canonical Text.
    3. Ask the Base Branch Reader for the base ledger source. This is always attempted, even when TASKS.md is absent in the working tree.
  - Build Task Ledgers from TASKS.md texts (working tree and base).
  - Build Mark lists from the three documents that can contain marks.
  - Pass all facts to the Rule Evaluator.
  - Record every `Failed` fact as an execution error, in gathering order, regardless of whether a rule would use it.
- Inputs: Request with base branch name, all facts above.
- Outputs: Outcome with ordered violations and ordered execution errors.
- Ownership boundaries:
  - Owns the fact-gathering sequence and error aggregation.
  - Does not decide which rules run (that is the Rule Evaluator) or print.
  - Holds only the Workspace `Read` capability.

### C5. Sync Coordinator

- Responsibilities:
  - Read the local AGENTS.md fact first. Record an execution error and stop when it is `Absent` or `Failed`.
  - Only after that, obtain the Canonical Text for the selected sync document. Record an execution error and stop when it is `Failed`.
  - Compare local text and Canonical Text for exact equality.
  - When equal: return `UpToDate` and make no write.
  - When different: obtain the difference from the Difference Calculator, then replace the AGENTS.md content with the Canonical Text.
  - When replacement fails: keep the computed difference in the Outcome and record an execution error.
- Inputs: Request with the selected document source, local AGENTS.md fact, Canonical Text fact.
- Outputs: Outcome with `UpToDate | Replaced(difference) | ReplaceFailed(difference)`, and execution errors.
- Ownership boundaries:
  - Owns the order of local check, retrieval, comparison, difference, replacement.
  - Does not perform I/O itself, compute the difference itself, or print.
  - Holds the Workspace `Write` capability.

### C6. Project Workspace

- Responsibilities:
  - Provide the only access to files in the Project Directory.
  - `Read` operations: report presence of a path, read the text of a path. Each returns `Present | Absent | Failed`.
  - `Write` operations: create a file only when absent, create a directory only when absent, replace a file's whole content.
  - Judge presence by the existence of an entry at the path, regardless of entry kind.
  - Reject any path outside the Project Directory.
  - Report write success only when the full new content is stored.
- Inputs: relative paths, text content for writes.
- Outputs: facts for reads, success or `Failed` for writes.
- Ownership boundaries:
  - Owns all file-system interaction and path confinement.
  - Does not interpret content, and does not decide whether a write is allowed (the capability grant does).
  - Is the only component that writes files.

### C7. Canonical Document Provider

- Responsibilities:
  - Obtain the Canonical Text from the source selected by the request. The allowed sources are the two published Markdown documents in PRODUCT_SPEC.md.
  - Return the Canonical Text exactly as the canonical content, with no normalization.
  - Retrieve at most once per invocation. Later requests in the same invocation reuse the first result, including a failure.
  - Treat these as `Failed`: source unreachable, unsuccessful response, or an empty response from the source.
  - Keep the source's representation private. Other components never see it.
- Inputs: a source selection. Init supplies the user's selection; commands without a selection use PRODUCT_SPEC.md's default canonical document.
- Outputs: `Present(Canonical Text) | Failed(reason)`.
- Ownership boundaries:
  - Owns all network access and the extraction of the Canonical Text.
  - Does not compare, write, print, or access any address other than the canonical source.

### C8. Base Branch Reader

- Responsibilities:
  - Verify that the Project Directory belongs to a version-control repository and that the base branch exists.
  - Return the base TASKS.md at the same project-relative location as in the working tree.
  - Distinguish three results: repository or branch invalid, TASKS.md absent on the branch, TASKS.md present.
  - Never modify repository state and never use the network.
- Inputs: base branch name.
- Outputs: `Failed(reason)` (invalid repository or branch, or unreadable content) `| Absent | Present(text)`.
- Ownership boundaries:
  - Owns all repository access.
  - Does not parse the text, evaluate rules, or write anything.

### C9. Document Line Classifier

- Responsibilities:
  - Split a text into numbered lines (1-based). A line ends at LF or CRLF; the line break is not part of the line.
  - Tag each line as inside or outside a fenced code block.
  - Fence rule: an opening line starts with three or more backticks or tildes (indentation up to three spaces allowed). The block closes at a line made of the same character with at least the same count. An unclosed block extends to the end of the text.
- Inputs: text.
- Outputs: ordered list of `(line number, line text, insideFence)`.
- Ownership boundaries:
  - Owns the only definition of lines and fences.
  - Is pure. Does not recognize headings, tasks, or marks.

### C10. Task Ledger Reader

- Responsibilities:
  - Recognize Task Entries among classified lines.
  - A line is a Task Entry only when it is outside a fenced block and has the exact form: `-`, one space, `[`, exactly one of space, `x`, `X`, `]`, one space, then a task ID (the first run of non-whitespace characters), optionally followed by whitespace and any text. The line starts at column 1.
  - State is `complete` for `x` or `X`, `incomplete` for space.
  - All other lines are not Task Entries and are ignored. Parsing never fails.
  - Provide derivations over the ordered entries: total entry count, complete entry count, first incomplete task ID in order, ordered distinct task IDs.
- Inputs: text of one TASKS.md.
- Outputs: Task Ledger with the derivations above.
- Ownership boundaries:
  - Owns the Task Entry grammar and single-ledger derivations.
  - Does not compare ledgers (that is the Rule Evaluator), does no I/O.
  - Is pure.

### C11. Unverified Mark Scanner

- Responsibilities:
  - Find Marks in the classified lines of one document.
  - A heading line is a line outside a fenced block that starts with one to six `#` characters followed by a space.
  - Each occurrence of the UNVERIFIED text within a heading line is one Mark.
  - Report each Mark with the document path, line number, and heading text.
- Inputs: document path, text of the document.
- Outputs: ordered list of Marks (document order).
- Ownership boundaries:
  - Owns the heading and Mark definitions.
  - Does not decide whether Marks are violations. Is pure.

### C12. Rule Evaluator

- Responsibilities:
  - Own the five rule identifiers and their fixed evaluation order:
    1. documents-exist
    2. task-ids-unchanged
    3. single-task-completion
    4. no-unverified
    5. agents-canonical
  - Evaluate each rule only when its preconditions hold; a rule whose preconditions do not hold produces no violation.
  - Own the violation message text. Required content is listed per rule below.
  - Derive newly complete tasks from the working ledger and the base ledger.
- Inputs (Facts bundle):
  - Presence state of each authoritative document.
  - Working AGENTS.md text.
  - Canonical Text fact.
  - Working ledger fact (`Present | Absent | Failed`).
  - Base ledger fact (`Present | Absent | Failed`).
  - Mark lists for PRODUCT_SPEC.md, TECH_STACK.md, ARCHITECTURE.md (each `Present | Absent | Failed`).
- Outputs: ordered list of Violations.
- Ownership boundaries:
  - Owns rule logic, preconditions, violation ordering, and violation content.
  - Does not read, write, fetch, print, or decide how facts are obtained. Is pure.
- Rule definitions:
  - **documents-exist**
    - Precondition: the document's presence is not `Failed` (evaluated per document).
    - One violation per `Absent` document, in the fixed document order. Path is that document.
  - **task-ids-unchanged**
    - Precondition: working ledger `Present` and base ledger `Present` or `Absent`.
    - One violation per distinct base task ID that is not among the working ledger's IDs, in base ledger order. Path is TASKS.md. The message contains the task ID.
  - **single-task-completion**
    - Precondition: same as task-ids-unchanged.
    - One violation when the newly complete count is greater than one. Path is TASKS.md.
  - **no-unverified**
    - Precondition: same as task-ids-unchanged, and no Mark list is `Failed`.
    - One violation per Mark when the newly complete count is at least one, in document order PRODUCT_SPEC.md, TECH_STACK.md, ARCHITECTURE.md, then line number. Path is the document. The message contains the line number.
    - An `Absent` document contributes zero Marks.
  - **agents-canonical**
    - Precondition: working AGENTS.md `Present` and Canonical Text `Present`.
    - One violation when the two texts are not exactly equal. Path is AGENTS.md.
- Newly complete task definition:
  - An ID is complete in the base ledger when at least one base entry with that ID is complete.
  - A working entry is newly complete when it is complete and its ID is not complete in the base ledger.
  - A base ledger `Absent` means no entries.
  - Entries are counted individually; duplicates are separate entries.

### C13. Difference Calculator

- Responsibilities:
  - Compute a minimal line-level difference between a local text and the Canonical Text.
  - Return entries in document order, each marked `removed` (from the local text) or `added` (from the Canonical Text), with the line text and line position. Unchanged lines are omitted.
  - Resolve ties deterministically: `removed` entries precede `added` entries at the same position.
- Inputs: local text, canonical text.
- Outputs: ordered list of difference entries (empty only when the texts have no line-level difference).
- Ownership boundaries:
  - Owns difference computation only.
  - Does not decide whether texts differ (the Sync Coordinator compares exact equality), does no I/O. Is pure.

## 5. Component Responsibilities (summary)

| Responsibility                              | Sole owner                                              |
| ------------------------------------------- | ------------------------------------------------------- |
| Recognizing commands and options            | Command Interface                                       |
| Printed text, line order, exit status       | Command Interface                                       |
| Choosing the order of work within a command | Each Coordinator                                        |
| Reading files, presence checks              | Project Workspace                                       |
| Creating and replacing files                | Project Workspace (Write granted only to Init and Sync) |
| Network access                              | Canonical Document Provider                             |
| Repository access                           | Base Branch Reader                                      |
| Line numbering and fence recognition        | Document Line Classifier                                |
| Task Entry grammar and ledger derivations   | Task Ledger Reader                                      |
| Heading and Mark definitions                | Unverified Mark Scanner                                 |
| Rule identifiers, order, logic, messages    | Rule Evaluator                                          |
| Line-level difference                       | Difference Calculator                                   |
| Execution error aggregation                 | Each Coordinator (per invocation)                       |

## 6. Responsibility Boundaries

- Only the Command Interface prints or sets an exit status.
- Coordinators return Outcomes; they never print.
- Only the Project Workspace touches the file system.
- Only the Canonical Document Provider touches the network.
- Only the Base Branch Reader touches the repository.
- Pure components (Document Line Classifier, Task Ledger Reader, Unverified Mark Scanner, Rule Evaluator, Difference Calculator) never perform I/O and never read time, environment, or global state.
- The Rule Evaluator decides what is a violation. Coordinators decide what is an execution error. Neither does the other's job.
- The Command Interface never inspects facts; it only renders Outcomes.
- Status and Check Coordinators never receive the Workspace `Write` capability.
- No component outside the Canonical Document Provider knows the source representation of the Canonical Text.

## 7. Data Flow

### init

```text
Request → Init Coordinator
  Command Interface → validated document selection
  Canonical Document Provider(selection) → Canonical Text fact
  [Failed] → Outcome(error), no write
  [Present] → Workspace presence(AGENTS.md), presence(.replworks/)
             → Workspace Write: create absent items
  → Outcome(items, errors) → Command Interface → lines + status
```

### status

```text
Request → Status Coordinator
  Workspace Read(TASKS.md) → fact
  [Absent | Failed] → Outcome(error)
  [Present] → Task Ledger Reader (via Document Line Classifier) → Task Ledger
  → Outcome(total, completeCount, firstIncompleteId | none) → Command Interface
```

### check

```text
Request(base branch) → Check Coordinator
  Workspace Read × 5 authoritative documents → presence + text facts
  Canonical Document Provider → Canonical Text fact
  Base Branch Reader → base ledger source fact
  Task Ledger Reader × up to 2 (working text, base text) → ledgers
  Unverified Mark Scanner × up to 3 → Mark lists
  → Facts bundle → Rule Evaluator → Violations
  → Outcome(violations, errors) → Command Interface → lines + status
```

### sync

```text
Request → Sync Coordinator
  Workspace Read(AGENTS.md) → fact
  Command Interface → validated document selection
  [Absent | Failed] → Outcome(error), no write
  Canonical Document Provider(selection) → Canonical Text fact
  [Failed] → Outcome(error), no write
  exact comparison
    equal → Outcome(UpToDate)
    different → Difference Calculator → difference
              → Workspace Write: replace AGENTS.md
              → Outcome(Replaced | ReplaceFailed, difference)
  → Command Interface → lines + status
```

### Exit status mapping (owned by Command Interface)

- init, status, sync: 2 when the Outcome has at least one execution error, otherwise 0.
- check: 2 when at least one execution error exists, otherwise 1 when at least one violation exists, otherwise 0.
- Unknown command or option: one error line and 2.

### Output line content (owned by Command Interface)

- init: one line per processed item, containing the item path and whether it was created or skipped.
- status: complete count, total count, and first incomplete task ID or a no-incomplete-task indication.
- check: violation lines first (Rule Evaluator order), then execution error lines (gathering order). A violation line contains rule identifier, file path, message. An error line contains the subject and message.
- sync: an up-to-date indication, or the difference lines followed by any error line.
- All paths printed are relative to the Project Directory.

## 8. Architectural Rules

- AR-01: One invocation executes exactly one command and keeps no information afterward.
- AR-02: Every fact is exactly one of `Present`, `Absent`, `Failed`. `Absent` is never an error by itself. A coordinator turns `Failed` into an execution error and may turn `Absent` into an execution error only where PRODUCT_SPEC.md defines it (status without TASKS.md, sync without AGENTS.md).
- AR-03: Pure components are deterministic: equal inputs yield equal outputs, including order.
- AR-04: The Workspace `Write` capability is granted only to the Init and Sync Coordinators.
- AR-05: The Canonical Text is retrieved at most once per invocation, and every consumer in the invocation uses that same result.
- AR-06: A coordinator that must write obtains every fact it needs before its first write.
- AR-07: Violations are produced only by the Rule Evaluator. Execution errors are produced only by coordinators from `Failed` facts or failed writes.
- AR-08: A `Failed` fact disables only the rules that depend on it; independent rules still run.
- AR-09: All comparisons of document text are exact. No normalization of line endings, whitespace, or case.
- AR-10: Output lines are rendered only after a coordinator has returned its Outcome.
- AR-11: Violation and error ordering is fixed by this document and never depends on completion timing.
- AR-12: The Rule Evaluator's rule set and order are fixed; adding a rule is an architecture change.

## 9. Failure Boundaries

| Failure                                                       | Detected by                                 | Contained by       | Effect                                                                                                                  |
| ------------------------------------------------------------- | ------------------------------------------- | ------------------ | ----------------------------------------------------------------------------------------------------------------------- |
| Unknown command or option                                     | Command Interface                           | Command Interface  | Error line, no coordinator runs, no file touched                                                                        |
| Init or sync input ends or fails before a valid selection    | Command Interface                           | Command Interface  | Error line, exit 2, no coordinator runs, no file touched                                                               |
| Canonical Text cannot be obtained or extracted                | Canonical Document Provider                 | Coordinator in use | init and sync: stop, no write. check: agents-canonical not evaluated, other rules still run                             |
| Project Directory not in a repository, or base branch missing | Base Branch Reader                          | Check Coordinator  | task-ids-unchanged, single-task-completion, no-unverified not evaluated. documents-exist and agents-canonical still run |
| TASKS.md absent in working tree (check)                       | Project Workspace                           | Check Coordinator  | Not an error. The three ledger-dependent rules are not evaluated                                                        |
| TASKS.md absent on base branch                                | Base Branch Reader                          | Check Coordinator  | Not an error. Base ledger has no entries                                                                                |
| TASKS.md absent in working tree (status)                      | Project Workspace                           | Status Coordinator | Execution error                                                                                                         |
| AGENTS.md absent (sync)                                       | Project Workspace                           | Sync Coordinator   | Execution error, no write, no retrieval                                                                                 |
| AGENTS.md absent (check)                                      | Project Workspace                           | Rule Evaluator     | documents-exist reports it. agents-canonical not evaluated                                                              |
| An authoritative document unreadable (check)                  | Project Workspace                           | Check Coordinator  | Execution error. Rules that depend on it are not evaluated                                                              |
| Write fails (init)                                            | Project Workspace                           | Init Coordinator   | Execution error. Stop. Items already created remain and are reported                                                    |
| Replacement fails (sync)                                      | Project Workspace                           | Sync Coordinator   | Execution error. The difference is still reported                                                                       |
| Malformed task or heading lines                               | Task Ledger Reader, Unverified Mark Scanner | Same components    | Ignored. Parsing never fails                                                                                            |

- Execution error: a fact could not be obtained or a write could not be performed.
- Violation: a rule failed on facts that were obtained.
- The two never convert into each other.

## 10. Non-Goals

- No rule registry, plug-in mechanism, or user-defined rules.
- No interpretation of document content beyond: presence, exact text, Task Entries, and heading Marks.
- No caching of the Canonical Text, and no fallback source for it, across invocations.
- No rollback or transaction across multiple writes.
- No automatic repair of rule violations by check.
- No resolution or ranking of conflicts between documents.
- No interactive prompting during an invocation except the required document selection for init and sync.
- No pluggable output renderers or alternative output encodings.

## 11. Architectural Invariants

- Every file write is performed by the Project Workspace through a granted `Write` capability.
- Only the Init and Sync Coordinators ever hold that capability.
- Every network access is performed by the Canonical Document Provider.
- Every repository access is performed by the Base Branch Reader and is read-only.
- Every printed line and every exit status is produced by the Command Interface.
- Every violation is produced by the Rule Evaluator from facts, and every execution error from a `Failed` fact or failed write.
- The Canonical Text used within one invocation is always one single retrieved result from the selected or default source.
- Rule evaluation order and output order are always the fixed orders defined here.
- Task Entry and Mark recognition always follow the single shared line and fence definition.
- No component retains information between invocations.

## 12. Resolved Decisions

Each entry: ambiguity → decision → why.

- RD-01: PRODUCT_SPEC.md names printed output and exit codes but no owner → the Command Interface owns both, exclusively → one owner prevents duplicated or divergent output logic.
- RD-02: The TASKS.md location is only implied by the authoritative document list → the working-tree TASKS.md is the authoritative document `.replworks/TASKS.md`, and the base TASKS.md is the same project-relative path on the base branch → confirmed by the user.
- RD-03: The grammar of a Task Entry is undefined → exact checkbox list line form defined in C10; `x` and `X` mean complete; anything else is not an entry; lines inside fenced blocks are not entries → confirmed by the user, with the fence exclusion following the same rule used for headings.
- RD-04: Duplicate task IDs and "absent or incomplete on base" are ambiguous → entries are counted individually, and an ID is complete in base when any base entry with that ID is complete; task-ids-unchanged compares distinct IDs → keeps counts per entry (status totals) while keeping ID identity stable across branches.
- RD-05: Whether agents-canonical normalizes content is undefined → exact comparison with no normalization, also used by sync → confirmed by the user; matches "identical" in sync.
- RD-06: The canonical source's representation is not defined by this architecture → the Canonical Document Provider hides the response representation and returns the exact response text; "cannot be obtained" includes an unsuccessful response or unavailable response → source-specific representation handling stays inside the provider.
- RD-07: When the canonical text is retrieved is undefined → exactly once per invocation, using init's or sync's validated selection or the default document for check, and always attempted by init, sync (after the local check), and check → ERR-001 requires an error line from check in every case where retrieval fails, so retrieval cannot be skipped.
- RD-08: init may meet an existing AGENTS.md while the canonical text is unobtainable → init obtains the Canonical Text first and writes nothing on failure, even when AGENTS.md already exists → ERR-001 applies unconditionally to init.
- RD-09: sync with both an absent AGENTS.md and an unobtainable canonical source would produce two errors → the local check runs first and stops, so exactly one error line is produced and no network access is spent → keeps ERR-004 and ERR-001 each satisfied with one error line.
- RD-10: Repository validity in check is only tied to TASKS.md → check always verifies repository and base branch validity, even when TASKS.md is absent in the working tree → confirmed by the user; ERR-002 is unconditional.
- RD-11: FR-017 does not cover an absent AGENTS.md → agents-canonical is not evaluated; documents-exist reports the absence → confirmed by the user; avoids a double report of one cause.
- RD-12: Absent documents and the no-unverified rule → an absent document contributes zero Marks → confirmed by the user.
- RD-13: "Heading line" is undefined → a heading line is a line outside a fenced block starting with one to six `#` followed by a space; each occurrence of the UNVERIFIED text in it is one Mark and one violation → confirmed by the user and follows the Inputs definition ("the text").
- RD-14: Unreadable input files are not covered by ERR-001 to ERR-005 → a read or write failure is an execution error, and only the rules that depend on that fact are skipped; errors are recorded whenever a read is attempted, regardless of rule applicability → confirmed by the user; produces deterministic error reporting.
- RD-15: Order of check output lines is undefined → violations first in fixed rule order and fixed order inside each rule, then execution errors in gathering order → deterministic output for CI comparison.
- RD-16: Which documents and rules depend on which facts is not stated → dependencies are fixed in C12 → one place defines every precondition.
- RD-17: Behavior on an empty ledger for status is undefined → total 0, complete 0, and the no-incomplete-task indication → follows FR-008 to FR-011 literally.
- RD-18: Presence semantics for init are undefined when the path exists with another entry kind → presence is existence of an entry at the path, regardless of kind, and the item is skipped → FR-003 forbids modifying anything that exists.
- RD-19: Write failures during init are not covered → stop at the first failure, no rollback, items created before it are still reported as created, plus an execution error → keeps the output truthful about the file system.
- RD-20: Sync prints the difference and replaces the file, but the order is undefined → the difference is computed before replacement and rendered after the Outcome is returned; if replacement fails, the difference is still shown with the error → the user always sees what would have changed.
- RD-21: FR-032 (project directory only) and FR-033 (two allowed network addresses) versus the External Systems list → access to the version-control repository and either published document is allowed only through C8 and C7 respectively; every other access goes through C6 and is confined to the Project Directory → keeps all external access in single named owners.
- RD-22: Usage help and similar interface-level assistance are not listed in PRODUCT_SPEC.md → the Command Interface produces them itself, they touch no other component, and they perform no read, write, or network access → they are not use cases; PRODUCT_SPEC.md still needs a later update to list them, since ERR-005 does not mention them.
