# Agent Instructions

## Tooling

This repository uses `mise` for tool version management and task execution.

### Tools

- **mise**: The primary tool manager. Install it from https://mise.jdx.dev/.
- **workspaced**: Used for linting and formatting. Configured via `mise`.

### Tasks

All tasks are run via `mise run <task>`.

- `lint`: Runs `workspaced codebase lint`.
- `fmt`: Runs `workspaced codebase format`.
- `test`: Runs tests (aggregates `test:*` tasks).
- `codegen`: Updates generated code (aggregates `codegen:*` tasks).
- `install`: Installs dependencies (aggregates `install:*` tasks).
- `ci`: Runs the CI pipeline (`lint` and `test`).

## CI/CD

The repository uses GitHub Actions for CI/CD. The workflow `.github/workflows/autorelease.yml` handles:
1.  Installation of tools and dependencies.
2.  Code generation and PR creation if changes are detected.
3.  CI checks (linting and testing).
4.  Automated releases and artifact uploads (for dispatch events or tags).

## Conventions

- Always use `mise` to run tasks.
- Do not install individual linters manually; use `workspaced`.
- Configure new tasks in `mise.toml` or as `[task]:*` dependencies if they are project-specific.
