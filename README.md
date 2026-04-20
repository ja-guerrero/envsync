# envsync

**A schema-driven linter and validator for `.env` files.**

Declare what environment variables your project needs, enforce them in CI, and catch typos, missing values, and wrong types before they hit production.

> **Status: early development (v0.1).** The linter and validator work end-to-end. Backends, multi-environment support, and sync commands are on the roadmap below.

---

## Why envsync?

Every project has a `.env` file. Every project eventually has a bug because:

- A required variable was missing and the app crashed at runtime.
- A variable was misspelled and silently defaulted to the wrong value.
- A value was the wrong type (`PORT=three thousand`) and nothing noticed until prod.
- `.env.example` drifted out of sync with the real `.env`.
- Different developers had different values and nobody knew whose was right.

`envsync` gives your project a **schema** — a single source of truth for what env vars exist, what types they are, what values are allowed, and which are required. You commit the schema to git. Every developer and every CI run validates against it.

Think of it as a type checker for your environment.

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
```

**2. Lint the schema itself:**

```bash
envsync lint
```

Confirms your `.envsync.yaml` is syntactically valid and internally consistent (no contradictory rules, valid regexes, known types, etc.).

**3. Validate your `.env` against the schema:**

```bash
envsync validate
```

Reads `.env` from the current directory, checks every variable against the schema, and reports any violations.

If everything is good, exits silently with status 0. If anything fails, prints a list of issues and exits non-zero — perfect for CI and pre-commit hooks.

---

## Commands

### `envsync lint`

Validates that your schema file (`.envsync.yaml`) is well-formed.

```bash
envsync lint [--config path/to/.envsync.yaml]
```

Catches:

- Invalid YAML syntax
- Unknown fields (catches typos like `requred: true`)
- Invalid declarations (e.g. `min`/`max` on non-numeric types, `enum` on non-string types, `required` combined with `default`)
- Invalid regex in `format:`
- Unknown types

### `envsync validate`

Validates a `.env` file against the schema.

```bash
envsync validate [--config path/to/.envsync.yaml] [--env-file path/to/.env]
```

Catches:

- Required variables that are missing
- Values that don't match the declared type
- Values outside the declared enum
- Values that don't match the declared regex
- Numeric values outside min/max bounds

---

## Schema reference

### Top-level fields

```yaml
version: 1 # Required. Schema version.
project: # Optional.
  name: myapp # Project name, used in error messages.
vars: # Map of variable declarations.
  ...
```

### Variable fields

Every variable can declare any subset of these:

| Field         | Type   | Description                                                                           |
| ------------- | ------ | ------------------------------------------------------------------------------------- |
| `type`        | string | Value type: `string` (default), `number`, `bool`, `url`, `json`                       |
| `required`    | bool   | If true, the variable must be present                                                 |
| `default`     | string | Default value (for documentation/future features; not yet applied at validation time) |
| `enum`        | list   | Allowed values (strings only)                                                         |
| `format`      | string | Regex the value must match                                                            |
| `min`         | number | Minimum numeric value (requires `type: number`)                                       |
| `max`         | number | Maximum numeric value (requires `type: number`)                                       |
| `secret`      | bool   | Marks the variable as sensitive (affects display in future commands)                  |
| `description` | string | Free-text description for documentation                                               |

### Types

| Type     | Accepts                                                                             |
| -------- | ----------------------------------------------------------------------------------- |
| `string` | Any value                                                                           |
| `number` | Anything `strconv.ParseFloat` accepts (integers and floats)                         |
| `bool`   | `true`, `false`, `1`, `0`, `T`, `F`, `TRUE`, `FALSE` (via Go's `strconv.ParseBool`) |
| `url`    | Absolute URLs (must have scheme and host)                                           |
| `json`   | Any valid JSON                                                                      |

### Example

```yaml
version: 1

vars:
  # Required, must be a valid URL
  DATABASE_URL:
    required: true
    type: url

  # Optional number with bounds
  PORT:
    type: number
    min: 1
    max: 65535

  # Enum with default
  LOG_LEVEL:
    enum: [debug, info, warn, error]
    default: info

  # Regex-validated
  SEMVER:
    format: "^\\d+\\.\\d+\\.\\d+$"

  # Free-form with documentation
  APP_NAME:
    description: "Used in logs and error reports"
```

---

## `.env` file format

`envsync` accepts standard `.env` syntax:

```bash
# Comments start with #
APP_NAME=myapp
PORT=3000

# Quoted values
DATABASE_URL="postgres://user:pass@localhost/db"
API_KEY='sk_test_abc123'

# Export prefix is stripped
export REDIS_URL=redis://localhost:6379

# Blank lines are ignored

FEATURE_FLAG=
```

**Supported:**

- `KEY=value` pairs
- Comments (`#` at start of line)
- Blank lines
- `export ` prefix (stripped on read)
- Double-quoted values (with Go escape sequences: `\n`, `\t`, etc.)
- Single-quoted values (literal, no escapes)

**Rejected with a clear error:**

- Lines without `=`
- Empty keys
- Keys with whitespace
- Mismatched quotes
- Duplicate keys

---

## CI integration

### GitHub Actions

```yaml
- name: Validate environment
  run: |
    go install github.com/ja-guerrero/envsync@latest
    envsync validate --env-file .env.ci
```

### Pre-commit hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/sh
envsync lint && envsync validate
```

---

## Roadmap

**In progress / planned:**

- `envsync init` — scaffold `.envsync.yaml` from an existing `.env`
- `envsync diff` — compare two `.env` files against the schema
- Environment-scoped requirements (`required_in: [staging, production]`)
- Conditional requirements (`required_if: { CACHE_BACKEND: redis }`)
- Default value application (`envsync apply` or similar)
- Secret backends: HashiCorp Vault, 1Password, AWS Secrets Manager, Doppler
- User-level config at `~/.config/envsync/config.yaml` for backend auth
- `envsync sync` — pull secrets from a backend into a local `.env`
- Line numbers in error messages (currently reports variable name only)

**Not planned:**

- Runtime injection into `os.Environ` (use `dotenv` or your app's built-in env loading)
- Variable interpolation inside `.env` values (`${OTHER_VAR}`)

---

## Development

```bash
# Clone
git clone https://github.com/ja-guerrero/envsync
cd envsync

# Run tests
go test ./...

# Build
go build -o envsync .

# Run
./envsync lint
./envsync validate
```

### Project layout

```
envsync/
├── cmd/                      # Cobra commands
│   ├── root.go
│   ├── lint.go
│   └── validate.go
├── internal/
│   ├── config/               # .envsync.yaml loading and validation
│   ├── schema/               # Variable schema and value validation
│   └── envfile/              # .env file parsing
├── main.go
└── README.md
```

---

## License

MIT

---

## Contributing

Issues and PRs welcome. This project is early — the most valuable contributions right now are:

1. Real-world `.env` files that trip up the parser (open an issue with a reproduction).
2. Feedback on schema expressiveness: what do you wish you could declare that you can't?
3. Typo fixes in docs.

Please open an issue before starting on a large feature.
