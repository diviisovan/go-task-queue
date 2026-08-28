# go-task-queue

A background task processing API built in Go using only the standard library (no web framework, no ORM).

## What does it do?

A REST API that accepts tasks, persists them to MySQL, and processes them asynchronously through a goroutine-powered worker that consumes a buffered channel. When a task is created via `POST /tasks`, it is saved as `pending` and enqueued onto the channel. A background goroutine picks it up, marks it `processing`, executes it (word-count analysis of the description), and marks it `completed` or `failed`. On shutdown (SIGINT/SIGTERM), the worker drains the queue before exiting — no task is silently dropped.

### Endpoints

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

## Where did Claude get it right, and where did you have to fix its output?

**Right:** The `cmd/` + `internal/` project layout, the `context.WithTimeout` pattern for database calls, and the interface-based approach for testing handlers without a real database — Claude produced correct, idiomatic Go for all of these on the first attempt.

**Fixed:** Claude initially generated the project with Gin and GORM (my first Go project, `go-commerce`, accepted that). For this project I pushed back — the deliverable specifically asks for stdlib. Claude also proposed using `log.Println` instead of `log/slog`; I asked it to show me the structured logging alternative and switched. The worker's drain-on-shutdown logic needed manual adjustment: Claude's first version used a bare `close(queue)` which panics if the producer writes after close — I replaced it with a context-cancellation pattern where the producer checks `ctx.Done()` before sending.

**General pattern:** Claude is good at scaffolding correct structure fast. Where it fails is in choosing the *right* library for the constraint — it defaults to the most popular option (Gin, GORM) rather than what the task requires (stdlib). Catching that requires reading the assignment first and knowing what to ask for.

## Which Go concept confused you most, and what do you understand now that you didn't on Monday?

**Pointer receivers vs value receivers.** Coming from Node.js where everything is reference-typed, Go's distinction between `func (s Store)` and `func (s *Store)` was genuinely confusing. A value receiver copies the entire struct — so a method on `Store` (which holds a `*sql.DB`) would get its own copy of the struct, but the `*sql.DB` pointer inside it still points to the same pool, so mutations to the pool work but mutations to the struct's own fields don't. That half-working behavior is worse than a clean failure, because it passes basic tests and breaks later.

What I understand now: if the method changes the receiver, or the receiver is large, or it holds a pointer that makes copying misleading — use a pointer receiver. The Go wiki's CodeReviewComments page says it plainly: "if in doubt, use a pointer receiver." I use pointer receivers on every method in this project not because I proved each one needs it, but because the consistent rule is safer than case-by-case reasoning at my current level.

**Second confusion: `context.Context` propagation.** In Node.js, request cancellation is implicit (the client disconnects and the response stream errors). In Go, `context` must be threaded explicitly through every function call. It felt like ceremony until I saw what happens without it: a cancelled HTTP request still runs its database query to completion, holding a connection. Once I traced the full path — `r.Context()` → `context.WithTimeout` → `db.QueryContext(ctx, ...)` → MySQL driver checks `ctx.Done()` before each network round-trip — the explicit design made sense. It is more typing, but every cancellation point is visible in the code rather than hidden in framework middleware.

## How to run

### Prerequisites

- Go 1.22+
- MySQL 8.0+ running on localhost:3306

### Setup

```bash
# Create the database
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS go_task_queue"

# Clone and run
cd D:\projects\go-test
go mod download
go run ./cmd/server
```

The server starts on `:8080`. Set `DSN` env var to override the default MySQL connection string.

### Test

```bash
go test ./... -v
```

Store tests require a running MySQL instance with a `go_task_queue_test` database:

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
