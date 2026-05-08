# Executify Roadmap — From Learning Project to Hostable Showcase

> Goal: turn the current Clean-Architecture skeleton into a production-grade,
> hostable code-execution platform that you can confidently demo as a portfolio
> piece — comparable in feel (not yet scale) to Judge0 / LeetCode / HackerRank.

This document is structured in three layers:

1. **Critical fixes** — security and data-integrity issues that should be addressed before *anyone* hits the deployed system.
2. **Phased roadmap** — eight phases, each leaving the system more demoable than the last.
3. **Reference appendix** — cross-cutting concerns (testing, observability, deployment, costs) that touch multiple phases.

Each item ends with a concrete `Action:` so you can convert this directly into tickets.

---

## TL;DR — what's blocking "showcase-ready" today

The current codebase is a clean, working MVP for a single trusted user submitting Python to one problem. It is **not** ready to host because:

| Blocker | Where | Impact |
|---|---|---|
| `Dockerfile` is 0 bytes | `Dockerfile` | Cannot build a container image at all |
| `configs/config.yaml` is 0 bytes; YAML is never read | `internal/config/config.go` | Config story is misleading |
| Submissions have no `user_id` | `internal/domain/submission.go`, `migrations/000001` | No ownership, no leaderboard, no per-user history |
| `PUT /api/v1/problems` is open to any authenticated user | `cmd/server/main.go:122` | Trivial RCE: any user can plant `wrapper_code` and `input_file` paths |
| Worker drops in-flight jobs on SIGINT | `internal/adapter/worker/{run,submit}_worker.go` | Submissions stuck in `PROCESSING` forever after restart |
| No DLQ / retry / idempotency on the queue | `internal/adapter/queue/redis/producer.go` | Silent job loss on crash |
| Verdict is binary `COMPLETED/FAILED` | `internal/domain/submission.go:8-13` | Cannot distinguish WA / TLE / MLE / RE / CE |
| 11 hash-named `*meta.txt` files committed in repo root | sandbox artifact leak | `code_runner.go:56` writes meta to CWD |
| Default JWT secret `"change-me-in-production"` | `internal/config/config.go:43` | Forge-any-token if env unset |
| No tests, no CI, no lint config | repo-wide | Showcase quality story is empty |

These are addressed across Phases 0–2 below.

---

## Phase 0 — Critical security & data-integrity fixes (must do before any public deploy)

These do not deliver new features, but each one is a vulnerability or footgun.

### 0.1 Add roles and lock down problem upsert

`User` has no `Role` (see `internal/domain/user.go`). `PUT /api/v1/problems` is gated by JWT only (`cmd/server/main.go:122`). Any registered user can:
- set `wrapper_code` to malicious Python that exfiltrates env vars / opens reverse shell when the next victim submits to that problem,
- set `input_file: "/etc/passwd"` (read by `os.ReadFile` in `code_runner.go:248`).

**Action:**
- Add `users.role ENUM('user','admin') NOT NULL DEFAULT 'user'` (migration 000006).
- Add `Role` field to `domain.User` and the entity, propagate through JWT claims.
- New middleware `RequireRole("admin")` in `internal/adapter/http/middleware/`.
- Apply it to `PUT /problems`, `DELETE /problems`, future `POST /languages`, etc.
- Whitelist allowed `input_file`/`expected_output_file` paths to a fixed root (e.g. `/var/lib/executify/problems/<id>/`) — reject any path containing `..` or not under that root.

### 0.2 Couple submissions to users

`submissions` table has no `user_id` (`migrations/000001_init_projects.up.sql`). Yet auth middleware exists and stuffs user info in context.

**Action:**
- Migration 000007 adds `user_id INT NOT NULL`, FK to `users(id)`, index on `(user_id, created_at DESC)`.
- Middleware sets `userID` in gin context (already does in `internal/adapter/http/middleware/auth.go`).
- `SubmissionHandler.Submit` and `Run` must read it and pass to usecase.
- `GET /submissions/:id` must verify ownership (or admin role) before responding.

### 0.3 Fail-fast on missing secrets

`internal/config/config.go:43` silently defaults `JWT_SECRET` to `"change-me-in-production"`. `FILE_SECRET` defaults to `""` which collapses the FNV hash into something attackers can replay (see 0.5).

