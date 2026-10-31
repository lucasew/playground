# the-lang specification

This document is the law for the the-lang language and its toolchain command.

Status: draft
Genre: cli

The key words MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY in this
document are to be interpreted as described in BCP 14 (RFC 2119,
RFC 8174) when, and only when, they appear in all capitals.

## Intention

Job: A language where the only concurrent work is an actor that owns its memory, so shared mutation is unrepresentable and the program can see when each value dies.

Non-goals:

1. Actors outside this OS process. No cluster. No remote pid. No location transparency.
2. Compiling Go source. A Go idea is allowed only when the conversion keeps the original semantics.
3. Macros.
4. Class hierarchies and inheritance.
5. A second death API. No finalizer. No `Context`.
6. A mailbox capacity argument.
7. A pin keyword.

Inherited C:

- path:Cargo.toml — package `the-lang`, edition 2024. The compiler host is Rust. argv parsing is `usage-rs`.
- path:mise.toml — Rust 1.93.0.
- path:src/main.rs — the toolchain binary starts as a hello-world. It is not the language.

## Technique

| ID | Input | Rule | Output |
|----|-------|------|--------|
| TEC-01 | argv of one toolchain binary | The commands are `build`, `run`, `test`, and `generate`. Flags follow the usage argv grammar. Exit 0 is success, help, and version. Exit 2 is bad argv. Exit 1 is a command that ran and failed. | status and streams in the Public contract |
| TEC-02 | `spawn` of an actor | The runtime creates one actor with a fresh pid. A pid is a name. It does not keep the actor alive. The pid is never reused. | a pid of that actor type |
| TEC-03 | a call on a pid | The call returns `Promise[Result[T]]` when the handler returns `T`. `await` freezes the caller until the handler returns. `ok` carries the handler value. A dead callee is `err(error.dead)`. | `Result[T]` |
| TEC-04 | calls that arrive faster than the actor runs | The program passes no capacity. The runtime bounds the queue. Extra calls wait on their promises. A busy actor does not fail the call. | the same `Result[T]`, later |
| TEC-05 | a place that holds a value | Setting the place ends the previous value. The place ending ends the current value. Actor death ends the places it owns. `__close` runs then, once. The collector reclaims unreachable memory and runs no user code. | closed value, then free memory |
| TEC-06 | an `err` payload | The payload is an `error` value, not an enum. `__unwrap__` builds the cause chain. `is` walks that chain for a sentinel. `find[T]` walks it for a type. | `is` returns bool. `find` returns `Option[T]`. |
| TEC-07 | `as[Target](value)` | Try the source `__to__` whose result is `Target`. Then try `Target.__from__`. Then try a defined type and its underlying type. An `Err` does not continue the search. | the return type of the selected method |
| TEC-08 | an integer literal | `int` is 32-bit signed. `usize` is unsigned and has the pointer width. A literal in a typed place takes that place's type when the value fits. A computed `int` needs `as` to become another integer type. `==` on two values of the same integer type returns `bool`. | a value of the place's type |
| TEC-09 | `type Name enum` | Each variant is one of three forms: a name, ordered fields, named fields. `match` names every variant. `Result` and `Option` are prelude enums. | one variant |
| TEC-10 | an import line | `import { A, B } from "specifier"` binds those names. `import mod from "specifier"` binds the whole module, and calls are `mod.A`. An import that uses `*` is rejected. A specifier that starts with `./` or `../` is a path inside the module. A first element with a dot is a remote module path. The default URL is `https://` plus that path. A first element with no dot is the standard library. The string has no version and no scheme. `lang.mod` records requirements. `lang.lock` records checksums. Selection is minimal version selection. | bound names, or one module value |
| TEC-11 | `runtime.lockOSThread` | The call pins the actor to its current OS thread. No other actor runs on that thread until the matching unlock count. The main actor starts on the OS main thread. | a pinned actor |
| TEC-12 | a method declaration | Fields stay in the actor block and in the struct block. Methods stay outside. `self` is in scope. The last expression is the result. A name may have several argument lists. | a callable method |
| TEC-13 | `+`, `-`, `*` on a builtin integer | The operation wraps modulo 2 to the width. Signed values use two's complement wrap. | a value of the same integer type |
| TEC-14 | `build` and a program path | `build` writes one file. `--artifact binary` writes an executable. `--artifact c-shared` writes a C shared library. Absent `--artifact` means `binary`. `--os` and `--arch` select the target. Absent flags mean the host OS and the host architecture. `--output` is the file path. Absent `--output` means the base name of the path, in the current directory. | one file, empty stdout |
| TEC-15 | a `test` block | `build` and `run` omit every `test` block. A top-level `test` block is one short-lived actor. The `test` command starts that actor. The actor dies when the block ends. A `test` block inside a function stays in that actor and runs only under `test`. | exit 0 when every test actor finishes with no recorded failure. Exit 1 when one dies by `unwrap` or `require`, or finishes with a recorded `check` failure. |
| TEC-16 | `check` and `require` | Both are standard library modules. They are legal only inside a `test` block. The calls are `equal(expected, actual)`, `true(condition)`, `false(condition)`, and `ok(result)`. `check` records a failure and continues. `require` kills the actor. `require.ok` returns the `ok` value. `check.ok` returns `Option[T]`. | a continued turn, a recorded failure, or a dead actor |

