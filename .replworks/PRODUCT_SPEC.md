# PRODUCT_SPEC.md

# repl-cli

## Purpose

- repl-cli is a command-line program for REPL Works projects.
- repl-cli installs the canonical AGENTS.md into a directory.
- repl-cli reports task progress read from TASKS.md.
- repl-cli verifies a project against a fixed set of rules.
- repl-cli replaces a project's AGENTS.md with the canonical AGENTS.md.
- Project directory: the current working directory in which a command runs.
- REPL Works project: a project directory that contains AGENTS.md and a .replworks/ directory.

## Problem

- AGENTS.md is copied unchanged into every REPL Works project.
- No program verifies that a project follows the rules defined by AGENTS.md.
- No program verifies that a project's AGENTS.md matches the canonical AGENTS.md.

## Product Goals

- The product MUST enable the user to create the AGENTS.md file and the .replworks/ directory in a project directory.
- The product MUST enable the user to see how many tasks are complete.
- The product MUST enable the user to detect violations of the rules in FR-013 through FR-017.
- The product MUST enable the user to update AGENTS.md to the canonical AGENTS.md.

## Users

- A person who runs commands in a terminal.
- An AI agent that runs commands in a terminal.
- A continuous integration job that runs commands.

## Inputs

- Command and options given on the command line.
- Files in the project directory.
- Content of TASKS.md on the base branch of the git repository of the project directory.
- Canonical AGENTS.md documents:
  - https://www.repl.net/documents/agents.md
  - https://www.repl.net/documents/agents-auto.md
- Init document selection: the document selected interactively for the current `repl init` invocation. The selection is not persisted.
- Default canonical document: `https://www.repl.net/documents/agents.md`, used by commands that do not receive a document selection, currently `repl check`.
- Authoritative documents: the five files AGENTS.md, .replworks/PRODUCT_SPEC.md, .replworks/TECH_STACK.md, .replworks/ARCHITECTURE.md, .replworks/TASKS.md.
- Task: an entry in TASKS.md that has a task ID and a completion state.
- Task ID: the identifier of a task.
- Completion state: complete or incomplete, as recorded in TASKS.md.
- Newly complete task: a task whose completion state is complete in the working tree and is incomplete or absent in TASKS.md on the base branch.
- Base branch: the git branch against which the working tree is compared. The default is main.
- UNVERIFIED mark: the text [UNVERIFIED] within a heading line of .replworks/PRODUCT_SPEC.md, .replworks/TECH_STACK.md, or .replworks/ARCHITECTURE.md in the working tree.
- Violation: a failed result of one rule in FR-013 through FR-017.

## Outputs

- Text lines printed by commands.
- Exit code of each command.
- AGENTS.md file and .replworks/ directory created by repl init.
- AGENTS.md file replaced by repl sync.

## Constraints

- None stated.

## External Systems

- https://www.repl.net/documents/agents.md
  - Relied behavior: serves the default canonical AGENTS.md content and the first init option.
- https://www.repl.net/documents/agents-auto.md
  - Relied behavior: serves the automatic-run AGENTS.md content and the second init option.
- Git repository of the project directory
  - Relied behavior: provides the content of TASKS.md on the base branch.

## User Interface

- Surface: command line.
- Screens: None.

### Commands

- EL-001: repl init
  - Purpose: create AGENTS.md and the .replworks/ directory in the project directory.
  - Input: an interactive selection of one of the two published AGENTS.md documents.
  - Output: one line per created item and one line per skipped item, each containing the item path.
  - Exit: 0 on completion, 2 on execution error.
- EL-002: repl status
  - Purpose: show task progress.
  - Input: none.
  - Output: the number of complete tasks, the total number of tasks, and the task ID of the first incomplete task in TASKS.md order or an indication that no incomplete task exists.
  - Exit: 0 on success, 2 on execution error.
- EL-003: repl check
  - Purpose: verify the project against the rules in FR-013 through FR-017.
  - Input: option EL-004.
  - Output: one line per violation, containing the rule identifier, the file path, and a message. One line per execution error.
  - Exit: 0 when no violation and no execution error exist, 1 when at least one violation exists and no execution error exists, 2 when at least one execution error exists.
- EL-004: option --base <branch> of repl check
  - Purpose: set the base branch.
  - Input: a git branch name.
  - Output: none.
  - Exit: not applicable.
