# Skoolz Go API — Getting Started (for Node.js developers)

This is a step-by-step walkthrough of the project for someone coming from a
Node.js / Express background. It covers what each package does, the build &
run flow, and the new **Tasks CRUD** we just added.

---

## 1. Mental model: Node.js → Go

| Node.js                           | Go (this project)                                   |
| --------------------------------- | --------------------------------------------------- |
| `package.json`                    | `go.mod`                                            |
| `node_modules/`                   | `$GOPATH/pkg/mod` (managed by Go, not in repo)      |
| `npm install`                     | `go mod download` / `go mod tidy`                   |
| `npm run dev`                     | `make dev` (uses `air` for hot reload)              |
| `npm run build`                   | `make build` → produces a single binary `./main`    |
| `node index.js`                   | `./main serve-rest`                                 |
| Express `app.get('/x', handler)`  | `mux.Handle("GET /x", handler)` (Go 1.22+ ServeMux) |
| `req.body` (parsed by middleware) | `json.NewDecoder(r.Body).Decode(&req)`              |
| `res.json({...})`                 | `response.WriteOK(w, "msg", data)`                  |
| `process.env.FOO`                 | `os.Getenv("FOO")` or struct tags via `envconfig`   |
| `pg` / Prisma                     | `sqlx` + `lib/pq` (Postgres driver)                 |
| Knex migrations                   | Plain `.sql` files run by `cmd/migrate.go`          |

The biggest change: **Go compiles to a single static binary**. There is no
`node_modules` shipped with the app — `go build` produces one executable that
you run directly.

---

## 2. Project layout

```
go-api/
├── main.go                  # Entry point - just calls cmd.Execute()
├── go.mod / go.sum          # Module + locked dependencies
├── Makefile                 # All common commands (build/run/migrate/test)
├── .env                     # Local config (copied from example.env)
├── example.env              # Template — commit-safe defaults
├── Dockerfile
│
├── cmd/                     # CLI entry points (cobra commands)
│   ├── root.go              #   `master-service` root command
│   ├── rest-api.go          #   `serve-rest`  → boots HTTP server
│   ├── grpc.go              #   `serve-grpc`  → boots gRPC server (build tag)
│   ├── grpc_stub.go         #   no-op when built without -tags grpc
│   └── migrate.go           #   `migrate`     → runs SQL files in order
│
├── config/                  # Configuration loading
│   ├── config.go            #   Struct + env tags (envconfig style)
│   ├── load.go              #   Loads .env → populates struct
│   └── db.go                #   NewPostgresDB() — opens sqlx.DB
│
├── database/postgres/
│   └── repositories/        # SQL data access layer (think "models")
│       ├── user_repository.go
│       └── task_repository.go   ← NEW
│
├── internal/
│   ├── infrastructure/
│   │   ├── container/       # Singleton DI container
│   │   │   └── container.go #   Holds db, redis, logger, messaging
│   │   ├── database/postgres/migrations/
│   │   │   └── 001_create_tasks_table.sql   ← NEW
│   │   ├── external/
│   │   └── messaging/       # Kafka + NATS wrappers
│   │
│   ├── interfaces/
│   │   ├── http/            # REST layer
│   │   │   ├── server.go    #   net/http server + graceful shutdown
│   │   │   ├── routes/      #   route table (ServeMux registration)
│   │   │   ├── handlers/    #   per-resource HTTP handlers
│   │   │   │   ├── welcome_handler.go
│   │   │   │   ├── health_handler.go
│   │   │   │   ├── not_found_handler.go
│   │   │   │   └── task_handler.go         ← NEW
│   │   │   └── middleware/  #   logging, recover, auth, CORS
│   │   ├── grpc/            # gRPC layer (excluded unless built with -tags grpc)
│   │   └── cli/
│   │
│   ├── logger/              # slog + lumberjack JSON logger
│   └── shared/
│       ├── response/        # Helpers: WriteOK / WriteCreated / WriteBadRequest
│       ├── error/           # Typed API errors
│       ├── exceptions/
│       ├── types/
│       └── utils/
│
├── pkg/cache/               # Redis client wrapper
└── proto/                   # .proto definitions (gRPC schemas)
    ├── health/health.proto
    └── welcome/welcome.proto
```

### Why `internal/`?

Anything inside `internal/` can only be imported by code under the same
parent module (`skoolz/`). It's Go's enforced "private package" mechanism —
the compiler refuses external imports. Use it for everything that isn't a
deliberately reusable library.

