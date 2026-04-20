# envsync

**Schema-driven environment variable management.** Validate, lint, and sync `.env` files against a schema with secret backend support.

> **Status: v0.2** — Lint, validate, sync (Vault), diff, and `--json` output all work end-to-end.

---

## Why envsync?

Every project has a `.env` file. Every project eventually has a bug because:

- A required variable was missing and the app crashed at runtime.
- A variable was misspelled and silently defaulted to the wrong value.
- A value was the wrong type (`PORT=three thousand`) and nothing noticed until prod.
- `.env.example` drifted out of sync with the real `.env`.
- Different developers had different values and nobody knew whose was right.

`envsync` gives your project a **schema** — a single source of truth for what env vars exist, what types they are, what values are allowed, and which are required. You commit the schema to git. Every developer and every CI run validates against it.

---

## Install

```bash
go install github.com/ja-guerrero/envsync@latest
```

Requires Go 1.21+.

---

## Quick start

**1. Create `.envsync.yaml` at your project root:**

```yaml
version: 1

project:
  name: myapp

vars:
  DATABASE_URL:
    required: true
    type: url
    secret: true
    description: "Postgres connection string"

  PORT:
    type: number
    min: 1
    max: 65535
    default: "3000"

  LOG_LEVEL:
    type: string
    enum: [debug, info, warn, error]
    default: info

  STRIPE_SECRET_KEY:
    required: true
    secret: true
    format: "^sk_(test|live)_[A-Za-z0-9]+$"

environments:
  staging:
    backend: vault-stag
    mount: app
    path: myapp/env
    kv_version: 2
```

**2. Create `~/.envsync/config.yaml` for backend auth (not committed to git):**

```yaml
version: 1

backends:
  - name: vault-stag
    type: vault
    addr: https://vault-stag.internal:8200
```

**3. Lint the schema:**

```bash
envsync lint
```

**4. Validate your `.env` against the schema:**

```bash
envsync validate
```

**5. Sync secrets from a backend:**

```bash
envsync sync --env staging
envsync sync --env staging --dry-run   # preview without writing
```

---

## Commands

### `envsync lint`

Validates that your schema file is well-formed.

```bash
envsync lint [--schema .envsync.yaml] [--json]
```

Catches: invalid YAML, unknown fields (typos like `requred: true`), invalid declarations (`min`/`max` on non-numeric types, `required` + `default`), invalid regex in `format`, unknown types.

### `envsync validate`

Validates a `.env` file against the schema.

```bash
envsync validate [--schema .envsync.yaml] [--env-file .env] [--json]
```

Catches: missing required variables, wrong types, values outside enum, regex mismatches, numeric values outside min/max.

### `envsync sync`

Pulls secrets from configured backends into a local `.env` file.

```bash
envsync sync --env <environment> [--schema .envsync.yaml] [--output .env] [--config ~/.envsync/config.yaml] [--dry-run] [--secrets-only]
```

- Resolves the backend for the given environment
- Fetches secret variables from the backend (Vault, etc.)
- Merges with schema defaults for non-secret variables
- Validates the result against the schema
- Writes to `.env` (or shows changes with `--dry-run`)

Missing keys in the backend are non-fatal — the validator reports them if they're `required`.

### `envsync diff`

Shows what `sync` would change.

```bash
envsync diff [--env-file .env] [--json] [--verbose]
```

Color-coded output: `+` added (green), `-` removed (red), `~` changed (yellow).

### Global flags

All commands support:

| Flag | Description |
|------|-------------|
| `--json` | Output as structured JSON |
| `--verbose` | Show detailed output |
| `--no-color` | Disable color output |

---

## Schema reference

### Top-level fields

```yaml
version: 1
project:
  name: myapp
vars:
  ...
environments:
  ...
```

### Variable fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `string`, `number`, `bool`, `url`, `json` |
| `required` | bool | Variable must be present |
| `default` | string | Default value (applied during sync if not fetched) |
| `enum` | list | Allowed values (strings only) |
| `format` | string | Regex the value must match |
| `min` | number | Minimum (requires `type: number`) |
| `max` | number | Maximum (requires `type: number`) |
| `secret` | bool | Fetched from backend during sync |
| `description` | string | Documentation |
| `source` | object | Per-variable backend override (see below) |

