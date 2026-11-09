# Env MCP Specification

This document constrains the MCP and CLI that give an agent one named, standard env per unit of work and the only allowed way to run commands in it.

Status: approved
Genre: cli + schema (two surfaces: MCP tools, and the `eval` expression language), shipped as one binary with subcommands

The key words MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY in this
document are to be interpreted as described in BCP 14 (RFC 2119,
RFC 8174) when, and only when, they appear in all capitals.

## Intention

Job: When an agent starts a unit of work, this program puts it in a standard env (a Git worktree checkout with a fixed layout), addresses that env by handle, and runs every long-lived command inside it as a named, observable exec, so the user never relies on the agent's own choice of directory, terminal layout, and log handling. The same operations are reachable from an MCP client and from a terminal.

Non-goals:

1. A second Herdr. Herdr stays the terminal multiplexer. Neither surface asks the agent to name a workspace, a tab, a pane, and a Herdr command.
2. A general shell. Expressions start processes from `argv`. The language has no shell grammar, no globbing, no pipes, and no redirects.
3. Remote envs. v0 places every env on the local machine.
4. Pull requests, issue tracking, and credentials.
5. Resuming a closed env from durable storage.
6. A package manager. A build system. A test runner.
7. A file editor with fuzzy matching. A file editor with structural matching. `edit` replaces one exact unique string.
8. A REPL. `eval` is stateless.
9. A second constitution at any other path.
10. Two implementations. The MCP surface and the CLI surface are renderings of one operation table.

Inherited C (cite the file):

- None. The tree is empty; every tool below is a decision.

## Technique

| ID | Input | Rule | Output |
|----|-------|------|--------|
| TEC-01 | A repo identity `github:owner/repo@base` and a slug | Resolve the identity to a clone (the declared one, else a clone the store makes), then ensure one linked worktree at the slug branch, reusing the worktree already on that branch. One `(repo, slug)` pair is one env for the life of its record. | One env handle with one status token: `created` or `resumed` |
| TEC-02 | An env handle and the agent's current location | Place the agent's own session in the env checkout; repeat calls are a no-op beyond re-ensuring the standard layout. | The agent standing in the env |
| TEC-03 | One expression source, a map of named arguments, and a handle table | Parse into a tree, type-check the whole expression against the closed item table, then evaluate; a type error executes nothing. | One JSON value plus any handles minted |
| TEC-04 | A typed operation on an item | Every operation is total on its declared item types; an argument of another type fails the expression before evaluation. | A value of a declared type |
| TEC-05 | `argv`, an env, and an optional relative cwd | Start the process as a child of the env, append its output to an event stream, and return an exec handle immediately. | An exec handle with an offset-addressed stream |
| TEC-06 | An event stream, a line offset, and an optional RE2 pattern | Read whole lines from the offset, apply the pattern server-side, advance the cursor past every scanned line. | Lines, a next offset, and truncation notice |
| TEC-07 | A wait condition from the closed set `literal`, `pattern`, `exit`, plus a finite timeout | Poll the event stream until the condition holds; on timeout report the scanned lines and the state. | The stream handle and the matched lines |
| TEC-08 | A handle-addressed path | Resolve relative to the env root; reject `..`, an absolute path, and any resolved path outside the env root. | A file inside the env |
| TEC-09 | A new env and the layout record | Create one labeled place for the env with one agent shell; create one more labeled place per exec on demand. Never touch a place the server did not create. | A standard layout the user can see |
| TEC-10 | One operation from the table, its arguments, and a `format` | Run the operation, then render its `Result` through the chosen `format`. The MCP surface passes arguments as typed tool inputs; the CLI surface passes them as flags and positional arguments. | One rendered result with one exit status |

## Tooling