### Why `cmd/`?

Convention. Each subcommand of your binary lives here. `cobra` is the CLI
library — analogous to `commander` in Node.

---

## 3. Package-by-package: what does what

| Path                                                    | Responsibility                                                   | Node.js analogue                                  |
| ------------------------------------------------------- | ---------------------------------------------------------------- | ------------------------------------------------- |
| `main.go`                                               | Entry point, just calls `cmd.Execute()`                          | `index.js`                                        |
| `cmd/`                                                  | Cobra CLI commands (`serve-rest`, `migrate`, ...)                | `bin/cli.js` + `commander`                        |
| `config/`                                               | Loads `.env` into a typed `Config` struct via `envconfig`        | `dotenv` + manual mapping                         |
| `internal/infrastructure/container/`                    | Singleton DI: holds DB, Redis, logger, messaging                 | A `services` module exporting initialised clients |
| `internal/infrastructure/messaging/`                    | Kafka + NATS publishers/subscribers                              | `kafkajs` / `nats.js` wrappers                    |
| `internal/infrastructure/database/postgres/migrations/` | Plain `.sql` files run in alphabetical order by `cmd/migrate.go` | Knex/Prisma migration files                       |
| `database/postgres/repositories/`                       | SQL queries grouped by entity (`User`, `Task`)                   | Prisma models / repository pattern                |
| `internal/interfaces/http/server.go`                    | Constructs `http.Server`, ListenAndServe, graceful shutdown      | `app.listen(...)` + signal handling               |
| `internal/interfaces/http/routes/routes.go`             | Registers routes on `http.ServeMux`                              | `app.use('/api', router)`                         |
| `internal/interfaces/http/handlers/`                    | One file per resource — parses input, calls repo, writes JSON    | Express controllers                               |
| `internal/interfaces/http/middleware/`                  | Cross-cutting concerns (logging, auth, CORS)                     | Express middleware                                |
| `internal/shared/response/`                             | `WriteOK`, `WriteCreated`, `WriteNotFound`, ...                  | A `respond.js` helper                             |
| `internal/shared/error/`                                | API error type with status code + error code                     | Custom `HttpError` class                          |
| `internal/logger/`                                      | Structured slog + log rotation via `lumberjack`                  | `pino` or `winston`                               |
| `pkg/cache/`                                            | Redis client + helpers                                           | `ioredis` wrapper                                 |
| `proto/`                                                | gRPC `.proto` files; codegen produces `*.pb.go`                  | gRPC `.proto` files                               |

### Third-party libraries you'll see in `go.mod`

| Import                               | What it does                                            |
| ------------------------------------ | ------------------------------------------------------- |
| `github.com/spf13/cobra`             | CLI framework — defines the subcommands                 |
| `github.com/spf13/viper`             | Config loader (used here partially)                     |
| `github.com/joho/godotenv`           | Loads `.env` files                                      |
| `github.com/jmoiron/sqlx`            | A small extension over `database/sql` (struct scanning) |
| `github.com/lib/pq`                  | Postgres driver                                         |
| `github.com/google/uuid`             | UUID generation/parsing                                 |
| `github.com/redis/go-redis/v9`       | Redis client                                            |
| `github.com/segmentio/kafka-go`      | Kafka client                                            |
| `github.com/nats-io/nats.go`         | NATS client                                             |
| `github.com/golang-jwt/jwt/v5`       | JWT signing/verification                                |
| `github.com/go-playground/validator` | Struct-tag validation (like `class-validator`)          |
| `google.golang.org/grpc`             | gRPC runtime                                            |
| `gopkg.in/natefinch/lumberjack.v2`   | Rotating log file writer                                |
| `go.elastic.co/apm`                  | Elastic APM tracing                                     |

---

## 4. Prerequisites

You need:

1. **Go 1.24+** (`go version`)
2. **PostgreSQL 13+** running and reachable. Easiest: Docker.
3. (Optional) Docker for Kafka/Redis/NATS if you want messaging.

Verify Postgres is up — in this environment we use the `postgres-db` container:

```bash
docker ps | grep postgres
# postgres-db   postgres:15   0.0.0.0:5432->5432/tcp
docker exec postgres-db psql -U postgres -l
```

---

## 5. First-time setup

