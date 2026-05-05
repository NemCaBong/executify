# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Executify is a backend online judge system written in Go. It executes code submissions in isolated sandbox environments (via the `github.com/NemCaBong/go-isolate` library) and supports competitive programming problem evaluation with multiple programming languages.

## Commands

### Running the Application

```bash
# Start infrastructure (MySQL + Redis)
docker-compose up -d

# Run the HTTP API server
go run ./cmd/server/main.go

# Run submit worker
go run ./cmd/worker/main.go submit

# Run run-mode worker
go run ./cmd/worker/main.go run
```

### Building

```bash
go build ./...
```

### Tests

```bash
go test ./...
```

## Architecture

The project follows Clean Architecture with a strict dependency direction: `adapter → application → domain`.

```
cmd/
├── server/       # HTTP API entry point
└── worker/       # Worker process entry points (submit, run modes)
internal/
├── domain/       # Core entities: Submission, Problem, Language, CodeRunner
├── application/  # Use cases + repository/queue interfaces
│   ├── submission/
│   ├── problem/
│   ├── queue/
│   └── worker/
├── adapter/      # Concrete implementations
│   ├── http/     # Gin handlers, request/response DTOs
│   ├── repository/ # GORM-based MySQL repositories + DB entity models
│   ├── queue/redis/ # Redis-backed message queue producer
│   └── worker/   # Submit and run worker implementations
└── config/       # Environment-based configuration
```

### Data Flow

**Submit flow**: `POST /api/v1/submissions` → handler → submission usecase → save to DB (SUBMITTED) → enqueue to `submit_queue` → worker BLPop → CodeRunner executes in sandbox → update DB with results.

**Run flow**: Same path but uses `POST /api/v1/submissions/run`, `run_queue`, and user-provided stdin instead of problem input file.

### Key Design Points

- **CodeRunner** (`internal/domain/code_runner.go`): Manages the sandbox lifecycle (init → compile → execute → cleanup). Box IDs are derived from submission IDs using Mersenne prime modulus (`(2^31 - 1)`).
- **Wrapper code**: Problems can define a `wrapper_code` with a `{{.}}` placeholder where user code is injected before execution.
- **Workers**: Use Redis `BLPop` with 0 timeout. Worker count is configurable via `RUN_WORKER_COUNT` / `SUBMIT_WORKER_COUNT` env vars. On shutdown, in-flight executions finish before the process exits.
- **Repository entities** (`internal/adapter/repository/entity/`): Separate DB models with `.ToDomain()` mappers to keep GORM out of the domain layer.

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/submissions` | Submit code for evaluation against expected output |
| `POST` | `/api/v1/submissions/run` | Run code with custom stdin |
| `GET`  | `/api/v1/submissions/:id` | Get submission status and results |
| `PUT`  | `/api/v1/problems` | Upsert a problem |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `SUBMIT_WORKER_COUNT` | `1` | Submit worker goroutine count |
| `RUN_WORKER_COUNT` | `1` | Run worker goroutine count |
| `MYSQL_HOST/PORT/USER/PASSWORD/DATABASE` | — | MySQL connection |
| `REDIS_ADDRESS` | `localhost:6379` | Redis address |
| `REDIS_SUBMIT_QUEUE` | `executify:queue:submit` | Submit queue key |
| `REDIS_RUN_QUEUE` | `executify:queue:run` | Run queue key |

## Database

Migrations are in `migrations/`. Three tables: `languages`, `problems`, `submissions`.

`problems` stores `input_file` and `expected_output_file` paths (filesystem paths accessible to the worker), plus resource limits (CPU time, wall time, memory, stack, process count).
