# go-task-queue

A background task processing API built in Go using only the standard library (no web framework, no ORM).

## What it does

A REST API that accepts tasks, persists them to MySQL, and processes them asynchronously through a goroutine-powered worker that consumes a buffered channel. When a task is created via `POST /tasks`, it is saved as `pending` and enqueued onto the channel. A background goroutine picks it up, marks it `processing`, executes it (word-count analysis of the description), and marks it `completed` or `failed`. On shutdown (SIGINT/SIGTERM), the worker drains the queue before exiting — no task is silently dropped.

### Why this exists alongside `go-commerce` — a deliberate comparison, not duplicate work

I built the Week 1 deliverable twice, on purpose, to find out what a framework hides.

| | [`go-commerce`](https://github.com/diviisovan/go-commerce) | `go-task-queue` (this repo) |
|---|---|---|
| HTTP | Gin | `net/http`, Go 1.22 method patterns |
| Database | GORM | `database/sql`, hand-written SQL |
| Routing | `r.GET("/products/:id", …)` | `mux.HandleFunc("GET /tasks/{id}", …)` |
| Request binding | `c.ShouldBindJSON` + `binding:` tags | `json.NewDecoder` + a `Validate()` method I wrote |
| Middleware | `gin.HandlerFunc`, `c.Next()` | `func(http.Handler) http.Handler` |
| Response | `c.JSON(200, obj)` | one `writeJSON` helper |
| Migrations | `AutoMigrate` | explicit `CREATE TABLE IF NOT EXISTS` |

**What I learned from doing both.** Gin's `c.JSON` and GORM's `AutoMigrate` are roughly twenty lines of standard library each — convenient, but they hid decisions I did not know I was making. Three concrete examples:

- **Gin's binding tags** validate for you, so I never thought about what happens to a field a client sends but I did not declare. Writing the stdlib version made me add `DisallowUnknownFields`, which is what stops a client setting a field it should not control.
- **`c.JSON` hides the write order.** Doing it by hand taught me that headers must be set, then `WriteHeader(status)`, then the body — write the body first and you have silently sent a `200`.
- **GORM's soft delete rewrites every query** with `WHERE deleted_at IS NULL`. I only understood that after writing the SQL myself and noticing the clause was not there.

Since Go 1.22 added method patterns and `{wildcards}` to `net/http`, a service this size genuinely does not need a router library. I would still reach for GORM on a large CRUD admin, but I now choose it knowing what it costs rather than by default.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/tasks` | Create a task and enqueue it |
| `GET` | `/tasks` | List all tasks (newest first) |
| `GET` | `/tasks/{id}` | Get a single task by ID |
| `GET` | `/health` | Health check with queue length |
| `GET` | `/docs` | Swagger UI |
| `GET` | `/swagger.json` | OpenAPI 3.0 spec |

### Stack

- **HTTP**: `net/http` with Go 1.22+ method-pattern routing (`POST /tasks`, `GET /tasks/{id}`) — no Gin, no Chi, no framework
- **Database**: `database/sql` + `github.com/go-sql-driver/mysql` — no GORM, no ORM
- **Logging**: `log/slog` with JSON handler — structured, leveled
- **Concurrency**: goroutine + buffered channel for background processing, `context.WithTimeout` for per-request and per-job deadlines, signal-based graceful shutdown
- **Tests**: `go test` with interface-based mocks (handler, worker) and real MySQL integration tests (store)

### Where the eight requirements are met

| # | Requirement | Where |
|---|---|---|
| 1 | REST API, ≥3 endpoints | `internal/handler/handler.go:36-39` — create / list / get / health |
| 2 | Real database | `internal/store/store.go` — MySQL via `database/sql` |
| 3 | Goroutine + channel doing real work | `internal/worker/worker.go:27` buffered `chan model.Task`; consumer goroutine `:46`; `select` on `ctx.Done()` `:51` |
| 4 | `context` for cancellation and timeouts | 8 × `WithTimeout`, 3 × `WithCancel`; per-job deadline `internal/worker/worker.go:73` |
| 5 | Structured logging | `log/slog` JSON handler, `cmd/server/main.go`; zero `log.Println` |
| 6 | Unit tests pass | 15 test functions across handler / store / worker |
| 7 | Dockerfile builds and runs | multi-stage, `golang:1.25-alpine` → `alpine:3.20` |
| 8 | README answers the three questions | the three sections below |

## How to run

### Prerequisites

- Go 1.23+
- MySQL 8.0+ running on localhost:3306

### Setup

```bash
git clone https://github.com/diviisovan/go-task-queue.git
cd go-task-queue

# Create the database
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS go_task_queue"

go mod download
go run ./cmd/server
```

The server starts on `:8080`. Set `DSN` env var to override the default MySQL connection string.

### Test

```bash
go test ./... -v
```

Store tests need a MySQL instance with a `go_task_queue_test` database. Without one they **skip** rather than fail, so `go test ./...` is green on a machine with no database — which is what lets CI run them:

```bash
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS go_task_queue_test"
go test ./internal/store/ -v
```

### Docker

```bash
docker build -t go-task-queue .
docker run -p 8080:8080 -e DSN="root:password@tcp(host.docker.internal:3306)/go_task_queue?parseTime=true" go-task-queue
```

### Try it

```bash
# Create a task
curl -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"Test task","description":"This is a test"}'

