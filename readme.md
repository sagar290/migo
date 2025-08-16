<p align="center">
  <img src="docs/images/migo.png" style="max-width: 300px" alt="Migo a laravel style migration cli">
</p>

# Migo — Migration Tool for Go

**Migo** is a lightweight, Go-based database migration CLI and library inspired by tools like Laravel's migration system. It helps you version your schema changes, apply migrations, rollback, and manage migration files using a clean, intuitive workflow.

---
![Screenshot of Migo CLI](docs/images/screenshot2.png)

![Screenshot of Migo CLI](docs/images/screenshot1.png)

## 🛠 Features

- Define migrations using single-file templates with `[UP]` and `[DOWN]` blocks.
- Core commands:
    - `migo make "description"` — scaffold a new migration.
    - `migo up`, `migo down`, `migo refresh`, `migo fresh` — manage migrations.
- Supports flags like `--steps`, and `--dry-run`.
- Compatible with multiple SQL dialects (Postgres, MySQL, SQLite, SQL Server).
- Uses GORM under the hood; easy to integrate into your Go project.
- Zero-dependency CLI in a compact binary.

---

## 🚀 Getting Started

### Install

```bash
go install github.com/sagar290/migo@latest
```

Make sure `$GOPATH/bin` (or `$GOBIN`) is in your `PATH`.


### 📥 Quick Download Example (Linux) [Releases](https://github.com/sagar290/migo/releases/)
```
curl -L -o migo https://github.com/sagar290/migo/releases/download/<version_number>/migo-darwin-amd64
chmod +x migo
./migo --help
```
### Create a migration



```bash
migo make "create drivers table"
```

This generates a timestamped `.sql` file with UP/DOWN placeholders:

```sql
[UP]
-- Write your UP migration here
[/UP]

[DOWN]
-- Write your DOWN migration here
[/DOWN]
```

### Apply pending migrations

```bash
migo up
```

Use flags:

- `--steps=2` — apply only the next 2 migrations.
- `--dry-run` — preview what would run without executing.

### Roll back migrations

```bash
migo down
```

- By default rolls back the last batch.
- Use `--steps=1` to rollback only one migration.
- Supports `--dry-run`.

### Refresh the database

```bash
migo refresh
```

Rolls back and re-applies migrations in one command. Use flags for safety (`--dry-run`, `--force`, etc.).

### Fresh start (drop everything)

```bash
migo fresh
```

Drops all tables (except migrations history), then re-runs all migrations. Use with caution; supports `--dry-run`, `--seed`, `--force`.

---

## ⚙️ Configuration

Configure `migo` via `migo.yml`:

```yaml
migo:
  db_type: postgres
  db_url: postgres://username:password@localhost:5432/migoTest
  migrations_dir: ./migrations
  schema: public
  migration_table: migo_migrations
```

You can override settings via environment variables or flags when initializing the CLI.
The environment variable will looks like this
```
MIGO_DB_TYPE
MIGO_DB_URL
MIGO_MIGRATION_DIR
```
---

## 🤝 Contributing

1. Fork the project and create a feature branch
2. Write tests for new features or bugs
3. Submit a pull request with a clear explanation

---

## 📜 License

[MIT License](LICENSE)

