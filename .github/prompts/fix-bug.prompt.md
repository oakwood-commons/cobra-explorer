---
description: "Fix a bug in cobra-explorer with test-first approach"
---

# Fix Bug

## Context

You are fixing a bug in cobra-explorer. Follow a test-first (TDD) approach.

## Steps

1. Understand the bug: what's the expected vs actual behavior?
2. Identify the relevant package under `internal/`
3. Write a failing test that reproduces the bug:
   - Create a Cobra command fixture if needed
   - Call Update() with the triggering message
   - Assert the expected model state
4. Fix the minimal code to make the test pass
5. Run `task ci` to verify no regressions
6. Add a CHANGELOG entry under `[Unreleased] > Fixed`

## Rules

- Do NOT test View() string output — test model state
- Keep fixes minimal — don't refactor unrelated code
- If the fix changes behavior, check the design docs to confirm alignment
- For platform-specific bugs (clipboard, etc.), use build tags for targeted fixes