## Tooling

| TEC | Tool | Relation | We do not | Cite |
|-----|------|----------|-----------|------|
| TEC-01 | usage-rs | depend | a second argv parser, and the `usage` binary | https://usage.jdx.dev/spec/argv |
| TEC-01 | Rust | implement | a Go compiler for the-lang | path:Cargo.toml |
| TEC-10 | Go module resolution | implement | a version inside the import string | https://go.dev/ref/mod#minimal-version-selection |

| Cell | Pick | C or D | Implements | Cite if C |
|------|------|--------|------------|-----------|
| Language | Rust hosts the compiler. the-lang is the product. | C | TEC-01 | path:Cargo.toml, path:mise.toml |
| Runtime | per-actor GC and a task scheduler | D | TEC-02, TEC-04, TEC-05 | |
| Persistence | none | D | | |
| UI | none | D | | |
| Packaging | `lang.mod` and `lang.lock` | D | TEC-10 | |
| Identity | pid | D | TEC-02 | |
| Host OS | the OS that runs the toolchain | C | TEC-11 | path:mise.toml |

## Terminology

| Concept | Approved | Banned |
|---------|----------|--------|
| toolchain binary | toolchain | compiler driver, CLI tool |
| actor identity | pid | process id, goroutine |
| one actor's turn | turn | thread, goroutine |
| call to an actor | call | send, RPC keyword |
| async handle | promise | future, task |
| success and failure value | Result | exception |
| absence value | Option | null |
| failure payload | error | Fault, exception |
| cause chain | cause | wrap string |
| conversion | as | cast, from |
| size count | usize | size_t, int |
| default signed integer | int | i32 alias, platform int |
| test actor | test actor | agent |
| test failure that continues | check | assert keyword, testing.T |
| test failure that kills | require | assert keyword, FailNow |
| teardown method | __close | __del__, finalizer |
| module requirements file | lang.mod | go.mod, package.json |

## Public contract

The toolchain is one binary. The first argument is the command. The commands are `build`, `run`, `test`, and `generate`.

A flag is one argument. A boolean is `--name`. A short boolean is `-n`. A value is `--name=value`. The same value may also be `--name` followed by the next argument. Boolean short flags cluster. `--` ends flags. Later tokens are positionals. A lone `-` is a positional. Flags may sit before positionals. Flags may sit after positionals until `--`. An unknown flag is a usage error. A missing value is a usage error. The root flags are `--help`, `-h`, and `--version`. Each one is valid on every command.

