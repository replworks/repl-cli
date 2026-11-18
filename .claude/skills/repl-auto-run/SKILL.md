---
name: repl-auto-run
description: "Execute TASKS.md tasks end to end according to AGENTS.md AUTO_RUN. Use when the user asks to run the tasks autonomously or says to proceed automatically. (자동으로 진행해 달라는 요청을 받으면 사용)"
---

Read AGENTS.md, then start working on TASKS.md.

The rules below extend AGENTS.md's "Stop and ask the human".
They change nothing else in AGENTS.md. If anything conflicts with
AGENTS.md, AGENTS.md takes precedence — except for "how to reach
the human", which follows these rules.

## How to ask

### Phase 1: before the first task branch exists (bootstrap)
- Ask the human directly in the session.
- Do not use ask-human.sh. Do not poll.

### Phase 2: after AUTO_RUN starts
- When a "Stop and ask the human" situation arises during AUTO_RUN,
  do not just type the question into the session and end the turn.
  Post the question with ask-human.sh and wait for the answer.
- Do not guess and proceed instead of asking.

Channel selection:
- Current branch has code changes (commits ahead of main, or uncommitted
  changes) -> PR channel
- No code changes -> Issue channel

PR channel:
1. If the branch has no PR yet: commit uncommitted changes as
   `chore: wip t-NNN` (bot account), push the branch, and open a draft PR
   as the bot account (Conventional Commits title, assignee: bot account).
   This draft PR is the task's final PR. Never open a second PR for the task.
   When the task is done, mark it ready and perform AUTO_RUN step 6.
2. Run `.replworks/scripts/ask-human.sh ask pr "<question>"`.

Issue channel:
1. Run `.replworks/scripts/ask-human.sh ask issue "<question>"`.
   One question per issue. Do not create commits or PRs for a question.

Question format:
- Multiple choice by default; free response allowed when it fits better.
  2-4 options, one line each.
- Mark the recommended option with "(recommended)".
- Ask "what is correct", not "how the code works".
- Link the relevant document section or diff.

Waiting for the answer:
1. ask prints a token. Run `.replworks/scripts/ask-human.sh wait <token>`.
2. If the output is `PENDING`, run the same command again, up to 120 times.
   Do nothing else while waiting.
3. Any other output is the human's answer. Continue accordingly.
4. After receiving the answer, react to the comment:
   - proceeded as answered: `+1`
   - ambiguous, asked again: `confused`
   - blocked by a conflict with AGENTS.md: `-1`
5. If there is still no answer after 120 polls, leave a comment on the PR
   or issue stating the task ID and the condition being waited on, and stop.
   Do not modify TASKS.md. Do not proceed arbitrarily.

If ask-human.sh fails (e.g. no PR), read the error, fix it as described
above, and run it again. If the script itself does not work, report that
in the session and stop.