- EL-005: repl sync
  - Purpose: replace AGENTS.md with the canonical AGENTS.md.
  - Input: an interactive selection of one of the two published AGENTS.md documents.
  - Output: an indication that AGENTS.md is up to date, or a line-by-line difference between the local AGENTS.md and the canonical AGENTS.md.
  - Exit: 0 on completion, 2 on execution error.

### Rule Identifiers

- documents-exist
- task-ids-unchanged
- single-task-completion
- no-unverified
- agents-canonical

## Functional Requirements

### init

- FR-001: repl init MUST present the user with exactly these choices, in this order: `https://www.repl.net/documents/agents.md` and `https://www.repl.net/documents/agents-auto.md`, then create AGENTS.md with the content of the selected document when AGENTS.md does not exist.
- FR-002: repl init MUST create the .replworks/ directory in the project directory when it does not exist.
- FR-003: repl init MUST NOT modify an existing file.
- FR-004: repl init MUST NOT create any file or directory other than AGENTS.md and .replworks/.
- FR-005: repl init MUST print the path of each item it created.
- FR-006: repl init MUST print the path of each item it skipped because the item already exists.
- FR-007: repl init MUST exit with code 0 when it completes, including when it skipped items.

### status

- FR-008: repl status MUST print the number of complete tasks.
- FR-009: repl status MUST print the total number of tasks.
- FR-010: repl status MUST print the task ID of the first incomplete task in TASKS.md order.
- FR-011: repl status MUST print an indication that no incomplete task exists when every task is complete.
- FR-012: repl status MUST NOT modify any file.

### check

- FR-013: repl check MUST report a violation of rule documents-exist for each authoritative document that does not exist.
- FR-014: repl check MUST report a violation of rule task-ids-unchanged for each task ID that exists in TASKS.md on the base branch and does not exist in TASKS.md in the working tree.
- FR-015: repl check MUST report a violation of rule single-task-completion when the working tree contains more than one newly complete task.
- FR-016: repl check MUST report a violation of rule no-unverified for each UNVERIFIED mark when the working tree contains at least one newly complete task.
- FR-017: repl check MUST report a violation of rule agents-canonical when the content of AGENTS.md differs from the canonical AGENTS.md.
- FR-018: repl check MUST print each violation on its own line containing the rule identifier, the file path, and a message.
- FR-019: repl check MUST NOT modify any file.
- FR-020: repl check MUST use main as the base branch when EL-004 is not given.
- FR-021: repl check MUST use the branch given by EL-004 as the base branch when EL-004 is given.
- FR-022: repl check MUST treat the base branch as having no tasks when TASKS.md does not exist on the base branch.
- FR-023: repl check MUST NOT evaluate rules task-ids-unchanged, single-task-completion, and no-unverified when TASKS.md does not exist in the working tree.
- FR-024: repl check MUST exit with code 0 when no violation and no execution error exist.
- FR-025: repl check MUST exit with code 1 when at least one violation exists and no execution error exists.
- FR-026: repl check MUST exit with code 2 when at least one execution error exists, regardless of violations.

### sync

- FR-027: repl sync MUST present the user with exactly the same two document choices as repl init and use the selected document as its canonical content.
- FR-028: repl sync MUST NOT modify AGENTS.md when its content is identical to the selected document.
- FR-029: repl sync MUST print an indication that AGENTS.md is up to date when its content is identical to the selected document.
- FR-030: repl sync MUST print the line-by-line difference between the local AGENTS.md and the selected document when the contents differ.
- FR-031: repl sync MUST replace the content of AGENTS.md with the selected document's content when the contents differ.

### all commands

- FR-032: Every command MUST operate only on the project directory.
- FR-033: repl-cli MUST NOT access any network resource other than https://www.repl.net/documents/agents.md and https://www.repl.net/documents/agents-auto.md.
- FR-034: Only repl init and repl sync MUST write files.

## User Flows

- Flow 1: the user runs repl init in an empty directory -> the user selects one of the two published documents -> AGENTS.md contains that document and .replworks/ exists -> the user runs repl check.
- Flow 2: the user runs repl status -> the user reads the number of complete tasks and the next task ID.
- Flow 3: the user runs repl check in a continuous integration job -> the job fails when the exit code is not 0.
- Flow 4: the user runs repl sync -> the user selects a published document -> the user reads the difference -> AGENTS.md equals the selected document.

