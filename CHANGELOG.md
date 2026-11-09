# Changelog

## v2.1.0

### Changed

- New projects keep only the core API and the example resource. The outbox,
  idempotency keys, OTLP tracing, and the admin listener are opt-in with
  `-with outbox,idempotency,otlp,admin`; `-drop` still removes components.
  Compared with the v2.0.0 default, a fresh project has 127 instead of 151
  files, about 7,400 instead of 8,800 lines of hand-written Go (tests included,
  `internal/gen` excluded), and 171 instead of 266 third-party packages in the
  API build.
- `forge add resource` runs `go run ./cmd/dev generate` itself, so the tree
  builds immediately. Pass `-no-generate` to batch several resources and
  generate once. The printed next steps no longer include
  `go run ./cmd/migrate up`, which failed without `FORGE_DB_URL`; integration
  tests apply migrations to their own databases.
- `dev check` stops at generation or compile errors; otherwise it runs lint and
  unit tests to completion and names every failing step. It also formats the
  Go files changed in the working tree (formatters only, never lint fixes);
  `dev ci` stays check-only.
- The example widget validates names like generated resources: blank or
  whitespace-only names now return 400.
- Generated idempotency code is a thin per-resource adapter over shared
  `internal/idempotency` and `internal/platform/httpjson` helpers. A generated
  resource is 606 lines in the default project and 786 with every component,
  down from 929. Stored keys, hashes, statuses, and messages are unchanged.
- Agent-facing docs use one spelling per command (`go run ./cmd/dev <verb>`,
  `go test ...`). Release verification moved to
  `.github/RELEASING.md#verify-a-release`.
- Pinned tools: golangci-lint v2.15.0, oapi-codegen v2.8.2, govulncheck v1.9.0,
  Task v3.54.0, x/tools v0.52.0. Dependencies: kin-openapi v0.153.0,
  testcontainers-go v0.45.0. Container images move to Debian 13
  (`golang:1.27.3-trixie`, distroless `static-debian13:nonroot`).

### Added

- `go run ./cmd/dev fmt [packages]` runs the formatters lint enforces.

### Fixed

- `forge add resource` rejects `Main`, `Internal`, and `Testdata`, which
  produced a project that failed to build or whose tests were silently skipped,
  and rejects non-ASCII names before writing. `/.forge-resource.lock` is
  ignored by Git.

### Removed

- The unused `auth.Middleware` and `RouterDeps.Mounts`. Contract routes
  authenticate through the OpenAPI validator; routes outside the contract go on
  `BaseRouter`, unvalidated and unauthenticated.

## v2.0.0

- The scaffold's Go module is now `github.com/y0f/gophling/v2`. Create a project
  with `go run github.com/y0f/gophling/v2/cmd/forge@v2.0.0 new myapp`.
  Existing initialized applications keep their own module paths and do not
  upgrade automatically.
- The repository remains `github.com/y0f/gophling`. Clone URLs and the scaffold
  identity in `.forge.json` remain stable; the launcher still selects its own
  release and records the resolved source revision.
- A project module such as `example.com/team/api/v2` keeps that full import
  path, while its README title and default release name use `api` and its
  independent tools module uses `example.com/team/api/tools`. Compose names
  include a hash of the full module path to distinguish application modules.
- The selected scaffold source must use the launcher's exact module path.
  A v2 launcher rejects v1 source before transformation; use the launcher and
  original module path from the selected older release. For interrupted
  initialization, recover with the release that performed it.
- Remove trailing blank lines left in Markdown by initialization's optional
  sections, so fresh full and minimal projects pass Git's whitespace check.
  Markdown hard breaks and existing file permissions are preserved.

## v1.3.0

### Upgrade notes

- The project is now named `gophling`. New installations use
  `go run github.com/y0f/gophling/cmd/forge@v1.3.0 new myapp`.
  Existing initialized projects do not upgrade automatically; review the
  relevant changes before applying them to your application.
  Replace launcher commands using `github.com/y0f/go-api-scaffolding` with the
  new module path. Releases through v1.2.0 retain the former module path.
- `forge new` selects the launcher's release by default and records the exact
  scaffold revision in `.forge.json`. `-ref` selects another tag or commit.
  Initialization verifies a staged copy before publishing and preserves file
  permissions through staging and recovery. If publication is interrupted, use
  the same launcher's `init -recover` command before editing.
- Generated resources now extend OpenAPI, SQL, migrations, tests, and explicit
  application wiring together. Run `go run ./cmd/dev generate` and apply the
  migration before serving a new resource.
- Initialization retains original license notices and adds the new module
  owner's notice.