| TEC | Tool | Relation | We do not | Cite |
|-----|------|----------|-----------|------|
| TEC-03 | `lewkit/x/parse` | adopt | write an LR parser; write a lexer | lewtec/lewkit `x/parse/doc.go` |
| TEC-03 | `github.com/modelcontextprotocol/go-sdk` | wrap | implement JSON-RPC; implement the transport; hand-write tool schemas | modelcontextprotocol/go-sdk README |
| TEC-10 | `lewkit/x/cmd` | adopt | implement subcommand dispatch; implement flag parsing; implement usage text; implement `--version`; implement exit codes; implement the `format` argument | lewtec/lewkit `x/cmd` |
| TEC-10 | `lewkit/x/text/table` | adopt | implement a table renderer; invent a CSV or JSONL encoder | lewtec/lewkit `x/text/table/format.go` |
| TEC-09 | Herdr socket API already in this path | wrap | implement a terminal multiplexer | lewtec/herdr-reflow `pkg/herdr/client.go` |
| TEC-01 | `internal/store`: clones and worktrees | implement | re-derive paths in each backend | `internal/store/store.go` |
| TEC-01 | `lewkit/x/git` and `git clone` | adopt | implement Git worktree bookkeeping | lewtec/lewkit `x/git` |
| TEC-01 | `git worktree add` | adopt | reimplement worktrees | `git` binary |
| TEC-01, TEC-02 | Herdr `worktree.create`, `worktree.open`, `pane.move` | wrap | implement checkout placement; implement sidebar layout | lewtec/herdr-reflow `pkg/herdr/reorder.go` |
| TEC-05 | the `runner.Executor` seam over Herdr `tab.create`, `pane.send_text`, `pane.send_keys` | adopt | implement a PTY; call Herdr from the operation table | `internal/runner/runner.go`, lewtec/herdr-reflow `pkg/herdr/client.go` |
| TEC-06, TEC-07 | `runner.Executor.Read` over Herdr `pane.read`; the sentinel line carries the exit status | wrap | implement terminal emulation; implement scrollback | `internal/runner/herdr`, lewtec/herdr-reflow |
| TEC-04, TEC-05, TEC-06, TEC-07, TEC-08 | the type checker, the evaluator, and the path resolver | implement | adopt a general Lisp interpreter | none |
| TEC-10 | output formatting | adopt | invent a serialization contract | lewtec/lewkit `x/text/table` |

| Cell | Pick | C or D | Implements | Cite if C |
|------|------|--------|------------|-----------|
| Language | Go 1.27 | C | TEC-03 | `lewkit` and `herdr-reflow` are Go; `x/parse` is a Go library |
| Runtime | one binary, `env-mcp`, with an MCP transport and terminal subcommands | D | TEC-03, TEC-10 | |
| Persistence | server state directory: one JSON record per env and one append-only event stream per exec | D | TEC-01, TEC-06 | |
| UI | none; Herdr is the visible surface | D | TEC-09 | |
| Packaging | `mise` project tasks and a GoReleaser release, like `herdr-reflow` | D | TEC-03, TEC-10 | |
| Identity | env handle is `repo/slug`, minted by the server, valid for the life of the record | D | TEC-01 | |
| Host OS | macOS and Linux | D | TEC-09 | |

Three cited gaps stay open by decision:

- No Lisp evaluator exists in lewkit and none exists in a LEWTEC project (`none` after reading `lewkit/x/*` and `lewtec/herdr-reflow`). TEC-04, TEC-05, TEC-06, TEC-07 and TEC-08 are ours.
- No typed item table exists anywhere in the org. The table in Types is the whole definition.
- Herdr has no typed handle layer; the wrapper maps one MCP exec to one Herdr place and nothing more.

## Terminology

| Concept | Approved | Banned |
|---------|----------|--------|
| The unit of work the agent stands in | env | workspace, sandbox, container, session, branch |
| The identity the agent passes to name a repo | repo | origin, remote, project |
| A started process with an event stream | exec | job, process, task, slot, background task |
| The handle-addressed output of an exec | event stream | log buffer, scrollback, stdout, terminal |
| The named place in Herdr that shows an exec and the agent shell | place (internal only) | tab, pane, workspace, window (in any public surface) |
| A value in the expression language | item | object, value, node |
| One expression source | expr | script, program, query, command |
| The terminal surface of the same operations | CLI surface | client, wrapper, debug mode |
| The layout rule for a new env | standard layout | convention, template, defaults |

`tab`, `pane`, and `workspace` MAY appear in code and in Herdr-facing internal types. They MUST NOT appear in a tool name, a tool argument, a tool result, a CLI subcommand name, a CLI flag, and the expression language.

## Types

| Type | Exported | Identity or value | Mutable | Nil/error | Callers MUST NOT |
|------|----------|-------------------|---------|-----------|------------------|
| `Repo` | yes | value; the string `github:owner/repo@base` | no | error on a wrong scheme; error on a wrong segment count; error on an empty base | construct one that skips `Parse` |
| `Env` | yes | identity; `repo/slug` | no | error on an unknown handle | pass an `Env` that `env` did not mint; expect a checkout path field |
| `Exec` | yes | identity; minted per start | no | error on unknown handle | reuse an `Exec` after `env close` |
| `Stream` | yes | value handle over one `Exec` at an offset | no | error when the offset exceeds the stream | read past `truncated_from` without noticing |
| `File` | yes | value; an env-relative path | no | error when the path escapes | escape the env root; pass an absolute path |
| `Value` | yes | sum of the item types and the JSON scalars and lists | no | error on a type mismatch | assume an untagged scalar is an item |
| `Result` | yes | identity of one call: its value, its minted handles, and its effect trace | no | error only for protocol failure | treat a reported failure as success |
| `Layout` | no | one env record: places the MCP created | yes | missing record means no env | touch a place absent from the record |

