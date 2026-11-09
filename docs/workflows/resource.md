# Add a resource

```sh
go run ./cmd/forge add resource Note
```

The generator adds `internal/modules/note`, a migration, SQL registration,
OpenAPI operations, tests, and composition wiring, then runs
`go run ./cmd/dev generate`. List, create, get, update, and delete use the same
contract and pagination conventions as the API. Add `-dry-run` to inspect the
planned files before writing them, or `-no-generate` to add several resources
before generating once. Names must be ASCII; `main`, `internal`, and `testdata`
are rejected.
Generated SQL identifiers must fit PostgreSQL's 63-byte limit, including the
`_list_idx` suffix; shorten names rejected by preflight. SQL names are quoted,
so resource names such as `Reference` can use reserved words as table names.
Preflight also checks literal public table/index declarations in migration Up
sections, including drops and renames. It does not parse arbitrary SQL or infer
all database objects; generation and real-Postgres integration tests are still
required.

Concurrent generation in one checkout is locked. After an interrupted command,
inspect `.forge-resource.lock` and confirm its process has stopped before
removing the lock. Review the diff and restore unfinished generated source
before retrying. Separate worktrees avoid sharing that lock.

1. Edit the module's domain and service for its business rules.
2. Update query SQL, migrations, and OpenAPI schemas for the data model.
3. Choose access rules. Reads are public; writes require `notes:write`.
   Issue that scope or grant it to selected roles in `internal/auth/principal.go`.
4. Run `go run ./cmd/dev generate` after SQL or contract edits.
5. Test successful requests, invalid input, and denied access.
6. Run `go run ./cmd/dev check` and `go test -tags integration -count=1 ./...`.
   Integration tests apply migrations to isolated databases.

To migrate your development database, set `FORGE_DB_URL` (see `.env.example`)
and run `go run ./cmd/migrate up`.

The logged development token has the `admin` role. For local Note writes, add
`notes:write` to that role's existing grants in `internal/auth/principal.go`
(create the entry if absent), then rebuild/restart the API. New resources receive
no automatic role grants. See [authorization](authorization.md) for other policies.

Do not edit `internal/gen`. Module files outside that directory are ordinary
application code. The route must appear in `/openapi.yaml` and work through
the public router before the feature is complete.

For an operation on an existing resource, add its contract path and operation
ID, regenerate, implement the typed handler, and delegate it in
`cmd/api/resources.go`. Keep policy decisions in the service.

<!-- forge:begin example -->
`internal/modules/widget` includes a complete persistence and HTTP example.
<!-- forge:end example -->