`--help` and `-h` write usage to stdout and exit 0. `--version` writes the version to stdout and exits 0.

Diagnostics go to stderr. A successful `build` writes no success text to stdout. It writes one file. `build` accepts `--artifact`, `--os`, `--arch`, and `--output`. `--artifact` is `binary` or `c-shared`. Absent `--artifact` means `binary`. `binary` is one executable. `c-shared` is one C shared library. `--os` names the target OS. `--arch` names the target architecture. The host is the default for both. The required OS names are `linux`, `darwin`, and `windows`. The required architecture names are `amd64` and `arm64`. `--output` names the file path. When `--output` is absent, the file name is the base name of the path argument and the file is written in the current directory. `run`, `test`, and `generate` reject those four flags.

`run` uses the program status after the program starts. A toolchain failure before that start exits 1 when the argv was valid. It exits 2 when the argv was bad.

## Types

| Command | Type it mutates | Transition | Bad input |
|---------|-----------------|------------|-----------|
| build | program at path | write one file for `--artifact`, `--os`, and `--arch`. Success exits 0 with empty stdout. | bad argv exits 2. A compile failure exits 1. |
| run | program at path | start the program. `test` blocks are absent. After start, the status is the program status. | bad argv exits 2. A failure before start exits 1. |
| test | program at path | start one actor for each top-level `test` block. `require` kills that actor. `check` records a failure and continues. | bad argv exits 2. A killed test actor or a recorded `check` failure exits 1. |
| generate | program at path | the command exists. The marker grammar is later work. | bad argv exits 2. |

## Invariants

| ID | Predicate | On | Forbidden bypass |
|----|-----------|----|------------------|
| INV-01 | A pid does not keep its actor alive. | pid | a hidden retain |
| INV-02 | A pid is never reused. | pid | recycle the name |
| INV-03 | Two actors cannot both mutate one value. | value | a shared mutable place |
| INV-04 | A promise dies with the actor that created it. | promise | a promise stored in another actor |
| INV-05 | `await` freezes that actor. No other turn of that actor runs. | actor | reenter on await |
| INV-06 | The collector runs no user code. | GC | __del__, __gc |
| INV-07 | `int` and `usize` are distinct types. | integers | implicit store of a computed int into usize |
| INV-08 | `match` names every variant. | enum | a default arm that hides a variant |
| INV-09 | An import string has no version and no scheme. | import | a URL with `@version` in the source |
| INV-12 | An import line has no `*`. | import | `import * from "place"` and `import * as name from "place"` |
| INV-10 | Builtin integer `+`, `-`, and `*` wrap. | integer | a trap on overflow |
| INV-11 | `require` kills the test actor. `check` does not. | test actor | a language `assert` keyword |

## Errors

| Public operation | Bad input | One reaction |
|------------------|-----------|--------------|
| toolchain argv | unknown command, unknown flag, missing value, bad value | exit 2, message on stderr |
| build, test, generate | valid argv and a failed command | exit 1, diagnostics on stderr |
| run | valid argv and the program starts | the program status |
| await | callee is dead | `err(error.dead)` |
| unwrap | `err` | the calling actor dies |
| unwrap | `none` | the calling actor dies |
| require | a failed check inside a `test` block | the test actor dies |
| check | a failed check inside a `test` block | the failure is recorded and the turn continues |
| check, require | a call outside a `test` block | the program is rejected |
| as | no __to__, no __from__, and no underlying-type conversion | the program is rejected |
| __close | returns `err` | the owning actor dies |
| import | unknown specifier | the program is rejected |
| import | a `*` in the import line | the program is rejected |

## Quality

N/A for actors and capabilities: genre=cli.

| Concern | Measure |
|---------|----------------------------------|
| exit contract | Usage is exit 2. A failed command is exit 1. Help and version are exit 0 on stdout. `run` after start uses the program status. |
| untrusted input | The path argument is user input. `build` MUST NOT run the program. `run` and `test` run it only after a valid argv. |