## Error Conditions

- ERR-001: The canonical AGENTS.md cannot be obtained.
  - repl init MUST NOT create or modify any file.
  - repl init MUST print an error line.
  - repl init MUST exit with code 2.
  - repl sync MUST NOT modify any file.
  - repl sync MUST print an error line.
  - repl sync MUST exit with code 2.
  - repl check MUST NOT evaluate rule agents-canonical.
  - repl check MUST print an error line.
  - repl check MUST evaluate the other rules.
  - repl check MUST exit with code 2.
- ERR-002: The project directory is not a git repository, or the base branch does not exist.
  - repl check MUST NOT evaluate rules task-ids-unchanged, single-task-completion, and no-unverified.
  - repl check MUST print an error line.
  - repl check MUST evaluate rules documents-exist and agents-canonical.
  - repl check MUST exit with code 2.
- ERR-003: TASKS.md does not exist in the working tree when repl status runs.
  - repl status MUST print an error line.
  - repl status MUST exit with code 2.
- ERR-004: AGENTS.md does not exist when repl sync runs.
  - repl sync MUST NOT create any file.
  - repl sync MUST print an error line.
  - repl sync MUST exit with code 2.
- ERR-005: An unknown command or an unknown option is given.
  - repl-cli MUST NOT modify any file.
  - repl-cli MUST print an error line.
  - repl-cli MUST exit with code 2.
- ERR-006: The init input ends or fails before the user selects one of the two init documents.
  - repl init MUST NOT create or modify any file.
  - repl init MUST print an error line.
  - repl init MUST exit with code 2.
- ERR-007: The sync input ends or fails before the user selects one of the two sync documents.
  - repl sync MUST NOT modify AGENTS.md.
  - repl sync MUST print an error line.
  - repl sync MUST exit with code 2.

## Non-Goals

- repl-cli MUST NOT execute AI models.
- repl-cli MUST NOT store runtime state or project state.
- repl-cli MUST NOT verify browser tests for graphical-interface tasks.

## Acceptance Criteria