**Action:**
- In `Load()`, if `APP_ENV != "dev"` and `JWT_SECRET` is empty/default, `log.Fatal` on startup.
- Same for `FILE_SECRET`. Provide a `make gen-secrets` target that prints `openssl rand -hex 32` strings.
- Add `.env.example` checked into the repo.

### 0.4 Eliminate shell injection in `runOnce`

`code_runner.go:153` builds a bash command with `fmt.Sprintf("%s <<< %q 3>%q", runCmd, stdin, fname)`. `%q` is *Go* quoting, not bash-safe. The here-string carries `inputLine` straight into bash.

**Action:**
- Stop using `<<<`. Pipe stdin through the isolate executor's API instead (see `isolate.WithStdinFile(...)` or write the test-case content to a file in the sandbox first, then redirect with `<filename`).
- For fd3, prefer a wrapper that opens the path with `os.fdopen(...)` from Python rather than relying on bash redirection.
- Replace shell-quoting helpers with `shellescape` library or manual single-quote-escape.

### 0.5 Cryptographic filename hashes

`internal/domain/utils.go` uses FNV-64 with a per-process secret. FNV is non-cryptographic; with a known/empty secret, an attacker can predict box paths and race files.

**Action:**
- Switch to `hmac.New(sha256.New, []byte(secret))` and hex-encode 16 bytes.
- Generate `FILE_SECRET` per-deployment (not per-restart) and load via secrets manager.

### 0.6 Don't write sandbox metadata to CWD

`code_runner.go:56-58` does `Meta(r.metaFileName)` with a bare relative filename. The executor writes that file relative to the worker process CWD — which is the repo root when running `go run .`. Result: 11 `*meta.txt` files committed to git already.

**Action:**
- Create a per-execution temp dir: `dir, _ := os.MkdirTemp("", fmt.Sprintf("executify-%d-", subID))`.
- Pass absolute paths into `Meta()`, `Stdout()`, `Stderr()` (or use the in-sandbox files via `WriteToSandbox` only).
- `defer os.RemoveAll(dir)` after `Cleanup`.
- Add `*meta.txt`, `*stdout.txt`, `*stderr.txt`, `*actual_output.txt`, `.idea/` to `.gitignore`.
- Delete the 11 leftover files now.

### 0.7 Fix box-ID collisions

`code_runner.go:47` uses `submission.ID % BoxModulus` with default `BoxModulus = 65535`. After 65 535 submissions you get collisions; concurrent workers can land on the same isolate box.

**Action:**
- Maintain a small pool of box IDs (e.g. 0..N-1 where N = `SUBMIT_WORKER_COUNT + RUN_WORKER_COUNT`).
- Acquire/release with a Redis `SETNX` on `executify:box:<id>` (TTL = wall_time_limit + 30s safety).
- Or: assign each worker goroutine a fixed box ID at startup (1 box per goroutine — simple and race-free).

### 0.8 Bound source-code, input, and output sizes

No limits exist. A 50 MB submission body OOMs the worker; a program printing 1 GB OOMs `os.ReadFile` in `getActualOutput()`.

**Action:**
- Reject `source_code` > 64 KB at the request DTO (`request/submission_request.go`).
- In the wrapper, read at most N bytes from fd3 (`os.read` loop with cap).
- Truncate stored `actual_output` / `stderr` to e.g. 16 KB with a `(truncated)` marker.

### 0.9 Sanitize handler error responses

Handlers return `gin.H{"error": err.Error()}` directly (`submission_handler.go:46/51/96`, similar in others). GORM/Redis/internal errors leak to clients.

**Action:**
- Define a `pkg/httperr` package with `BadRequest / Unauthorized / Forbidden / NotFound / Internal` helpers that return `{ "code": "...", "message": "..." }`.
- Internal errors get a generic `"internal_error"` message; the real error goes to logs with a request ID.

### 0.10 Rate limit auth and submit endpoints

No rate limiting anywhere. `POST /auth/login` runs bcrypt server-side per request — perfect for cost-amplification attacks.

**Action:**
- Use `github.com/ulule/limiter/v3` with Redis store.
- `/auth/login` and `/auth/register`: 10 req / minute / IP.
- `/submissions` and `/submissions/run`: 30 req / minute / user (and 5 / second / user).
- Document headers `X-RateLimit-*`.

---

## Phase 1 — Make it shippable (Dockerfile, config, CI)

The goal of this phase is "I can `git clone && make up` on a fresh machine and have a working stack."

### 1.1 Real Dockerfile

The current `Dockerfile` is empty.