## Security

In scope: bad argv MUST NOT run user code. `generate` runs only the external generators the source names.

Why it cannot happen (if claiming none): this draft does not claim a sandbox.

Residual risk: `run` and `test` execute the user program with the rights of the toolchain process.

## Success

- [ ] A stored pid of a finished actor does not keep that actor alive, and `await` of a call on it yields `err(error.dead)`.
- [ ] `await` of a live call yields `Result[T]`, and `unwrap` of `err` kills the caller.
- [ ] `as[string](color)` runs `Color.__to__` when that method returns `string`.
- [ ] A computed `int` stored in a `usize` place is rejected unless the source writes `as[usize]`.
- [ ] `import { fprintf } from "fmt"` resolves as a standard library path and does not contain a version.
- [ ] `import require from "require"` makes `require.equal` and `require.ok` available, and does not bind `equal` by itself.
- [ ] `import * from "require"` is rejected.
- [ ] `Counter{}` is rejected. `Counter[int]{}` is accepted.
- [ ] `i32` maximum plus 1 wraps to `i32` minimum.
- [ ] `the-lang run` on `examples/01-counter.lang` writes `count: 1` to stderr and exits 0.
- [ ] `the-lang test` on `examples/04-test.lang` exits 0.
- [ ] A top-level `test` block is absent from the `build` output and runs as its own actor under `test`.
- [ ] `require.equal(1, 0)` inside a top-level `test` block kills that actor and `test` exits 1.
- [ ] `check.equal(1, 0)` inside a top-level `test` block lets the block finish, and `test` still exits 1.

## Later work

1. Borrow checking. It MUST NOT change which values are reachable.
2. A watch for collection. It is not a finalizer.
3. A remote call protocol. It MUST NOT change a pid or `as`. Clusters stay out.
4. The `generate` marker. The command exists. This draft does not name the marker.
5. A composed integer type whose operations do not wrap. Builtin integers keep TEC-13.
6. Native code generation. It implements `build` in TEC-14. Until then, `run` interprets `main` and `test` interprets test blocks. `build` exits 1 and reports that codegen is later.

## Assumptions

| ID | Fact | If false |
|----|------|----------|
| AS-01 | Target ABIs are LP64, LLP64, and ILP32, so C `int` is 32 bits. | `int` MUST be redefined before a new ABI. |
| AS-02 | Go minimal version selection stays the version rule this draft copies. | TEC-10 MUST be revised. |

## Decision history

- ADR-0001: Actors replace goroutines. Rejected: shared mutable threads.
- ADR-0002: One `call` on a pid. Rejected: a separate send.
- ADR-0003: The runtime bounds the mailbox. Rejected: a spawn capacity argument.
- ADR-0004: `__close` runs at the end of a place. Rejected: `__del__` in the collector.
- ADR-0005: `as` tries `__to__`, then `__from__`. Rejected: a `__str__` method.
- ADR-0006: `int` is C `int`. `usize` is `size_t`. Rejected: one unsized integer for both.
- ADR-0007: Errors form a cause chain. Rejected: a `Fault` enum.
- ADR-0008: Imports use braces and Go module paths. Rejected: a version inside the import string.
- ADR-0009: Builtin integers wrap. Rejected: a trap on overflow.
- ADR-0010: `build` writes one file for an OS and an architecture. Rejected: a mailbox-style capacity for the output.
- ADR-0014: `--artifact` is `binary` or `c-shared`. Absent means `binary`. Rejected: an artifact flag on `run` or `test`.
- ADR-0011: A top-level `test` block is one short-lived actor. Rejected: a `Test` function name as the only test form.
- ADR-0012: `check` and `require` live in the standard library. Rejected: a language `assert` keyword.
- ADR-0013: argv is a usage spec. Rejected: a copied lewkit parser.

## Examples

Examples are informative. They are not law. See `examples/`.