# List tasks (should show completed after ~100ms)
curl http://localhost:8080/tasks

# Get a single task
curl http://localhost:8080/tasks/1

# Health check
curl http://localhost:8080/health
```

Swagger UI at http://localhost:8080/docs

## Where Claude was right, and where I had to fix it

**Right:** The `cmd/` + `internal/` project layout, the `context.WithTimeout` pattern for database calls, and the interface-based approach for testing handlers without a real database — Claude produced correct, idiomatic Go for all of these on the first attempt.

**Fixed:** Claude initially generated the project with Gin and GORM (my first Go project, `go-commerce`, accepted that). For this project I pushed back — the deliverable specifically asks for stdlib. Claude also proposed using `log.Println` instead of `log/slog`; I asked it to show me the structured logging alternative and switched. The worker's drain-on-shutdown logic needed manual adjustment: Claude's first version used a bare `close(queue)` which panics if the producer writes after close — I replaced it with a context-cancellation pattern where the producer checks `ctx.Done()` before sending.

**General pattern:** Claude is good at scaffolding correct structure fast. Where it fails is in choosing the *right* library for the constraint — it defaults to the most popular option (Gin, GORM) rather than what the task requires (stdlib). Catching that requires reading the assignment first and knowing what to ask for.

## The Go concept that confused me most

**Pointer receivers vs value receivers.** Coming from Node.js where everything is reference-typed, Go's distinction between `func (s Store)` and `func (s *Store)` was genuinely confusing. A value receiver copies the entire struct — so a method on `Store` (which holds a `*sql.DB`) would get its own copy of the struct, but the `*sql.DB` pointer inside it still points to the same pool, so mutations to the pool work but mutations to the struct's own fields don't. That half-working behavior is worse than a clean failure, because it passes basic tests and breaks later.

**Second confusion: `context.Context` propagation.** In Node.js, request cancellation is implicit (the client disconnects and the response stream errors). In Go, `context` must be threaded explicitly through every function call. It felt like ceremony rather than engineering.

## What I understand now that I did not before

**On receivers:** if the method changes the receiver, or the receiver is large, or it holds a pointer that makes copying misleading — use a pointer receiver. The Go wiki's CodeReviewComments page says it plainly: "if in doubt, use a pointer receiver." I use pointer receivers on every method in this project not because I proved each one needs it, but because the consistent rule is safer than case-by-case reasoning at my current level.

**On `context`:** I saw what happens without it — a cancelled HTTP request still runs its database query to completion, holding a connection. Once I traced the full path — `r.Context()` → `context.WithTimeout` → `db.QueryContext(ctx, ...)` → MySQL driver checks `ctx.Done()` before each network round-trip — the explicit design made sense. It is more typing, but every cancellation point is visible in the code rather than hidden in framework middleware.

**On frameworks, from building the same brief twice:** I can now say specifically what Gin and GORM were doing for me, because I had to write each piece myself — unknown-field rejection, response write ordering, and the soft-delete clause GORM silently adds to every query. That is the difference between using a framework and depending on one.
