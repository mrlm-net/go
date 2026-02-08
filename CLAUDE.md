# github.com/mrlm-net/go

Go utility library — small, focused packages solving common problems not covered by the standard library. Deepmerge, radix tree routing, console tooling, and more.

## Stack

- **Language:** Go 1.25+
- **Testing:** `go test` + `github.com/stretchr/testify`
- **No frameworks** — pure standard library + minimal deps

## Plugin Directive

All development MUST use `mrlm-xyz:mrlm` agents, skills, and commands.

### Commands

| Command | Description |
|---------|-------------|
| `/mrlm:init` | Bootstrap or augment project structure and CLAUDE.md |
| `/mrlm:plan` | Plan tasks, write backlog items, create implementation plan |
| `/mrlm:analyse` | Analyse task and create implementation plan |
| `/mrlm:make` | Execute full SDLC via delivery manager orchestration |
| `/mrlm:review` | Code review, testing, and validation |
| `/mrlm:ask` | Quick questions without side effects (read-only) |
| `/mrlm:write` | Articles, documentation, marketing content |

### Agents

| Agent | Responsibility |
|-------|----------------|
| `delivery-manager` | SDLC orchestration, plan ownership, agent coordination |
| `business-analyst` | Requirements, stakeholder analysis, backlog writing |
| `software-architect` | System design, interfaces, ADRs |
| `software-engineer` | Code implementation (Go, TypeScript, Svelte, Wails) |
| `platform-engineer` | Infrastructure, CI/CD, deployment |
| `qa-engineer` | E2E testing, performance, security validation |
| `personal-writer` | Documentation, blog posts, marketing materials |

## Project Structure

```
pkg/                    # Public packages (importable API)
  data/
    deepmerge/          # Generic deep merge for maps, slices, structs
    radix/              # Radix tree data structure
  console/              # CLI application framework
    pipeline/           # Middleware pipeline for command processing
  routing/              # HTTP router built on radix tree
examples/               # Usage examples per package
```

## Commands

```bash
go fmt ./...              # Format
go vet ./...              # Lint
go test ./...             # Test all
go test -bench=. ./...    # Benchmarks
go test -cover ./...      # Coverage
go build -o bin/ ./examples/...  # Build examples to bin/
```

## Build Rules

- **All `go build` output MUST go to `bin/`** — use `go build -o bin/ ./...` or `go build -o bin/<name> ./examples/<name>`
- The `bin/` directory is gitignored — never commit compiled binaries
- Never run `go build` without `-o bin/` to avoid leaving binaries in the repo root

## Conventions

- Packages under `pkg/` — each self-contained with own tests
- Examples under `examples/<pkg>/main.go`
- Follow Go Proverbs: clear > clever, small interfaces, errors are values
- Zero dependencies beyond testify for tests
- Table-driven tests, benchmarks for performance-sensitive code
- Export only what's necessary; keep API surface minimal

## Workflow

1. `/mrlm:init` — bootstrap or augment project
2. `/mrlm:plan` — plan tasks, create backlog and implementation plan
3. `/mrlm:make` — implement via delivery manager orchestration
4. `/mrlm:review` — review and validate
5. `/mrlm:ask` — quick questions (read-only)
6. `/mrlm:write` — generate documentation or content