**Action:** multi-stage build, distroless final image, non-root user, static binary.

```Dockerfile
# build stage
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

# runtime stage — needs isolate binary on host, so workers run on host or in privileged image
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
COPY --from=build /out/worker /worker
USER nonroot:nonroot
ENTRYPOINT ["/server"]
```

Note: workers need `isolate` (kernel cgroups) — they cannot run in distroless. Build a separate `Dockerfile.worker` based on `ubuntu:24.04` with isolate installed and run as a privileged sidecar.

### 1.2 Compose: add the app, healthchecks, .env

`docker-compose.yml` only defines `db` and `redis`. The mysql password is hardcoded and 8.0.22 is old.

**Action:**
- Add `executify-server` and `executify-worker-{submit,run}` services.
- Use `${MYSQL_PASSWORD:?error}` to require .env.
- Add `healthcheck:` for mysql (`mysqladmin ping`) and redis (`redis-cli ping`).
- Bump MySQL to `8.0.39` or `mysql:8.4`.
- Add `migrate` one-shot service that runs `golang-migrate up` before workers start.
- Volume-mount `/var/local/lib/isolate` for sandbox state.

### 1.3 Real config layer

`configs/config.yaml` is empty; `viper` is imported but only used for unused defaults. Pure `os.Getenv` is fine but advertise it honestly.

**Option A (recommended for now):** delete `configs/config.yaml` and the `viper` import; document env-only.

**Option B:** wire up viper properly — `viper.SetConfigFile("configs/config.yaml")`, `viper.AutomaticEnv()`, `viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))`. This is overkill until you have many environments.

### 1.4 Makefile / Justfile

No build orchestration exists.

**Action:** create a `Makefile`:

```make
.PHONY: build test lint up down migrate seed

build:
	go build ./...

test:
	go test -race -cover ./...

lint:
	golangci-lint run

up:
	docker compose up -d --build

migrate:
	migrate -path migrations -database "$(DATABASE_URL)" up

seed:
	go run ./cmd/seed
```

### 1.5 GitHub Actions CI

No `.github/`. Add `ci.yml`:

- Job `test`: setup Go, run `go vet`, `golangci-lint`, `go test -race -cover`.
- Job `build`: build server and worker images, push to GHCR on tag.
- Job `migrate-check`: spin up MySQL service container, run all up migrations then all down migrations to catch breakage early (requires writing the missing `*.down.sql` files).

### 1.6 Tests baseline

No `*_test.go` files exist. You don't need 80% coverage on day one but you should have:

- Unit tests on `domain/utils.go` (hash determinism, no collision over 10k IDs).
- Unit tests on `domain.splitLines`.
- Table-driven tests on the submission status state machine.
- Repository tests using `testcontainers-go` (MySQL + Redis ephemeral containers).
- HTTP handler tests via `httptest` + `gin.SetMode(gin.TestMode)`.
- One end-to-end test: submit a Python solution, poll, assert COMPLETED.

---

## Phase 2 — Verdict model, granular results, and the missing API surface

This phase makes the platform feel like a real judge.

### 2.1 Rich verdict enum

`internal/domain/submission.go:8-13` only has `COMPLETED / FAILED / PROCESSING / SUBMITTED`. Real judges distinguish:

```
QUEUED, COMPILING, RUNNING,
ACCEPTED,                          // all cases passed
WRONG_ANSWER,                      // ran but output mismatch
TIME_LIMIT_EXCEEDED,
MEMORY_LIMIT_EXCEEDED,
RUNTIME_ERROR,                     // non-zero exit, signal
COMPILATION_ERROR,
INTERNAL_ERROR                     // judge-side fault
```

`code_runner.go:189` already has a `resultSucceeded` helper that reads `result.Meta.IsSuccess()` — extend it to return a *typed* verdict by inspecting `result.Meta.Status` ("TO" → TLE, "SG" → RE, "ML" → MLE, etc. — see the isolate library docs).