- The minimum Go version is 1.26.10; development tools use Go 1.27.3. Bootstrap
  downloads the preferred toolchain when needed. `go run ./cmd/dev check` runs
  the fast checks without Docker; `ci` adds integration, race, vulnerability,
  generated-project, and container checks.
- Idempotent write replays recheck current permissions. Shutdown deadlines now
  cancel unfinished HTTP requests. Sensitive parent log groups are redacted.
- Request validation rejects explicit `null` for non-nullable fields, including
  fields with defaults. Handlers still apply defaults when a value is omitted.
  Validation no longer rewrites request bodies to inject defaults. When applying
  this change to an existing service, retries against idempotency entries hashed
  from the older normalized body can return 409 until those entries expire;
  do not blindly retry an existing operation with a new key.
- `task down` retains the development database; use `task reset` to delete it.
  Development ports bind to loopback by default.
- Fresh projects derive their Compose project name from the full module path,
  separating applications with the same basename. Existing initialized projects
  and volumes are unchanged. A different project name selects a different volume;
  migrate any data deliberately and keep the original volume until verified.
  Worktrees sharing a module need explicit project-name and port overrides for
  simultaneous development stacks.
- Shared agent instructions, worktree guidance, evaluation runners, and a
  recorded resource workflow document the development process. Compatibility
  guidance and measured client results are reported separately.
- Pull requests now require a yes/no AI-assistance declaration, tool and scope
  details when applicable, and actual validation evidence. Copilot setup is
  checked by CI and repository review instructions cover critical API behavior.
- Public GitHub.com releases add build provenance attestations alongside signed
  checksums and SBOMs. Uploaded artifacts are verified before draft publication;
  reruns verify published immutable releases without replacing their assets.

## v1.2.0

### Upgrade notes

- `GET /v1/widgets` now rejects `offset` above 10000 with a 400, because the
  database reads every skipped row. Clients that page to the end with
  `offset += limit` get a 400 once they pass 10000, and widgets past
  offset 10000 plus `limit` can no longer be listed, although `total` still
  counts them. If you need deeper pages, raise `maximum` on the `Offset`
  parameter in `api/openapi.yaml` and run `task generate`.
- `FORGE_ENV` now defaults to `production` (v1.1.0 defaulted to
  `development`). Set it explicitly. A deployment that left it unset now
  refuses to start without `FORGE_AUTH_JWKS_URL` or
  `FORGE_AUTH_PUBLIC_KEY_PATH`, and one that has a verifier switches to
  production behaviour: CORS no longer reflects any origin, so list browser
  origins in `FORGE_HTTP_CORS_ALLOWED_ORIGINS`.
- `FORGE_LOG_LEVEL` must be `debug`, `info`, `warn`/`warning` or `error`, and
  `FORGE_LOG_FORMAT` must be `json` or `text` (case-insensitive, surrounding
  whitespace ignored). Other values, such as `trace`, used to fall back to
  `info` or `json` and now stop startup.
- A create with an `Idempotency-Key` from a token without a `sub` claim now
  returns 400 instead of sharing one key namespace with every other
  subject-less caller. Send such requests without the header, or issue tokens
  with a subject.
- The outbox relay no longer stops a batch at the first message the publisher
  rejects: that message records its error, backs off (2s doubling up to one
  hour) and is retried later while the messages behind it are published. So
  delivery order is no longer guaranteed; consumers that relied on it must
  order by their own data.
- Migration `00005_widgets_list_index.sql` replaces the widgets list index
  with `CREATE INDEX CONCURRENTLY` outside a transaction, so reads and writes
  on `widgets` continue while it builds. If it fails part-way, run
  `migrate up` again: it drops the half-built index and starts over.
- Idempotency keys are now stored under an escaped subject, so subjects that
  contain `:` can no longer collide with other subjects' keys. For a subject
  that contains `:` or `%`, a key stored by v1.1.0 is not found after the
  upgrade, so a retry of such a request within the key's TTL creates the
  resource again. Keys of all other subjects are unchanged.
- `migrate` now fails with `FORGE_DB_URL is required` when the variable is
  unset or empty. v1.1.0 fell back to the development default
  `postgres://forge:forge@localhost:5432/forge`; set `FORGE_DB_URL` (or copy
  `.env.example` to `.env` for `task migrate`) wherever migrations run.
- Migration `00006_outbox_published_at_idx.sql` builds its index with
  `CREATE INDEX CONCURRENTLY` outside a transaction, so outbox writes are not
  blocked while it runs. If it fails part-way, run `migrate up` again: it drops
  the invalid index and rebuilds it.