- AC-001: Given an empty directory and a valid selection of `https://www.repl.net/documents/agents.md`, when repl init runs, then AGENTS.md exists with that document's exact content, the .replworks/ directory exists, and the exit code is 0. (FR-001, FR-002, FR-005, FR-007, EL-001)
- AC-002: Given a directory with an existing AGENTS.md, when repl init runs, then the content of AGENTS.md is unchanged, the output contains the AGENTS.md path as skipped, and the exit code is 0. (FR-003, FR-006, FR-007, EL-001)
- AC-003: Given an empty directory, when repl init runs, then no file or directory other than AGENTS.md and .replworks/ exists. (FR-004, EL-001)
- AC-004: Given a TASKS.md with 5 tasks of which 2 are complete, when repl status runs, then the output contains 2, 5, and the task ID of the first incomplete task, and the exit code is 0. (FR-008, FR-009, FR-010, EL-002)
- AC-005: Given a TASKS.md in which every task is complete, when repl status runs, then the output contains an indication that no incomplete task exists. (FR-011, EL-002)
- AC-006: Given a project directory, when repl status runs, then no file is modified. (FR-012, EL-002)
- AC-007: Given a project directory without .replworks/PRODUCT_SPEC.md, when repl check runs, then one line reports rule documents-exist with the path of that file. (FR-013, FR-018, EL-003)
- AC-008: Given a TASKS.md in which a task ID that exists on the base branch is absent, when repl check runs, then one line reports rule task-ids-unchanged with that task ID. (FR-014, EL-003)
- AC-009: Given a working tree in which two tasks are newly complete, when repl check runs, then one line reports rule single-task-completion. (FR-015, EL-003)
- AC-010: Given a working tree in which one task is newly complete and one UNVERIFIED mark exists, when repl check runs, then one line reports rule no-unverified for that mark. (FR-016, EL-003)
- AC-011: Given a working tree in which no task is newly complete and one UNVERIFIED mark exists, when repl check runs, then no line reports rule no-unverified. (FR-016, EL-003)
- AC-012: Given an AGENTS.md that differs from the canonical AGENTS.md, when repl check runs, then one line reports rule agents-canonical. (FR-017, EL-003)
- AC-013: Given a project that violates no rule, when repl check runs, then the exit code is 0. (FR-024, EL-003)
- AC-014: Given a project that violates at least one rule and has no execution error, when repl check runs, then the exit code is 1. (FR-025, EL-003)
- AC-015: Given no EL-004 option, when repl check runs, then the comparison uses the branch main. (FR-020, EL-003)
- AC-016: Given EL-004 with the branch name dev, when repl check runs, then the comparison uses the branch dev. (FR-021, EL-004)
- AC-017: Given a TASKS.md that does not exist on the base branch, when repl check runs, then every complete task in the working tree is a newly complete task. (FR-022, EL-003)
- AC-018: Given a project directory without TASKS.md, when repl check runs, then no line reports rules task-ids-unchanged, single-task-completion, or no-unverified. (FR-023, EL-003)
- AC-019: Given a project directory, when repl check runs, then no file is modified. (FR-019, EL-003)
- AC-020: Given an AGENTS.md identical to the selected document, when repl sync runs, then the file is unmodified, the output indicates that AGENTS.md is up to date, and the exit code is 0. (FR-027, FR-028, FR-031, EL-005)
- AC-021: Given an AGENTS.md that differs from the selected document, when repl sync runs, then the output contains the line-by-line difference, AGENTS.md equals the selected document, and the exit code is 0. (FR-027, FR-029, FR-030, FR-031, EL-005)
- AC-022: Given that the selected init document is unreachable, when repl init runs, then no file is created or modified, an error line is printed, and the exit code is 2. (ERR-001, EL-001)
- AC-023: Given that the selected sync document is unreachable, when repl sync runs, then AGENTS.md is unmodified, an error line is printed, and the exit code is 2. (ERR-001, EL-005)
- AC-024: Given that https://www.repl.net/documents/agents.md is unreachable and a violation of rule documents-exist exists, when repl check runs, then the violation is reported, an error line is printed, and the exit code is 2. (ERR-001, FR-026, EL-003)
- AC-025: Given a directory that is not a git repository, when repl check runs, then rules documents-exist and agents-canonical are evaluated, an error line is printed, and the exit code is 2. (ERR-002, EL-003)
- AC-026: Given a base branch that does not exist, when repl check runs, then an error line is printed and the exit code is 2. (ERR-002, EL-003, EL-004)
- AC-027: Given a project directory without TASKS.md, when repl status runs, then an error line is printed and the exit code is 2. (ERR-003, EL-002)
- AC-028: Given a project directory without AGENTS.md, when repl sync runs, then no file is created, an error line is printed, and the exit code is 2. (ERR-004, EL-005)
- AC-029: Given an unknown command, when repl-cli runs, then no file is modified, an error line is printed, and the exit code is 2. (ERR-005)
- AC-030: Given an unknown option, when repl-cli runs, then no file is modified, an error line is printed, and the exit code is 2. (ERR-005)
- AC-031: Given any command, when it runs, then no network request is sent to any address other than the two published document URLs, and init and sync request only their selected URL. (FR-033)
- AC-032: Given repl status or repl check, when it runs, then no file is created, modified, or deleted. (FR-012, FR-019, FR-034)
- AC-033: Given an empty directory and a valid selection of `https://www.repl.net/documents/agents-auto.md`, when repl init runs, then AGENTS.md exists with that document's exact content, the .replworks/ directory exists, and the exit code is 0. (FR-001, FR-002, FR-005, FR-007, EL-001)
- AC-034: Given an empty directory and an invalid init selection followed by a valid selection, when repl init runs, then repl init presents the choices again, creates the selected document, and exits with code 0. (FR-001, FR-002, EL-001)
- AC-035: Given an empty directory and an init input that ends before a valid selection, when repl init runs, then no file is created or modified, an error line is printed, and the exit code is 2. (ERR-006, EL-001)
- AC-036: Given an AGENTS.md and an invalid sync selection followed by a valid selection, when repl sync runs, then repl sync presents the choices again and completes using the selected document. (FR-027, EL-005)
- AC-037: Given an AGENTS.md and a sync input that ends before a valid selection, when repl sync runs, then AGENTS.md is unmodified, an error line is printed, and the exit code is 2. (ERR-007, EL-005)