**Action:**
- Migration 000008 widens `submissions.status` to allow new values, adds CHECK constraint.
- Add `Verdict` type with a `Classify(meta)` constructor.
- `executeSubmit` and `executeRun` set the typed verdict.
- `compile()` returning non-zero exit becomes `COMPILATION_ERROR` (today it's swallowed as a Go error — `code_runner.go:111`).

### 2.2 Per-test-case results

Today: one input file, one expected file, two text columns on `submissions`. Real judges store each test case's verdict.

**Action:**
- New table `test_cases (id, problem_id, idx, input_text, expected_output_text, is_sample, weight, hidden)`.
- New table `submission_test_results (id, submission_id, test_case_id, verdict, time_used_ms, memory_used_kb, actual_output_truncated)`.
- Migrate existing `problems.input_file`/`expected_output_file` reading code into a one-time importer.
- `GET /submissions/:id` returns sample-case details only; hidden cases collapse to verdict + counts.
- Enable partial scoring: total = sum(weight where verdict == ACCEPTED).

### 2.3 The missing endpoints

Needed for any real frontend:

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/me` | Current user |
| PATCH | `/api/v1/me` | Update profile |
| POST | `/api/v1/auth/password-reset` | Email link flow |
| GET | `/api/v1/problems` | Paginated, filterable by tag/difficulty/status |
| GET | `/api/v1/problems/:slug` | Public detail (samples only) |
| DELETE | `/api/v1/problems/:id` | Admin |
| GET | `/api/v1/problems/:id/submissions` | This problem's submissions |
| GET | `/api/v1/submissions` | My submissions, paginated |
| GET | `/api/v1/languages` | Supported languages with versions |
| GET | `/api/v1/tags` | All tags |
| GET | `/api/v1/leaderboard` | Top users by accepted-count |
| GET | `/api/v1/healthz` | Liveness |
| GET | `/api/v1/readyz` | Readiness (db + redis ping) |
| GET | `/metrics` | Prometheus |

**Action per endpoint:**
- DTO in `request/`, response in `response/`, handler thin, usecase does the work.
- Standardize a `PaginatedResponse[T]{ Items []T, Total int, Page, PageSize int }`.
- Add `?cursor=` based pagination for submissions (more performant than offset).

### 2.4 Standardized error envelope

Today every handler invents its own `gin.H{"error": "..."}`. Pick one shape and stick to it:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "source_code is required",
    "request_id": "01HZX..."
  }
}
```

`code` should be a stable enum that the frontend i18n's against.

### 2.5 Add slugs and richer problem metadata

Add to `problems`:
- `slug VARCHAR(100) UNIQUE` for clean URLs (`/problems/two-sum`)
- `difficulty ENUM('easy','medium','hard')`
- `is_public BOOLEAN DEFAULT TRUE`
- `author_id INT REFERENCES users(id)`
- `accepted_count INT DEFAULT 0`, `submission_count INT DEFAULT 0` (denormalized; updated via worker)
- `editorial_markdown TEXT`
- `constraints TEXT`

---

## Phase 3 — Operations, observability, and queue resilience

If you want to run this 24/7 without babysitting it, these have to be solid.

### 3.1 Replace plain Redis lists with a real queue protocol

Today: `LPUSH` from producer, `BLPOP` from worker (`internal/adapter/queue/redis/producer.go`, `internal/adapter/worker/submit_worker.go:49`). No ack, no retry, no DLQ. Crash-after-pop = silent loss.

**Action:** pick one:

- **Asynq** (`github.com/hibiken/asynq`) — Redis-backed, has retry, DLQ, scheduled tasks, web UI. Idiomatic Go. **Recommended.**
- **River** (`github.com/riverqueue/river`) — Postgres-backed; if you migrate from MySQL to Postgres later, this becomes interesting.
- **NATS JetStream** — overkill but very production-y if you go multi-host.

Whichever you pick, the important properties:
- Visibility timeout (in-flight job is invisible to other workers but reappears if not acked in time).
- Bounded retries with exponential backoff.
- DLQ for poison messages.
- Idempotency key = `submission_id`; reject duplicate enqueues.

### 3.2 Graceful worker shutdown that actually drains

`run_worker.go:62` and `submit_worker.go:62` use `context.Background()` for the per-job execution to avoid SIGINT killing the in-flight run — but the goroutine then exits at the top of the next loop iteration without waiting on the running goroutine.

**Action:**
- Track in-flight jobs in a `sync.WaitGroup` *separate* from the main one.
- On shutdown signal: stop accepting new BLPOP, wait for in-flight WG with timeout (e.g. wall-time limit + 10s).
- Force-kill remaining isolate sandboxes after grace period.

### 3.3 Structured logging

Today: stdlib `log` with text format, copy-paste log strings (run_worker.go:89 says "Submit worker" — copy-paste bug from submit_worker.go).

**Action:**
- Switch to `log/slog` (stdlib in Go 1.21+).
- JSON handler in production, text in dev.
- Always include `submission_id`, `user_id`, `request_id`, `box_id`, `verdict`.
- One init function: `func NewLogger(env string) *slog.Logger`.

### 3.4 Tracing & metrics

Nothing today.

**Action:**
- Add OpenTelemetry SDK with OTLP exporter (vendor-neutral).
- Trace spans: `http.request` → `usecase.submit` → `queue.enqueue`; `worker.poll` → `worker.handle` → `code_runner.execute` → `code_runner.{init,compile,run,cleanup}`.
- Prometheus metrics:
  - `executify_submissions_total{verdict,language}`
  - `executify_submission_duration_seconds{verdict}` histogram
  - `executify_queue_depth{queue}` gauge (sample via `LLEN` every N seconds)
  - `executify_worker_in_flight{queue}` gauge
  - `executify_compile_failures_total{language}`
- Grafana dashboard JSON committed under `deploy/grafana/`.

### 3.5 Health and readiness probes

Add `/healthz` (always 200 if the binary is alive) and `/readyz` (checks DB ping + Redis ping; 503 if either is down). Use these as Kubernetes / Compose healthchecks.

### 3.6 De-duplicate the two workers

`run_worker.go` and `submit_worker.go` are 95% identical (audit confirmed). Extract:

```go
type Mode int
const (ModeSubmit Mode = iota; ModeRun)

func handle(ctx context.Context, m Mode, deps Deps, msg queue.Message) error { ... }
```

Single `worker.go` with `mode` switch on construction. Removes copy-paste log bugs and halves maintenance cost.

### 3.7 Database hardening

From the audit:
- No FKs (commented out in 000001:49-50)
- No indexes on `(user_id, created_at)`, `(problem_id, status)`
- No `down.sql` files
- `submissions.type` is an orphan column

**Action:**
- New migration that adds FKs (after backfilling `user_id`).
- Add the missing indexes.
- Write `down.sql` for all 5 existing migrations and every new one.
- Drop the unused `type` column.
- Add CHECK constraints on `status` enums.

---

## Phase 4 — Multi-language sandbox

Today only Python is exercised. Make it pluralistic.

### 4.1 Language registry redesign

`internal/domain/language.go` has `id, name, compile_cmd, run_cmd, source_file`. That's brittle. Real judges store:

- `name` (display): "Python", `version`: "3.11"
- `slug`: "python311"
- `source_file`: "main.py"
- `compile_cmd`: optional, with placeholders `{src}` `{out}`
- `run_cmd`: with placeholder `{exe}` or `{src}`
- `default_time_multiplier`: e.g. 3.0 for interpreted languages
- `default_memory_multiplier`: 1.5 for JVM/CLR
- `runtime_image`: optional Docker tag if you go to per-language sidecar containers
- `is_active` flag

Languages to support for showcase: **Python 3.11, C 17 (gcc), C++ 20 (g++), Java 17, Go 1.22, Rust 1.75, JavaScript (Node 20), TypeScript (ts-node), Kotlin, C#**.

### 4.2 Wrapper-code per language

The current `wrapper_code` on `problems` is Python-specific. Multi-language requires either:

- **Per-(problem, language) wrapper template** — best UX for problem authors, more setup.
- **Per-language harness with a standardized `solve()` ABI** the user fills in — what you already have, but generalized.

For C++ the harness might look like:

```cpp
#include <bits/stdc++.h>
using namespace std;

string solve(const string& line);

{{.}}                                              // user code, must define solve(...)

int main() {
    string line; getline(cin, line);
    string result = solve(line);
    int fd3_dup = dup(3); close(3);
    write(fd3_dup, result.c_str(), result.size());
    close(fd3_dup);
    return 0;
}
```

Schema: `problem_language_templates(problem_id, language_id, wrapper_code, template_code)`.

### 4.3 Compile-cache

Compiled binaries for the same source can be cached by `sha256(source_code + language_id + flags)` in Redis. Saves time on resubmit-after-WA.

---

## Phase 5 — Product features (contests, leaderboards, social)

This is where it stops feeling like a homework project.

### 5.1 Contests

- `contests(id, slug, title, description_md, start_at, end_at, scoring_kind ENUM('icpc','ioi','codeforces'), visibility)`
- `contest_problems(contest_id, problem_id, order, points)`
- `contest_participants(contest_id, user_id, joined_at)`
- During a contest, submissions for a contest problem are scored under the contest's rules, frozen leaderboard last 30 minutes, etc.

### 5.2 Leaderboards & profiles

- `user_stats(user_id, accepted_count, submission_count, easy_count, medium_count, hard_count, current_streak_days, longest_streak_days)` — updated by a worker on each verdict.
- Public profile page: `/u/<username>` with submission heatmap, language breakdown, accepted problems.
- Global leaderboard with weekly/monthly windows.

### 5.3 Discussions and editorials

- `discussions(id, problem_id, user_id, title, body_md, created_at)`
- `comments(id, discussion_id, user_id, body_md, parent_id, created_at)`
- Editorials are admin-authored discussions tagged `editorial`.
- Markdown rendering server-side (gomarkdown) with an HTML sanitizer (bluemonday).

### 5.4 Notifications

WebSocket or SSE channel for real-time submission verdict push (instead of polling `GET /submissions/:id`). Use `gorilla/websocket` or `r3labs/sse`. Channel auth via short-lived JWT.

---

## Phase 6 — Frontend

You already have an empty `web/public/index.html`. Pick one path:

- **Decoupled SPA** (recommended): Next.js 14 (app router) under `/web` directory, deployed to Vercel / static-hosted. Backend stays an API. Pros: fast iteration, modern stack, easy to showcase.
- **Server-rendered**: Templ / a-h/templ + HTMX. Single binary deploy. Less impressive for portfolio, but charming.

Either way, the frontend needs:
- Login / Register / Profile
- Problem list with filters (difficulty, tags, status)
- Problem detail with markdown statement, samples, code editor (Monaco), language picker, submit/run buttons
- Submission detail with per-test-case verdict, runtime, memory, expandable diff for sample-case WA
- Leaderboard
- Contest dashboard
- Admin: problem CRUD with file upload for hidden test cases

---

## Phase 7 — Performance & scale

Once feature-complete, make it fast.

### 7.1 Horizontal scaling

- Server: stateless behind a load balancer; sticky sessions not required (JWT).
- Workers: scale independently. Each worker = N goroutines = N isolate boxes. Cap goroutines = #cpu so isolate doesn't thrash.
- Use a *separate worker pool per language* if some languages are heavier (Java/Rust).

### 7.2 Database

- Read replicas for `GET /submissions`, `GET /problems`. Writes go to primary.
- Add Redis caching for hot reads:
  - Problem detail: `cache:problem:<slug>` TTL 5 min, invalidate on upsert.
  - Language list: `cache:languages` TTL 1 hour.
  - Leaderboard top-100: rebuild every 60s by a worker.

### 7.3 Storage

Keep test-case files in S3-compatible storage (MinIO locally), not on the worker FS. Stream them to the sandbox via `WriteToSandbox`.

### 7.4 Resource isolation

- Run workers in a CPU-pinned, memory-limited cgroup separate from the API.
- One isolate box per CPU core, max.
- Reject new submissions when queue depth > threshold (HTTP 503 with `Retry-After`).

---

## Phase 8 — Polish for showcase

These are the "the recruiter is reading this" touches.

### 8.1 OpenAPI 3.1 spec

Generate from code with `go-swagger` or `huma`. Serve at `/api/v1/openapi.json` and `/docs` (Swagger UI). A live API explorer in the demo is *the* feature recruiters click on.

### 8.2 Real README and architecture diagram

Replace the empty `Dockerfile` and bare CLAUDE.md story with:
- README.md with quickstart, screenshots, a diagram (Excalidraw → SVG checked in)
- ARCHITECTURE.md explaining clean-architecture choices, queue model, sandbox model
- CONTRIBUTING.md with dev setup
- Optional: a 60-second loom video link

### 8.3 Demo data seed

`make seed`: creates 30 problems across 4 difficulty levels, 5 tags, 1 admin + 5 user accounts with sample submissions, and a running contest. Reset DB → seed → live demo in 30 seconds.

### 8.4 Status page

Public `/status` page (or use https://upptime.js.org/ via GitHub Pages) showing API uptime and queue depth.

### 8.5 Production deployment recipe

Pick one and write it up:
- **Fly.io**: cheapest for showcase, supports Docker, easy persistent volumes for isolate.
- **Railway / Render**: very simple, decent free tier.
- **Single VPS (Hetzner CX22) + Caddy**: most control, ~€5/month, isolate works natively, write a `deploy/install.sh` that sets it up from scratch.

For showcase value, the VPS path is the most impressive because you can say "I run my own bare-metal judge."

---

## Reference Appendix

### A. Test pyramid

| Layer | Tool | Examples |
|---|---|---|
| Unit | stdlib `testing` | hash determinism, splitLines edge cases, status state machine |
| Repo | testcontainers-go | mysql FK behavior, GORM mapper round-trips |
| Usecase | mocks (testify/mock or hand-written) | Submit usecase enqueues exactly once |
| HTTP | httptest | route auth, pagination edges, error envelope |
| Integration | docker compose + http calls | login → submit → poll → COMPLETED |
| Load | k6 / vegeta | sustained 100 RPS submissions for 5 min |
| Security | gosec, govulncheck | run in CI |

### B. Suggested file layout additions

```
deploy/
  k8s/                 # manifests for self-hosted k3s
  grafana/             # dashboards JSON
  prometheus/          # scrape config
  install.sh           # bare VPS bootstrap

pkg/
  httperr/             # standardized error envelope
  pagination/          # cursor + offset helpers
  ratelimit/           # middleware
  observability/       # slog, otel, prom registry

web/                   # Next.js frontend (separate go.mod or separate repo)

docs/
  ARCHITECTURE.md
  CONTRIBUTING.md
  RUNBOOK.md           # incident response

scripts/
  gen-openapi.sh
  load-test.sh
```

### C. Suggested cleanup of dead code (free wins)

- `internal/adapter/queue/redis/producer.go:21` `Publish` (LPUSH variant) — never called.
- `cmd/worker/main.go:19-23` `viper.SetDefault` block — defaults are read elsewhere.
- `internal/config/code_runner_config.go:14-18` recomputed `submitWorkerCount` not in returned struct.
- `internal/adapter/http/dto/` — empty dir.
- `submissions.type` column — orphaned.
- `cmd/worker/main.go:41` `os.Exit(1)` after `log.Fatal` — unreachable.
- 11 leftover `*meta.txt` files in repo root.
- `.idea/` directory committed.

### D. Cost estimate for a public showcase

| Item | Provider | Monthly |
|---|---|---|
| VPS (4 vCPU, 8 GB) | Hetzner CX32 | €7 |
| Domain | Namecheap | $1 |
| Backups (S3 1 GB) | Backblaze B2 | $0 |
| Email (transactional, 10k/mo) | Resend | $0 (free tier) |
| Monitoring | Grafana Cloud | $0 (free tier) |
| **Total** | | **~€8 / $9** |

Plus: free tiers of Cloudflare for DNS+CDN+rate-limit, GitHub Actions for CI, GHCR for images.

### E. Suggested sequencing (pragmatic)

If you only have a few weekends, this order maximises showcase ROI per hour:

1. **Weekend 1** — Phase 0.1, 0.2, 0.6, 0.7, 0.10 (security + ownership)
2. **Weekend 2** — Phase 1 entirely (Dockerfile, compose, Makefile, CI skeleton)
3. **Weekend 3** — Phase 2.1, 2.2, 2.3 (verdicts, test cases, missing endpoints)
4. **Weekend 4** — Phase 3.1, 3.2, 3.3, 3.5 (Asynq, graceful shutdown, slog, healthz)
5. **Weekend 5** — Phase 4.1 + add C++ and Go support
6. **Weekend 6** — Phase 6 (Next.js frontend MVP)
7. **Weekend 7** — Phase 8 (OpenAPI, README, seed, deploy to Fly/Hetzner)
8. **Weekend 8+** — Phase 5 contests, leaderboards, polish

After Weekend 7 you have a hostable, demoable system. Phases 5 and 7 are gravy.

---

## Closing notes

The codebase is in surprisingly good shape architecturally — Clean Architecture is consistently applied, the domain layer is honest, repository entities are properly separated. The gaps are mostly *operational* and *security* rather than *structural*. That's the fortunate kind of debt: it's additive, not corrective.

Two pieces of personal advice:

1. **Resist scope creep on Phase 5.** Contests + leaderboards are sexy but you'll regret shipping them on top of a system that loses jobs in the queue.
2. **Phase 0.1 (admin role) is the highest-leverage hour you'll spend on this project.** Until that exists, you cannot safely deploy anywhere a stranger can register.

Good luck — this can be a genuinely impressive portfolio project once Phase 1 lands.
