# AGENTS.md — License Manager

Software license management system: Go/Gin backend + Vue 3/TypeScript frontend.
Chinese primary language (zh-CN), with English and Japanese i18n.

## Git workflow

Every round of work goes on its own branch, then merges back into `main`. Do not commit straight to `main` — a bug found mid-task would then sit in the history of the branch that shipped it.

### Branch naming

| Prefix | Use for |
| --- | --- |
| `feature/<slug>` | new capability |
| `fix/<slug>` | defect repair |
| `test/<slug>` | test-only work |

Existing branches follow `feature/s<n>-<slug>` for staged delivery (S0–S10) and plain slugs for ad-hoc work (`fix/ui-alignment`, `test/unit-coverage`). Match whichever fits the task.

### Cycle

```bash
git checkout -b fix/<slug>      # from up-to-date main
# ... work, verifying as you go ...
git add <files>                 # only what this task touched
git commit                      # logical commits, imperative mood
git checkout main
git merge --no-ff fix/<slug>
```

`--no-ff` keeps the branch as a node in the history. Commits are in **Chinese**, matching the existing history; the merge message explains **why** the change was made rather than restating the diff.

### Never push

**Commit locally; leave `origin` alone.** The maintainer pushes and handles remote authentication. Do not run `git push`, do not open PRs.

Before committing, `git status` to confirm the staged set: do not sweep in `.env*.local`, `production/`, or other untracked files that happen to be lying around.

## Commands

### Quick start (both services)

```bash
./scripts/dev.sh                    # starts backend (:18888) + frontend (:3000)
./scripts/dev.sh config.local.yaml  # uses custom backend config
```

### Backend (Go 1.23)

```bash
cd backend
go run ./cmd -config configs/config.local.yaml   # run from backend/ dir, NOT backend/cmd/
go test ./...                                      # run all tests
go test ./internal/service/...                     # service-layer tests only
go build -o ../bin/license-manager cmd/main.go     # binary build
```

### Frontend (pnpm, NOT npm/yarn)

```bash
cd frontend
pnpm install
pnpm run serve          # dev server on :3000
pnpm run build          # typecheck + production build
pnpm run type-check     # vue-tsc --noEmit (standalone typecheck)
pnpm run test           # vitest
```

ESLint is currently disabled (`lint` script is a no-op echo). Prettier config exists at `frontend/.prettierrc`.

## Critical gotchas

### Config path resolution depends on working directory

All relative paths in YAML configs (`i18n.config_path`, `license.rsa.*_key_path`) resolve from the **current working directory**, not from the config file's location.

- Running from `backend/` (normal local dev): use `configs/...` paths
- Running from `backend/cmd/` (Docker): use `../configs/...` paths

`backend/configs/config.dev.yaml` has `../configs/...` paths (Docker convention).
`backend/configs/config.local.yaml.example` has `configs/...` paths (local convention).
Mixing them up causes `log.Fatalf` on startup (missing i18n or RSA files).

### Config lookup order (`backend/cmd/main.go`)

1. `-config` / `-c` flag
2. `config.yaml` in cwd (Docker default)
3. `../configs/config.yaml` (fallback)

### pnpm 11 allowBuilds

`frontend/pnpm-workspace.yaml` is **required** — it whitelists `esbuild`, `@parcel/watcher`, `vue-demi` install scripts. Without it, `pnpm run serve` fails with `ERR_PNPM_IGNORED_BUILDS`. Do not delete this file.

### Frontend API base URL

`frontend/.env.development` points to a remote server. For local dev, create `frontend/.env.development.local` with:
```
VITE_API_BASE_URL=http://localhost:18888/
```
Vite loads `.local` files with higher priority. The `*.local` pattern is gitignored.

### MySQL required

Backend will not start without a reachable MySQL 8 instance. On first boot, `AutoMigrate` (`backend/internal/database/migration.go`) creates all tables and seeds the default admin (`admin` / `admin@123`). Setup:

```bash
cp backend/configs/config.local.yaml.example backend/configs/config.local.yaml
# edit host, username, password
```

Redis is mentioned in design docs but is **not used** in code.

## Architecture

```
backend/
  cmd/main.go              # entrypoint
  internal/
    api/
      handlers/            # HTTP handlers (one per domain)
      middleware/           # auth, RBAC permission, audit logging, API key
      routes/router.go     # single router file
    config/                # Viper config loading
    database/              # connection, AutoMigrate, RBAC/system-config seed data
    models/                # request/response structs + GORM models
    repository/            # data access layer (GORM queries)
    service/               # business logic
  pkg/
    i18n/                  # backend error i18n
    logger/                # Logrus wrapper
    utils/                 # license file parsing, signing keys, slugs, product codes
  configs/                 # YAML configs, RSA keys, i18n error messages
  migrations/              # historical SQL (001–011), NOT runtime — AutoMigrate is authoritative

frontend/
  src/
    api/                   # axios-based API modules (one file per domain)
    views/                 # page components
    components/            # shared layout, charts, common widgets
    composables/           # Vue composition functions
    router/                # Vue Router config
    store/                 # Pinia stores
    i18n/locales/          # zh.json, en.json, ja.json
    assets/styles/         # global SCSS, variables, mixins (auto-injected via vite config)
  .env.development         # default API URL (remote)
  .env.development.local   # local override (gitignored, create this)
```

### Layering rules

- Handlers: HTTP parsing and response formatting only. No business logic.
- Services: business rules, validation, orchestration.
- Repositories: GORM queries, no display logic.
- Frontend API types must match backend response models exactly.
- New visible text must have zh/en/ja i18n entries.

## Testing

Backend has tests in `service`, `repository`, `middleware`, and `pkg/utils` packages. Tests use stub repositories (no DB dependency for unit tests). CI skips tests (`go test` not run in release workflows).

No frontend test files exist yet; vitest is configured but unused.

## Prettier (frontend)

```json
{
  "semi": false,
  "singleQuote": true,
  "trailingComma": "none",
  "printWidth": 100,
  "tabWidth": 2,
  "endOfLine": "lf",
  "vueIndentScriptAndStyle": false
}
```

## i18n

- Frontend: `frontend/src/i18n/locales/{zh,en,ja}.json` — keys grouped by module
- Backend errors: `backend/configs/i18n/errors/` — loaded at startup via `pkg/i18n`
- New UI strings: add entries to all three frontend locale files
- Backend business errors: use the project's i18n error system, not raw error strings

## Project docs

- `.ai/` — requirements, design specs, delivery reports, project conventions
- `.ai/_项目开发规范.md` — coding standards (layering, naming, error handling, logging)
- `.ai/_项目全局信息.md` — module inventory and doc index
- `backend/docs/install/local_dev_setup.md` — detailed local dev setup guide
- `docs/reference/` — enterprise edition UI screenshots (visual alignment baseline)