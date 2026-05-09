# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run-server          # start server (port 3223)
make run-client          # start interactive client
make test                # run all tests
make test-cover          # tests with coverage report
make install-deps        # install minimock into ./bin/

go test -race ./...                            # tests with race detector
go test ./internal/database/... -run=TestName  # single test
go generate ./mocks/...                        # regenerate mocks

go run ./cmd/server/... --config=server.config.yaml  # server with custom config
go run ./cmd/client/... --address=host:port          # client to non-default address
```

Server logs are written to the file specified in `logging.output` (`app.log` by default).

## Architecture

The server is assembled in `cmd/server/main.go` from four independent layers:

```
TCP Client
    └── network/tcp_server  — accepts connections, enforces limit via semaphore channel,
        │                     handles PING/PONG before passing to DB, rejects oversized messages
        └── database.Database.HandleQuery
                ├── compute/parser  — parses string into Query {CommandID, Arguments}
                └── storage.Storage — generates txID, injects into context, delegates to engine
                        └── storage/engine/in_memory  — map + RWMutex behind Hashtable interface
```

**Key interfaces** (for test substitution):
- `database.computeLayer` / `database.storageLayer` — internal interfaces of `Database`
- `storage.Engine` — implemented by `InMemoryEngine`; `storage.Hashtable` — implemented by the built-in hash table
- `network.TCPServer` — implemented by `tcp_server.server`

**Mocks** are generated via `minimock` (`./bin/minimock`) and live in `./mocks/`. Directives are in `mocks/generate.go`.

**Context** carries `txID` (transaction ID) between layers via `internal/contextx/txIDctx`. Each `storage` call generates a unique ID with an atomic counter.

**Config** (`internal/config/app`): YAML with auto-defaults. Missing file starts silently with defaults. `MaxMessageSizeBytes` and `IdleTimeoutDuration` are parsed from string fields on load and are not serialized (yaml:"-").

**Connection limiting**: `semaphore` is a buffered channel of size `max_connections`. On overflow the connection is closed immediately; the client detects rejection via PING on startup.

## Working guidelines

### Commits
- One logical change per commit. No `Co-Authored-By` lines.
- Format: `type(scope): short description` — e.g. `feat(metadata): add ListObjects handler`.

### Tests
- Every feature or bug fix must include tests or an explicit offer to add them.
- Go: table-driven with `testify`, mocks via `minimock`.
- Integration tests requiring external infra must be skippable without it.

### Branches
- Format: `type/scope-description` — e.g. `feat/authz-cache-invalidation`, `fix/users-uuid-parsing`.

### Issues
- Title: specific and actionable — what's broken or what needs to be done.
- Body: `## What` (2-3 sentences) + `## Acceptance criteria` (checklist).
- For bugs: include reproduction steps in the What section.

### Pull requests
- Before creating a PR, check the current branch name. If it matches `claude/*` (auto-generated worktree branch), create a properly named branch first: `git checkout -b type/scope-description`.
- Title format matches commits: `type(scope): short description`.
- Link to issue via `closes #N` in PR body when applicable.
- Descriptions: concise, no filler.

### GitHub (gh CLI)
- Before any write action (review, comment, close, merge) — show draft and wait for confirmation.
- Before reviewing a PR — always run `gh pr diff` + `gh pr view` to understand context.
- Don't guess the intent of changes — ask if unclear.

### Code style
- No comments explaining WHAT — only WHY when non-obvious.
- No speculative abstractions. Implement only what the task requires.
- Error handling only at system boundaries (user input, external calls).
