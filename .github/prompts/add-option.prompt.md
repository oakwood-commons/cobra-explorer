---
description: "Add a new functional option to the cobra-explorer public API"
---

# Add Option

## Context

You are adding a new functional option to cobra-explorer's public API. The options pattern uses `Option func(*config)`.

## Steps

1. Add the new field to the `config` struct in `explore/options.go`
2. Create a `With<Name>(value) Option` function in `explore/options.go`
3. Add godoc comment explaining the option's behavior and default
4. Update `defaultConfig()` if the default is non-zero
5. Handle the config field in `explore/explore.go` when constructing `model.Options`
6. Pass through to the appropriate internal package
7. Update README.md options table
8. Add a CHANGELOG entry under `[Unreleased] > Added`

## Rules

- The public API must stay minimal — justify why this needs to be user-configurable
- Follow the naming pattern: `With<Thing>` for setting values, `With<Thing>Enabled` for booleans
- Default behavior should be the most common use case
- Document the default value in the godoc comment
