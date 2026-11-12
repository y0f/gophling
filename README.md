# gophling

<!-- forge:begin init -->
<img src="docs/assets/gophling.png" alt="gophling mascot" width="128" height="128" align="right">

Go REST API starter with a resource generator: PostgreSQL, sqlc, OpenAPI,
OIDC/JWT auth, and one-command checks, structured so developers and coding
agents can add features one module at a time.

`forge add resource` writes a module, migration, queries, OpenAPI operations,
tests, and wiring that build and pass the fast check without edits. Each
resource lives in its own package, so work on separate resources stays mostly
in separate files.

[![CI](https://github.com/y0f/gophling/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/y0f/gophling/actions/workflows/ci.yml)
[![GitHub stars](https://img.shields.io/github/stars/y0f/gophling?style=flat&color=2563eb)](https://github.com/y0f/gophling/stargazers)
[![Go Version](https://img.shields.io/github/go-mod/go-version/y0f/gophling?label=Go%20minimum&color=2563eb)](https://github.com/y0f/gophling/blob/main/go.mod)
[![License](https://img.shields.io/badge/license-MIT-2563eb.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/y0f/gophling?sort=semver&color=2563eb)](https://github.com/y0f/gophling/releases/latest)
[![Docker](https://img.shields.io/badge/Docker-Compose-2563eb?logo=docker&logoColor=white)](https://github.com/y0f/gophling/blob/main/deployments/docker-compose.yml)

## Start a project

Use a patched supported Go release (minimum 1.26.11) and Git. Bootstrap uses
the pinned Go 1.27.4 toolchain, downloading it if needed. Docker runs the local
database.

```sh
go run github.com/y0f/gophling/v2/cmd/forge@latest new myapp
cd myapp
go run ./cmd/dev bootstrap
```

The installer asks for a module path and the components to keep. New projects
keep the core API and the example `widget` resource; add components with
`-with`, remove the example with `-drop example`. For automation, pass
`-yes -module github.com/you/myapp`.

| Component | Adds |
| --- | --- |
| `outbox` | Events committed with state changes, relayed with retries |
| `idempotency` | `Idempotency-Key` replay for create requests |
| `otlp` | OTLP trace export plus a Collector, Tempo, and Grafana profile |
| `admin` | Token-protected pprof and expvar on a separate port |
| `release` | Signed binary releases: GoReleaser, cosign, SBOMs, provenance |

Keeping the first four adds 29 files, about 1,750 lines of hand-written Go (tests
included, `internal/gen` excluded), 4 migrations, and 103 third-party packages
to the API build.

`-ref <tag-or-SHA>` selects the scaffold revision, and `.forge.json` records it.
If initialization is interrupted while publishing files, run
`go run github.com/y0f/gophling/v2/cmd/forge@<release> init -recover` from the
project root with the release that performed it.
<!-- forge:end init -->

## Run and verify

```sh
go run ./cmd/dev doctor
go run ./cmd/dev check
./bin/task up
```

The fast check runs without a database. `./bin/task up` starts the API and
Postgres.

The API is at `http://localhost:8080`, the reference at `/docs`, and the
contract at `/openapi.yaml`. Read the development bearer token from the API log:

```sh
docker compose -f deployments/docker-compose.yml logs api
```

`./bin/task down` stops the stack and keeps data. `./bin/task reset` also
deletes the database volume.

## Add a feature

```sh
go run ./cmd/forge add resource Note
go run ./cmd/dev check
```

The generator adds a module, migration, SQL queries, OpenAPI operations, tests,
and route wiring, then regenerates the bindings. Reads are public; writes
require `notes:write` or an explicit role grant. The logged development token
has the `admin` role, which gets no automatic grant for new resources. Apply
the migration before serving the routes.

Alternatively, ask your coding agent to perform the same change:

> Add a Note resource. Keep reads public and require notes:write for writes.
> Test invalid input and denied writes, then run the fast and integration checks.

[Agent setup](docs/agents.md) covers instruction loading, local setup, and
hosted environments.
[Parallel work](docs/workflows/parallel.md) assigns modules to separate
worktrees and shared changes to one coordinator.
[The resource guide](docs/workflows/resource.md) gives the complete workflow.

## Included

| Area | Implementation |
| --- | --- |
| HTTP | `net/http`, chi, OpenAPI validation, generated interfaces, RFC 9457 errors |
| Data | PostgreSQL, pgx, sqlc, embedded goose migrations |
| Auth | OIDC/JWKS or RSA verification; service-level permissions |
| Operations | Structured logs, Prometheus metrics, health probes, graceful drain |
| Tests | Unit tests and isolated real-Postgres integration tests |
| Supply chain | Pinned tools and Actions, vulnerability checks, non-root image |

<!-- forge:begin outbox -->
Optional outbox: commit events with state changes and retry failed deliveries.
The example publisher only logs event metadata; replace it for external delivery.
<!-- forge:end outbox -->
<!-- forge:begin idempotency -->
Optional idempotency: replay create requests with the same caller, key, and
body; reject a changed body with 409.
<!-- forge:end idempotency -->
<!-- forge:begin otlp -->
Optional tracing: OTLP export with a local Collector, Tempo, and Grafana stack.
<!-- forge:end otlp -->
<!-- forge:begin admin -->
Optional admin listener: token-protected pprof and expvar on a separate port.
<!-- forge:end admin -->

Production mode requires a token verifier. The local Compose stack explicitly
enables development mode and its temporary signing key.

## Documentation

- [Setup and commands](docs/agents.md)
- [Parallel worktrees](docs/workflows/parallel.md)
- [Architecture and decisions](docs/architecture.md)
- [Schema changes](docs/workflows/schema.md)
- [Authorization](docs/workflows/authorization.md)
- [Testing](docs/workflows/testing.md)
<!-- forge:begin init -->
- [Fit and alternatives](docs/choosing.md)
- [Agent evaluation results and method](evals/README.md)
- [Recorded resource workflow](https://y0f.github.io/gophling/demo.html)
<!-- forge:end init -->
- [Contributing](.github/CONTRIBUTING.md)

<!-- forge:begin release -->
## Verify a release

Verify downloaded archives with the signed checksums and provenance steps in
[release operations](.github/RELEASING.md#verify-a-release).
<!-- forge:end release -->

MIT. See [LICENSE](LICENSE).