```bash
# 1. Copy env template
cp example.env .env

# 2. Verify the DB connection string in .env points to your Postgres
#    POSTGRES_HOST=localhost
#    POSTGRES_PORT=5432
#    POSTGRES_USER=postgres
#    POSTGRES_PASSWORD=password
#    POSTGRES_DATABASE=skoolz_db

# 3. Create the database if it doesn't exist
docker exec postgres-db psql -U postgres -c "CREATE DATABASE skoolz_db;"

# 4. Download Go dependencies (one-time)
go mod download
# or:
make install-deps
```

---

## 6. Build → Migrate → Run (the happy path)

```bash
# Build the binary (single file, ./main)
make build

# Apply pending SQL migrations to Postgres
make migrate

# Start the REST API on :9090
make run
```

Or all in one shot:

```bash
make start
```

You should see logs ending with:

```
🚀 Server starting port :9090
```

Stop the server:

```bash
make stop          # kills any ./main process
# or just Ctrl+C in the terminal where `make run` is running
```

---

## 7. Verifying it works

```bash
# Health
curl -s http://localhost:9090/health

# Welcome
curl -s http://localhost:9090/api/v1
```

---

## 8. Tasks CRUD — what we added and how to use it

### 8.1 Migration

`internal/infrastructure/database/postgres/migrations/001_create_tasks_table.sql`:

```sql
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC);
```

`cmd/migrate.go` creates a `migrations` book-keeping table, then runs every
`*.sql` file in alphabetical order, recording which ones it has executed
(idempotent).

### 8.2 Repository

`database/postgres/repositories/task_repository.go` is the data layer.
It exposes:

```go
NewTaskRepository(db *sqlx.DB) *TaskRepository
Create(ctx, *Task) error
GetByID(ctx, uuid.UUID) (*Task, error)
List(ctx, limit, offset int) ([]*Task, error)
Update(ctx, *Task) error
Delete(ctx, uuid.UUID) error
Count(ctx) (int64, error)
```

`db` struct tags map columns → fields; `json` tags shape the API output.

### 8.3 Handler

`internal/interfaces/http/handlers/task_handler.go` is the HTTP layer.
Each method:

1. Decodes the request body / path params
2. Calls the repository
3. Writes a JSON response via `internal/shared/response`

### 8.4 Routes

Registered in `internal/interfaces/http/routes/routes.go`:

| Method | Path                 | Handler              |
| ------ | -------------------- | -------------------- |
| POST   | `/api/v1/tasks`      | `taskHandler.Create` |
| GET    | `/api/v1/tasks`      | `taskHandler.List`   |
| GET    | `/api/v1/tasks/{id}` | `taskHandler.Get`    |
| PUT    | `/api/v1/tasks/{id}` | `taskHandler.Update` |
| DELETE | `/api/v1/tasks/{id}` | `taskHandler.Delete` |

Note: Go 1.22+ `http.ServeMux` natively supports method-prefixed patterns
(`"GET /api/v1/tasks/{id}"`) and `r.PathValue("id")` — no router library
needed.

### 8.5 Try it

```bash
# Create
curl -s -X POST http://localhost:9090/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","description":"finish CRUD","status":"pending"}'

# List (paginated)
curl -s "http://localhost:9090/api/v1/tasks?limit=20&offset=0"

# Get by id
curl -s http://localhost:9090/api/v1/tasks/<TASK_ID>

# Update
curl -s -X PUT http://localhost:9090/api/v1/tasks/<TASK_ID> \
  -H "Content-Type: application/json" \
  -d '{"title":"Learned Go","description":"done","status":"done"}'

# Delete
curl -s -X DELETE http://localhost:9090/api/v1/tasks/<TASK_ID>
```

---

## 8.6 Adding a new table (full migration workflow)

The migration runner (`cmd/migrate.go`) is dead simple:

1. Reads every `*.sql` file in `internal/infrastructure/database/postgres/migrations/` in alphabetical order.
2. Skips files already recorded in the bookkeeping table `migrations` (so it's safe to re-run).
3. Runs each new file inside a transaction and records the filename on success.

So adding a new table = drop a new `.sql` file with a higher number prefix, then run `make migrate`.

### Example: add a `categories` table

#### Step 1 — Create the migration file

Use the next sequential prefix (`002_`, `003_`, ...). Names are sorted as
strings, so always zero-pad.

```bash
# from the repo root
touch internal/infrastructure/database/postgres/migrations/002_create_categories_table.sql
```

Open the file and put your DDL in it:

```sql
-- 002_create_categories_table.sql
CREATE TABLE IF NOT EXISTS categories (
    id          UUID PRIMARY KEY,
    name        VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);
```

#### Step 2 — Run the migration

```bash
make migrate
```

You should see a line like:

```
Migration executed successfully  filename=002_create_categories_table.sql
```

If you re-run it, you'll get:

```
Migration already executed  filename=002_create_categories_table.sql
```

#### Step 3 — Verify the table exists in Postgres

```bash
docker exec postgres-db psql -U postgres -d skoolz_db -c "\dt"
docker exec postgres-db psql -U postgres -d skoolz_db -c "\d categories"
docker exec postgres-db psql -U postgres -d skoolz_db -c "SELECT * FROM migrations ORDER BY id;"
```

#### Step 4 — Add the repository (data layer)

Create `database/postgres/repositories/category_repository.go`:

```go
package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Category struct {
	ID          uuid.UUID `db:"id"          json:"id"`
	Name        string    `db:"name"        json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

type CategoryRepository struct{ db *sqlx.DB }

func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, c *Category) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now()
	c.CreatedAt, c.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO categories (id, name, description, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.Name, c.Description, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	var c Category
	err := r.db.GetContext(ctx, &c, `SELECT * FROM categories WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *CategoryRepository) List(ctx context.Context, limit, offset int) ([]*Category, error) {
	var out []*Category
	err := r.db.SelectContext(ctx, &out,
		`SELECT * FROM categories ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	return out, err
}
```

#### Step 5 — Add the HTTP handler

Create `internal/interfaces/http/handlers/category_handler.go` following the
exact same pattern as `task_handler.go`:

```go
package handlers

import (
	"encoding/json"
	"net/http"

	"skoolz/database/postgres/repositories"
	"skoolz/internal/infrastructure/container"
	"skoolz/internal/shared/response"
)

type CategoryHandler struct {
	repo *repositories.CategoryRepository
}

func NewCategoryHandler() *CategoryHandler {
	db := container.GetContainer().GetDB()
	return &CategoryHandler{repo: repositories.NewCategoryRepository(db)}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c repositories.Category
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		response.WriteBadRequest(w, "Invalid JSON body")
		return
	}
	if c.Name == "" {
		response.WriteBadRequest(w, "name is required")
		return
	}
	if err := h.repo.Create(r.Context(), &c); err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}
	response.WriteCreated(w, "Category created", &c)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cs, err := h.repo.List(r.Context(), 50, 0)
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}
	response.WriteOK(w, "Categories fetched", cs)
}
```

#### Step 6 — Wire the routes

Edit `internal/interfaces/http/routes/routes.go` and add:

```go
categoryHandler := handlers.NewCategoryHandler()

mux.Handle("POST /api/v1/categories", manager.With(http.HandlerFunc(categoryHandler.Create)))
mux.Handle("GET  /api/v1/categories", manager.With(http.HandlerFunc(categoryHandler.List)))
```

#### Step 7 — Rebuild and restart

```bash
make stop
make build
make run        # or `make start` to also re-run migrations
```

#### Step 8 — Test it

```bash
curl -s -X POST http://localhost:9090/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"Books","description":"Anything you can read"}'

curl -s http://localhost:9090/api/v1/categories
```

### Tips & rules of thumb

- **Always zero-pad the prefix** (`001_`, `002_`, ...). String sort = run order.
- **Never edit a migration after it's been applied in any environment** —
  write a new migration to alter the table instead. The runner identifies
  files by filename, so editing in place won't re-run them.
- **One logical change per file.** Easier to roll forward and review.
- **Use `IF NOT EXISTS` / `IF EXISTS`** in DDL so you can re-run on a fresh
  DB without errors.
- **Wrap data backfills in `BEGIN; ... COMMIT;`** — though `cmd/migrate.go`
  already runs each file inside a transaction.
- **Adding a column to an existing table** — example migration body:
  ```sql
  ALTER TABLE tasks ADD COLUMN IF NOT EXISTS priority INT NOT NULL DEFAULT 0;
  CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
  ```
- **Dropping is destructive.** Prefer renaming or marking deprecated until
  all environments are migrated.

### Quick command recap

```bash
# 1. Create the SQL file (next sequential number)
touch internal/infrastructure/database/postgres/migrations/00X_<name>.sql

# 2. Edit it with your DDL

# 3. Apply it
make migrate

# 4. Inspect the DB
docker exec postgres-db psql -U postgres -d skoolz_db -c "\dt"
docker exec postgres-db psql -U postgres -d skoolz_db -c "SELECT * FROM migrations;"

