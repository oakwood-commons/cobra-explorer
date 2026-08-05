---
description: "Pre-push release check for cobra-explorer: verify the CHANGELOG captures every change, confirm the README matches the current code, run all local checks (task ci) until green, then output a ready-to-run DCO-signed Conventional Commits git command WITHOUT committing."
name: "Prepare Release"
argument-hint: "Optional: release version or extra context for the commit message"
agent: "agent"
tools: ['codebase', 'search', 'editFiles', 'changes', 'runCommands', 'runTasks']
---

# Prepare Release

## Goal

Get the current branch ready to push. Verify the CHANGELOG and README are accurate,
prove that everything passes locally, and then hand the user a ready-to-execute,
DCO-signed commit command. You **generate** the commit command — you never run it.

## Hard Rules (read first)

- **NEVER run `git commit`, `git push`, `git commit --amend`, or any git write/history operation.**
  Staging, generating a message, and outputting a command are the only git-adjacent actions allowed.
- The final commit command you output MUST include `-s` (sign-off) so the DCO check passes on GitHub.
- Do not fabricate CHANGELOG or README content — every entry must be backed by an actual code change.
- If any local check fails, fix it (or report the blocker) before producing the commit command.
- Follow the repo conventions already in place: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
  [Semantic Versioning](https://semver.org/), and [Conventional Commits](https://www.conventionalcommits.org/).

## Steps

### 1. Survey the changes

Determine exactly what will be pushed so every later step is evidence-based:

- Identify the base branch (usually `main`) and the current branch (`git branch --show-current`).
- Review committed-on-branch changes: `git --no-pager diff main...HEAD`
- Review uncommitted changes: `git status` plus `git --no-pager diff` and `git --no-pager diff --staged`
- Read the changed files as needed to understand behavior, not just the diff lines.

Summarize the set of user-facing and internal changes before continuing.

### 2. Verify the CHANGELOG

- Read [CHANGELOG.md](../../CHANGELOG.md) and focus on the `## [Unreleased]` section.
- For every user-facing change from Step 1, confirm there is a matching entry under the correct
  Keep a Changelog category: **Added**, **Changed**, **Deprecated**, **Removed**, **Fixed**, or **Security**.
- Add any missing entries; fix inaccurate ones. Match the existing wording style and bullet formatting.
- Purely internal changes (tests, refactors, CI) don't need an entry unless they alter observable behavior.

**If a release version was supplied** (e.g. `0.2.0` or `v0.2.0`), also perform the version rollover:

1. **Validate it as semver** — `MAJOR.MINOR.PATCH` with an optional pre-release/build suffix. A leading
   `v` is allowed; strip it for the heading. If it isn't valid semver, or it isn't greater than the latest
   released version, stop and ask the user for a corrected version instead of guessing.
2. **Confirm there's something to release** — the current `[Unreleased]` must have at least one entry.
   If it's empty, stop and confirm with the user before creating an empty release section.
3. **Roll the section over**, preserving Keep a Changelog structure:
   - Leave a new, empty `## [Unreleased]` heading at the top (no categories under it).
   - Directly below it, add `## [X.Y.Z] - YYYY-MM-DD` (today's date, ISO format) containing the entries
     that were under Unreleased, keeping their categories and order.
4. **Rewrite the link reference block** at the bottom of the file exactly, using the repo's existing URL
   shape (tags are prefixed with `v`):
   - `[Unreleased]: <repo>/compare/vX.Y.Z...HEAD`
   - Insert `[X.Y.Z]: <repo>/compare/vPREV...vX.Y.Z`, where `PREV` is the previously latest released version.
   - Leave the oldest release's `releases/tag/vOLDEST` reference untouched.

### 3. Verify the README matches the code

Per the repo's documentation-sync rule, the README must reflect current behavior. Cross-check
[README.md](../../README.md) against the code and update anything stale:

- Functional options table vs. [explore/options.go](../../explore/options.go)
- Key bindings / navigation vs. the key handling in [internal/model](../../internal/model)
- Flag input types vs. [internal/flaginput](../../internal/flaginput)
- Built-in theme names vs. the theme registry in [internal/theme](../../internal/theme)
- Any behavioral descriptions, examples, screenshots, or code snippets that a change touched

If the observable behavior described in the README no longer matches the code, update the README
in this same pass. Leave it unchanged only if everything already matches.

### 4. Run all local checks

Run the full local CI and drive it to green:

- `task ci` (runs `fmt`, `vet`, `lint`, `test`, `build`)
- If formatting changed files, re-review and re-run.
- If a check fails, diagnose and fix the root cause, then re-run until everything passes.
- If a failure is genuinely unrelated to the branch and can't be fixed here, stop and report it
  clearly instead of producing a commit command for broken code.

Report the final pass/fail status of each check.

### 5. Generate the signed commit command (do NOT run it)

Only after Steps 2–4 are clean:

- Compose a single [Conventional Commits](https://www.conventionalcommits.org/) message
  (`type(scope): subject`) that accurately describes the branch's changes. Use a body with `-m`
  for context when the change is non-trivial.
- **If a release version was rolled over in Step 2**, make this a release commit instead: use
  `chore(release): vX.Y.Z` as the subject and summarize the notable CHANGELOG entries in the body.
- Output ONE ready-to-execute fenced `bash` block containing the sign-off (`-s`) commit command.
  Include `git add -A &&` so the command is runnable as-is (adjust if the user wants selective staging).
- Do not create git tags — tagging happens after the commit is pushed and is the user's decision.
- Present the command for the user to run — do not execute it yourself.

## Output Format

End your response with the checklist results and the command, for example:

```text
CHANGELOG: updated (added 2 entries under Unreleased > Added)
README:    updated options table; rest already current
Checks:    fmt ✓  vet ✓  lint ✓  test ✓  build ✓
```

```bash
git add -A && git commit -s \
  -m "feat(flaginput): add duration stepper input" \
  -m "Adds a type-aware stepper for time.Duration flags and documents it in the README and CHANGELOG."
```

Or, when a release version was supplied and rolled over in Step 2:

```bash
git add -A && git commit -s \
  -m "chore(release): v0.2.0" \
  -m "Cut the 0.2.0 release: duration stepper input, night theme fixes. See CHANGELOG for details."
```

## Rules Recap

- You did the verification and fixes; the user runs the commit.
- The command must be signed (`-s`) for DCO.
- No `git commit`/`git push` executed by you, ever.