The expression grammar has one form: an application `(op arg…)`. The grammar has no `let`, no `begin`, no lambda, and no user-defined name. `args.NAME` is the only variable form.

### Operation table

Each row is the complete set of operations on one item. No other operation exists.

| Item | Operations |
|------|-----------|
| `repo` | `repo "github:owner/repo@base"` |
| `env` | `env repo slug`; `list`; `close`; `file path`; `exec argv [cwd]` |
| `file` | `read`; `write str`; `edit old new`; `glob pat`; `grep pat` |
| `exec` | `wait literal pat timeout`; `wait pattern pat timeout`; `wait exit timeout`; `poll`; `logs`; `kill` |
| `stream` | `since offset`; `grep pat`; `tail n` |

`wait` has three forms. The first token after `wait` selects the form, and the form fixes the arity. The token is never a free-form string.

`env repo slug` and the tool `env_get` are the same operation with the same rules.

### Surfaces

One binary, `env-mcp`, with `x/cmd` subcommands. Every operation in the table is reachable from both surfaces, with the same arguments and the same rules. Each surface passes arguments the way its client expects: typed inputs over MCP, flags and positionals over the CLI.

| Surface | Entry | Arguments | Result rendering |
|---------|-------|-----------|------------------|
| MCP | the `eval` tool and the `env_get`, `exec`, `logs`, `poll`, `list`, `close` tools over stdio | typed tool inputs, standard MCP tool calling | the `format` argument, default `jsonl` |
| CLI | `env-mcp env get github:owner/repo@base <slug>`, `env-mcp env list`, `env-mcp env close <env>`, `env-mcp exec <env> <argv…>`, `env-mcp logs <exec>`, `env-mcp poll <exec>`, `env-mcp eval '<expr>' --arg NAME=VALUE`, `env-mcp serve` | flags and positional arguments; a trailing variadic argument accepts `--` and runs the same without it | the `format` argument from `x/cmd`, default `table` |

`serve` runs the MCP transport. Every other subcommand runs one operation and exits.

The MCP tool `env_get` and the expression operation `env repo slug` are the same operation. The CLI spelling is `env get`. One operation, three spellings, one rule set.

## Invariants

| ID | Predicate | On | Forbidden bypass |
|----|-----------|----|------------------|
| INV-01 | One `(repo, slug)` pair is one env with one worktree, for the life of the record | TEC-01 | creating a second worktree for the same branch |
| INV-02 | No MCP result and no expression value contains a checkout path, a host name, and a Herdr identifier | TEC-08, TEC-09 | returning a checkout path in a tool result; returning a tab id |
| INV-03 | A type error executes zero operations | TEC-04 | evaluating the prefix of a bad expression |
| INV-04 | Every file path and `exec` cwd resolves inside the env root | TEC-08 | resolving `..`; resolving an absolute path |
| INV-05 | Every exec is a child of its env and dies when the env closes | TEC-05 | leaving an exec running after `close` |
| INV-06 | The server closes, reorders, and renames only places in its own layout record | TEC-09 | touching a user tab |
| INV-07 | An event stream is append-only and is never rewritten | TEC-06 | rotating a stream in place |
| INV-08 | The stream cursor advances past every scanned line, whether it matched | TEC-06 | re-scanning unmatched lines |
| INV-09 | `eval` carries no state between calls except handles | TEC-03 | a definition visible in a later call |
| INV-10 | A close frees the env handle; later calls with it fail | TEC-01 | serving a closed handle |
| INV-11 | Every operation in the table is reachable from the MCP surface and from the CLI surface with identical arguments and identical rules | TEC-10 | a CLI flag with no expression or tool argument; a tool operation with no subcommand |
| INV-12 | An env holds at most 8 live execs | TEC-05 | starting a ninth exec on one env |

## Errors