# 5. Build & restart the API
make stop && make build && make run
```

---

## 9. The full Makefile cheat sheet

```bash
make help              # List all targets
make build             # Compile ./main (REST only — no proto needed)
make migrate           # Apply pending SQL migrations
make run               # Start the REST API on :9090
make start             # build + migrate + run
make dev               # Hot reload via `air` (install with `make install-dev-deps`)
make stop              # Kill any running ./main
make clean             # Remove ./main

make tidy              # go mod tidy
make install-deps      # go mod download
make fmt               # gofmt -s -w .
make vet               # go vet ./...
make test              # go test ./...

# Optional, only if you need gRPC:
make install-proto-deps
make proto             # regenerate *.pb.go from proto/*.proto (needs protoc)
make build-grpc        # go build -tags grpc

# Optional, if there's a sibling infra/ docker-compose for Kafka/Redis/NATS:
make infra-up
make infra-down
make infra-logs
```

---

## 10. About the gRPC build tag

The repo has `proto/*.proto` files but the generated `*.pb.go` files were
missing, so `go build ./...` failed. To keep the REST flow unblocked, every
file under `internal/interfaces/grpc/` and `cmd/grpc.go` now starts with:

```go
//go:build grpc
```

That means: **these files are only compiled when you pass `-tags grpc`.**
So:

- `make build` → REST only, always works
- `make build-grpc` → includes gRPC, only works after `make proto`

When you're ready to enable gRPC:

```bash
# 1. Install protoc itself
sudo apt install -y protobuf-compiler          # Debian/Ubuntu
# or: brew install protobuf                    # macOS

# 2. Install the Go plugins
make install-proto-deps

# 3. Generate the .pb.go files
make proto

# 4. Build with gRPC enabled
make build-grpc
./main serve-grpc
```

---

## 11. Hot reload during development

```bash
make install-dev-deps       # one-time: installs `air`
make dev                    # rebuild + restart on every save
```

`air` is the Go equivalent of `nodemon`.

---

## 12. Handy Go commands you'll use a lot

```bash
go version                  # Confirm toolchain
go mod tidy                 # Add missing / drop unused deps (like npm prune)
go mod download             # Fetch deps into module cache
go build ./...              # Compile every package (sanity check)
go run ./                   # Compile + run main (no binary persisted)
go test ./...               # Run all tests
go vet ./...                # Lint-style static checks
gofmt -s -w .               # Format every .go file in place
go doc net/http             # Read package documentation in the terminal
```

---

## 13. Common gotchas coming from Node

1. **Imports must be used.** Unused import = compile error. Same for unused
   local variables. Use `_ "github.com/lib/pq"` for "side-effect only"
   imports (the Postgres driver registers itself this way).
2. **No `try/catch`.** Errors are values: every fallible call returns
   `(value, error)`. Idiomatic pattern:
   ```go
   user, err := repo.GetByID(ctx, id)
   if err != nil { return err }
   if user == nil { return ErrNotFound }
   ```
3. **`context.Context` is everywhere.** It carries deadlines, cancellation,
   and request-scoped values. Always pass `r.Context()` from HTTP handlers
   into downstream calls (DB, RPC).
4. **Pointers vs values.** `*User` is a pointer (mutable, nullable);
   `User` is a value (copied on every call). Repositories typically return
   `*User` so `nil` can mean "not found".
5. **Struct tags drive everything.** JSON shape, DB columns, env vars, and
   validation all come from string tags after each field, e.g.
   `Email string \`db:"email" json:"email" validate:"required,email"\``.
6. **One binary, many subcommands.** `./main serve-rest`, `./main migrate`,
   `./main serve-grpc` — all the same compiled file with different cobra
   subcommands.

---

## 14. What's running where (this environment)

| Service  | Address           | Status                                               |
| -------- | ----------------- | ---------------------------------------------------- |
| Postgres | `localhost:5432`  | Running (Docker container `postgres-db`)             |
| REST API | `localhost:9090`  | `make run`                                           |
| Redis    | `localhost:6379`  | Not running — app continues without cache            |
| Kafka    | `localhost:9092`  | Not running — logs reconnect errors but server works |
| NATS     | `localhost:4222`  | Not running                                          |
| gRPC     | `localhost:50001` | Disabled until protos are generated                  |

If Kafka/Redis noise bothers you in dev, set `KAFKA_ENABLE=false` in `.env`
and ignore the Redis warning — neither blocks the REST API.

---

Happy hacking. When in doubt: `make help`.