### Per-variable source

Variables can override which backend they pull from:

```yaml
vars:
  SPECIAL_KEY:
    secret: true
    source:
      backend: vault-prod
      path: shared/keys
      key: special_key    # defaults to variable name if omitted
```

Without `source`, secret variables use the environment-level backend.

### Environment fields

```yaml
environments:
  staging:
    backend: vault-stag   # references a backend in user config by name
    mount: app             # backend-specific params (merged with user config)
    path: myapp/env
    kv_version: 2
```

### Types

| Type | Accepts |
|------|---------|
| `string` | Any value |
| `number` | Integers and floats |
| `bool` | `true` or `false` |
| `url` | Absolute URLs (must have scheme and host) |
| `json` | Any valid JSON |

---

## User config

Backend credentials live in `~/.envsync/config.yaml` (not committed to git):

```yaml
version: 1

backends:
  - name: vault-stag
    type: vault
    addr: https://vault-stag.internal:8200

  - name: vault-prod
    type: vault
    addr: https://vault-prod.internal:8200
```

Token resolution for Vault: `token` param > `VAULT_TOKEN` env var > `~/.vault-token` file.

---

## `.env` file format

```bash
# Comments start with #
APP_NAME=myapp
PORT=3000

DATABASE_URL="postgres://user:pass@localhost/db"  # inline comments supported
API_KEY='sk_test_abc123'

export REDIS_URL=redis://localhost:6379   # export prefix stripped

FOO=bar#notcomment    # hash without preceding space is part of value
FOO=bar # comment     # hash with preceding space starts a comment
```

**Supported:** `KEY=value`, comments, blank lines, `export` prefix, double-quoted values (escapes: `\n`, `\t`, `\"`, `\\`), single-quoted values (literal), inline comments.

**Rejected:** lines without `=`, empty keys, keys not matching `[A-Za-z_][A-Za-z0-9_]*`, mismatched quotes, duplicate keys, invalid escape sequences.

---

## CI integration

### GitHub Actions

```yaml
- name: Validate environment
  run: |
    go install github.com/ja-guerrero/envsync@latest
    envsync lint
    envsync validate --env-file .env.ci
```

### JSON output for automation

```bash
envsync validate --json | jq '.violations[]'
```

### Pre-commit hook

```bash
#!/bin/sh
envsync lint && envsync validate
```

---

## Roadmap

**Planned:**

- Per-variable source model redesign (environment-scoped backends with per-secret overrides)
- Additional backends: AWS Secrets Manager, AWS SSM, 1Password
- `envsync init` — scaffold `.envsync.yaml` from an existing `.env`
- Environment-scoped requirements (`required_in: [staging, production]`)
- Conditional requirements (`required_if: { CACHE_BACKEND: redis }`)

**Not planned:**

- Runtime injection into `os.Environ`
- Variable interpolation (`${OTHER_VAR}`)

---

## Development

```bash
git clone https://github.com/ja-guerrero/envsync
cd envsync

go test ./...
go build -o envsync .
```

### Project layout

```
envsync/
├── cmd/
│   ├── cli.go          # root command, global flags
│   ├── flags.go        # shared flag variables
│   ├── colors.go       # shared color styles
│   ├── output.go       # JSON output helpers
│   ├── sync.go         # sync command
│   ├── validate.go     # validate command
│   ├── lint.go         # lint command
│   └── diff.go         # diff command
├── internal/
│   ├── envfile/        # .env parsing and writing
│   ├── config/         # .envsync.yaml + user config loading
│   ├── schema/         # variable schema and validation
│   └── backend/        # Backend interface, Vault implementation
├── main.go
└── README.md
```

---

## License

MIT

---

## Contributing

Issues and PRs welcome. The most valuable contributions right now are:

1. Real-world `.env` files that trip up the parser (open an issue with a reproduction).
2. Feedback on schema expressiveness: what do you wish you could declare that you can't?
3. Typo fixes in docs.

Please open an issue before starting on a large feature.