| Public operation | Bad input | One reaction |
|------------------|-----------|--------------|
| `env_get(repo, slug)` and `env repo slug` | a malformed repo; an unknown scheme; a missing local checkout; an invalid branch slug | reject with a message naming the failing part; create nothing |
| `env_get` with a stale handle | closed handle | reject; the agent MUST call `env` again |
| `list` | none | empty list |
| `close` | unknown env | reject |
| `read`, `write`, `edit`, `glob`, `grep` | an absolute path; a `..` escape; a missing file | reject; no partial write |
| `edit` | the old string is absent; the old string occurs more than once | reject; file unchanged |
| `write` | parent directory missing | reject; no implicit mkdir |
| `exec` | an empty argv; a cwd that escapes; a closed env | reject; no process started |
| `exec` | the env is at its concurrent exec limit | reject; name the limit |
| `logs`, `since` | offset past the end of the stream | reject; report the last valid offset |
| `logs`, `grep`, `wait` | invalid RE2 pattern | reject before reading |
| `wait` | timeout elapsed without the condition | **fail** with the scanned lines so far and the state |
| `wait` | the process exited without the condition holding | fail with the exit code |
| `kill` | already exited | succeed; killing is idempotent |
| `eval` | a syntax error; an arity error; an unknown operation; a wrong item type | reject before any execution; report the failing subexpression |
| `eval` | runtime failure after a side effect | fail; report the effects already performed |
| `eval` | an `args` name not free in the expression; a free name absent from `args` | reject before execution |
| MCP transport | an unknown tool; malformed JSON; an argument that fails the tool's input schema | reject with a JSON-RPC error; the server stays up |
| CLI | an unknown subcommand; an unknown flag; a missing required argument; an invalid `format` | `x/cmd` prints usage to stderr and exits non-zero; no operation runs |
| CLI | the operation fails | print the failure through `format` and exit non-zero |

## Actors

| Actor | Obligations |
|-------|-------------|
| Agent | Address work by env handle and operation only; reach files inside an env through item operations; stop and close what it started |
| User | Watches places in Herdr; owns every tab the server did not create; may close and rename any place |
| Server | Enforce the operation table, the env-root resolver, the layout record, and the exec limit on both surfaces |
| Herdr | Renders places, owns terminal state, and reports output; is never addressed by the agent |

## Capabilities

| ID | Actor | Sea-level goal |
|----|-------|----------------|
| CAP-01 | Agent | Enter the standard env for a unit of work in one call |
| CAP-02 | Agent | Read, write, edit, glob, and grep files inside an env without naming a path |
| CAP-03 | Agent | Start a long-lived command and continue working while it runs |
| CAP-04 | Agent | Compose start, wait, and filter into one expression with named inputs |
| CAP-05 | Agent | Learn the result of a finished command through its exit code |
| CAP-06 | User | See every exec and the agent shell as separate places, labeled per env |
| CAP-08 | User | Run any of the same operations from a terminal through the CLI |
| CAP-07 | Agent | Close an env and be sure nothing it started is still running |

## Quality

| Concern | Measure, or why it cannot happen |
|---------|----------------------------------|
| Exit contract | Every result is tagged `value` or `failure`. A `failure` names the effects already performed. The CLI exits non-zero for a `failure` and zero for a `value`. The same failing expression reports the same failing subexpression on every run. |
| Untrusted input | Every path argument resolves under the env root; the resolver has a table test with `..`, absolute, symlink-escaping, and empty inputs, and each one is rejected |
| Identity | A second `env` call for one `(repo, slug)` creates no worktree: a test calls it twice and asserts one `git worktree list` row and one layout record |
| Layout safety | A test creates a user tab in the env place, closes and re-opens the env, and asserts the user tab is untouched |
| Stream integrity | A test writes 10k lines, reads from offset 0 in windows, and asserts the concatenation equals the source and `next_offset` is monotonic |
| Type safety | A test evaluates an expression whose last operation has a wrong item type and asserts the first operation did not run |
| Surface parity | A test enumerates the operation table and asserts each operation is reachable as a CLI subcommand and as an MCP tool argument path; the test fails when either surface loses one |

## Security

In scope:

- The env root is the boundary. An expression cannot read or write outside it.
- `exec` runs a process as the same OS user with the env root as cwd. There is no privilege boundary between the agent and the user's account.
- The server and Herdr share the user's OS account and the local socket.
- The CLI runs with the same authority as the MCP surface, so it is not a weaker channel.

Why it cannot happen (if claiming none):

- No path escape, per the resolver invariant and its table test.
- No hidden second state: `eval` is stateless and holds no definitions.

Residual risk:

- An agent that starts `bash -c '…'` as its own argv steps outside the operation table's intent. `argv` is not a sandbox; the operation table constrains composition, not intent.
- An exec can print a path. The API shape keeps paths out of results; it does not police text, and it MUST NOT try to.
- A compromised Herdr socket is a compromised user session regardless of this MCP.

## Success

- [ ] A new repo plus slug produces one worktree, one labeled place, and one agent shell, and the agent can edit a file in it without ever receiving a path.
- [ ] Calling `env_get` twice for one `(repo, slug)` leaves exactly one worktree and one layout record.
- [ ] One `eval` expression starts a server, waits for a line, and returns the matching lines, in a single MCP call.
- [ ] A shell command that fails with exit code 3 is observable as exit code 3 through the exec handle.
- [ ] Any path containing `..` and any path starting with `/` is rejected by every file and cwd operation.
- [ ] Closing an env leaves no process it started running.
- [ ] Every operation in the table runs from a terminal through the CLI with the same rules it has through the MCP.
- [ ] An operation that fails exits non-zero in the CLI and reports a tagged failure through the MCP.

## Later work

1. Remote envs: the same item table over a machine boundary.
2. Pull-request creation from an env.
3. Resuming a closed env from its record.
4. `begin` and named definitions inside one expression.
5. Push notifications of stream growth.
6. `wait` on a socket or an HTTP health check.
7. Multiple agents sharing one env.

## Assumptions

| ID | Fact | If false |
|----|------|----------|
| AS-01 | Herdr 0.9.5's socket API (protocol 22) keeps `worktree.open`, `tab.create`, `pane.send_text`, `pane.wait_for_output`, and `pane.move` on this machine | TEC-09 has no implementation and the layout rules move to `git` plus a plain PTY |
| AS-02 | `lewkit/x/parse` can express one application form with typed tokens | TEC-03 needs a hand-written lexer and a recursive-descent parser |
| AS-05 | `lewkit/x/cmd` provides the `format` argument, subcommand dispatch, and exit codes named in TEC-10 | TEC-10 has no implementation and the CLI owns its own flags and rendering |
| AS-03 | The user runs one server per Herdr session | Two servers would mint overlapping handles and race on the layout record |
| AS-04 | When no clone is declared, `git clone` reaches the host on its own credentials, per ADR-0008 | `env` cannot create an env for an undeclared repo, and the failure names the missing clone |

## Decision history

- ADR-0001: Herdr is the layout engine and the MCP derives location from Herdr state. Rejected: owning a second layout model beside Herdr's sidebar.
- ADR-0002: The agent's vocabulary is env and exec only. Rejected: exposing workspace, tab, and pane names.
- ADR-0003: The contract returns no path; every file access goes through handle operations. Rejected: returning the checkout path as the local leak.
- ADR-0004: `env_get` moves the agent into the env. Rejected: telling the agent to `cd` itself.
- ADR-0005: The public surface is a stateless typed expression and a small fixed tool set. Rejected: a tool-only surface; a stateful REPL.
- ADR-0006: Output is an offset-addressed append-only event stream filtered server-side by RE2. Rejected: a port probe as readiness evidence; an in-memory output buffer.
- ADR-0007: `exec` starts one named process and returns it immediately; readiness is a separate `wait`. Rejected: blocking `exec` on a readiness heuristic.
- ADR-0008: The server owns the checkout store. For one `github:owner/repo`, it uses a clone declared in config when there is one and otherwise clones the repo into its own store; it then cuts one linked worktree per `(repo, slug)` from that clone. Owners: the store alone decides where clones and worktrees live, and every backend receives a finished checkout. Cloning uses the host's own git credentials; the server holds none of its own. Rejected: requiring a pre-existing checkout, which made every repo a manual setup step.
- ADR-0009: The language has exactly one application form and no `begin`. Rejected: `begin` and named definitions in v0.
- ADR-0010: One binary with subcommands; every operation is reachable from the MCP surface and the CLI surface with identical rules. Rejected: a separate debug CLI with its own rules.
- ADR-0011: `x/cmd` owns dispatch, flags, exit codes, and rendering through `format`. Rejected: a bespoke `--json` flag and a hand-written serializer.
- ADR-0012: `Env` carries no path; the API shape keeps paths out of results and text is not policed. Rejected: rewriting absolute paths inside exec output.
- ADR-0013: An env holds at most 8 live execs. Rejected: an unbounded exec count.
- ADR-0014: Execution sits behind one `Executor` seam (`Open`, `Start`, `Status`, `Read`, `Kill`, `Close`) registered as a lewkit driver. `internal/runner/herdr` is the default (weight 90) and `internal/runner/host` is the daemon-free fallback (weight 10), opt-in through `ENV_MCP_ALLOW_HOST_RUNNER`; the registry picks by weight and compatibility, so losing the Herdr session is a clear error unless the caller opts into loose processes. The record stores `runner` and an opaque `place`, never a PID or tab. Rejected: an own daemon; a hard Herdr dependency; an abstract interface with one implementation.
