# Artisan - Database Migration Tool for Go

**Artisan** is a powerful, Laravel-inspired database migration tool for Go that provides an elegant and intuitive way to manage database migrations and seeders.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Commands](#-commands)
- [New in v1.2.0](#-new-in-v120)
- [Multi-Database Support](#️-multi-database-support)
- [Programmatic Usage](#-programmatic-usage)
- [Comparison with golang-migrate](#-artisan-vs-golang-migrate)
- [Advanced Usage](#-advanced-usage)
- [Project Structure](#️-project-structure)
- [Development](#️-development)
- [Author](#-author)
- [License](#-license)

## Features

- ✅ **Multi-Database Support** - MySQL, PostgreSQL, SQL Server, SQLite
- ✅ **Batch Tracking** - Rollback migrations by batch, not one-by-one
- ✅ **Transaction Safety** - Each migration runs in a transaction (atomic)
- ✅ **Migration Locking** - Prevents concurrent migrations
- ✅ **SQL-Based Migrations** - Simple SQL files, no Go code needed
- ✅ **Multi-Statement Support** - Execute multiple SQL statements per migration
- ✅ **Laravel-Style Commands** - Familiar syntax for Laravel developers
- ✅ **Auto-Naming** - Smart migration name generation
- ✅ **Built-in Seeders** - Seed your database with test data
- ✅ **Driver-Specific SQL** - Auto-generate correct SQL for your database
- ✅ **Migration Status** - See which migrations are pending/ran
- ✅ **Dry Run Mode** - Preview migrations before running

## 📦 Installation

```bash
go get github.com/hymns/go-artisan
```

Or clone and build:

```bash
git clone https://github.com/hymns/go-artisan.git
cd go-artisan
make build
```

Install globally:

```bash
make install
# or
sudo cp bin/artisan /usr/local/bin/
```

## 🎯 Quick Start

### 1. Setup Configuration

Copy the example configuration file and customize it:

```bash
# Copy example configuration
cp .env.example .env

# Edit .env with your database credentials
```

Example `.env` configuration:

```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_NAME=your_database
DB_USER=root
DB_PASS=

MIGRATIONS_PATH=./database/migrations
SEEDERS_PATH=./database/seeders
```

> **Note:** See [Multi-Database Support](#️-multi-database-support) section for PostgreSQL, SQL Server, and SQLite configuration.

### 2. Create Your First Migration

```bash
# Auto-generate migration name
artisan make:migration users

# Output: 2026_01_16_170530_create_users_table
```

### 3. Edit the Migration

```sql
-- Migration: create_users_table
-- Database: mysql

--UP--
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

--DOWN--
DROP TABLE IF EXISTS users;
```

### 4. Run Migrations

```bash
artisan migrate
# ✓ Migrated: 2026_01_16_170530_create_users_table
```

### 5. Create and Run Seeders

```bash
artisan make:seeder users
# ✓ Seeder created: users_seeder

# Edit database/seeders/users_seeder
# Then run:
artisan db:seed
# ✓ Seeded: users_seeder
```

## 📚 Commands

### Migration Commands

```bash
# Create migration with auto-naming
artisan make:migration users
artisan make:migration products

# Create migration with custom name
artisan make:migration users create_users_table

# Using flags
artisan make:migration --table=users custom_name

# Run all pending migrations
artisan migrate

# Run specific migration file
artisan migrate --path=./database/migrations/2026_01_16_170530_create_users_table

# Run migrations and seeders
artisan migrate --seed

# Rollback last batch (default: 1 step)
artisan migrate:rollback

# Rollback N batches
artisan migrate:rollback --step=3

# Rollback all, then re-run migrations (fresh start)
artisan migrate:fresh

# Rollback all, migrate, then seed (fresh system)
artisan migrate:fresh --seed

# Show migration status
artisan migrate:status

# Preview pending migrations (dry run)
artisan migrate:dry-run
```

### Seeder Commands

```bash
# Create seeder (auto-appends _seeder)
artisan make:seeder users
# Output: users_seeder

# Using flags
artisan make:seeder --seeder=products

# Run all seeders
artisan db:seed

# Run specific seeder file
artisan db:seed --path=./database/seeders/users_seeder
```

### Makefile Shortcuts

```bash
make build      # Build binary
make migrate    # Auto-build + migrate
make rollback   # Auto-build + rollback
make seed       # Auto-build + seed
make install    # Install globally
```

## 🗄️ Multi-Database Support

Artisan supports **MySQL**, **PostgreSQL**, **SQL Server**, and **SQLite** out of the box.

### MySQL Configuration

```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_NAME=mydb
DB_USER=root
DB_PASS=secret
```

### PostgreSQL Configuration

```env
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mydb
DB_USER=postgres
DB_PASS=secret
```

### SQL Server Configuration

```env
DB_DRIVER=sqlserver
DB_HOST=localhost
DB_PORT=1433
DB_NAME=mydb
DB_USER=sa
DB_PASS=secret
```

### SQLite Configuration

```env
DB_DRIVER=sqlite3
DB_NAME=./database.db
```

### Driver-Specific SQL Generation

Artisan automatically generates the correct SQL syntax for your database:

**MySQL:**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**PostgreSQL:**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**SQL Server:**
```sql
CREATE TABLE users (
    id INT IDENTITY(1,1) PRIMARY KEY,
    created_at DATETIME DEFAULT GETDATE(),
    updated_at DATETIME DEFAULT GETDATE()
);
```

**SQLite:**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🎉 New in v1.2.0

### Transaction Safety

All migrations and seeders now run in transactions for atomic operations:

```go
// Each migration runs in a transaction
// If any statement fails, entire migration rolls back
artisan migrate
```

**Benefits:**
- ✅ All-or-nothing execution
- ✅ Automatic rollback on error
- ✅ Data integrity guaranteed

### Migration Locking

Prevents concurrent migrations from running simultaneously:

```bash
# Terminal 1
artisan migrate
# Running...

# Terminal 2
artisan migrate
# ✗ migration is already running by another process
```

### Migration Status Command

See which migrations have been run and which are pending:

```bash
artisan migrate:status
```

**Output:**
```
Migration Status:

Migration                                          Batch      Ran
----------------------------------------------------------------------
2026_01_16_170530_create_users_table              1          YES
2026_01_16_170545_create_posts_table              1          YES
2026_01_16_180230_add_user_roles                  -          NO
```

- **Green YES** = Migration has been run
- **Yellow NO** = Migration is pending
- **Batch** = Which batch the migration was run in (for rollback)

### Dry Run Mode

Preview what migrations will run without executing them:

```bash
artisan migrate:dry-run
```

**Output:**
```
=== Dry Run - No changes will be made ===

Would migrate: 2026_01_16_180230_add_user_roles (Batch 2)
  Statement 1: ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user'...
  Statement 2: CREATE INDEX idx_users_role ON users(role)...

Total pending migrations: 1
```

### Database Placeholder Compatibility

Fixed SQL placeholder compatibility for all databases:
- MySQL/SQLite: `?`
- PostgreSQL: `$1, $2, $3`
- SQL Server: `@p1, @p2, @p3`

### Human-Readable Migration Names

Migration filenames now use date-based format:

**Old:** `1768501234_create_users_table`  
**New:** `2026_01_16_170530_create_users_table`

Format: `YYYY_MM_DD_HHMMSS_migration_name`

**Benefits:**
- ✅ Easy to identify when migration was created
- ✅ Human-readable at a glance
- ✅ Still sortable chronologically

---

## 🆚 Artisan vs golang-migrate

| Feature | Artisan v1.2.0 | golang-migrate |
|---------|----------------|----------------|
| **Batch Tracking** | ✅ Yes | ❌ No |
| **Rollback by Batch** | ✅ Yes | ❌ No (one-by-one only) |
| **Transaction Safety** | ✅ Yes | ✅ Yes |
| **Migration Locking** | ✅ Yes | ✅ Yes |
| **Migration Status** | ✅ Yes | ❌ No |
| **Dry Run Mode** | ✅ Yes | ❌ No |
| **Multi-Statement Support** | ✅ Yes | ⚠️ Limited |
| **Auto-Naming** | ✅ Yes | ❌ No |
| **Built-in Seeders** | ✅ Yes | ❌ No |
| **Laravel-Style Commands** | ✅ Yes | ❌ No |
| **Human-Readable Filenames** | ✅ Yes (date-based) | ❌ No (Unix timestamp) |
| **Driver-Specific Templates** | ✅ Yes | ❌ No |
| **Multi-Database Support** | ✅ MySQL, Postgres, SQLite | ✅ Many drivers |
| **SQL-Based** | ✅ Pure SQL | ✅ Pure SQL |
| **Migration Format** | Simple text files | Up/Down separate files |

### Why Choose Artisan?

#### 1. **Batch Tracking System**

**Artisan:**
```bash
artisan migrate
# Batch 1: users, posts, comments

artisan migrate:rollback
# Rolls back entire Batch 1 (all 3 migrations)
```

**golang-migrate:**
```bash
migrate up
# Migration 1, 2, 3

migrate down 1
# Only rolls back migration 3
# Need to run 3 times to rollback all
```

#### 2. **Multi-Statement Migrations**

**Artisan:**
```sql
--UP--
CREATE TABLE users (...);
CREATE TABLE posts (...);
CREATE INDEX idx_users_email ON users(email);
INSERT INTO users (name) VALUES ('Admin');

--DOWN--
DROP TABLE posts;
DROP TABLE users;
```

All statements execute in one migration file!

#### 3. **Laravel-Style Workflow**

**Artisan:**
```bash
artisan make:migration users          # Auto-names: create_users_table
artisan migrate                       # Run migrations
artisan migrate:rollback --step=2     # Rollback 2 batches
artisan make:seeder users             # Create seeder
artisan db:seed                       # Run seeders
```

**golang-migrate:**
```bash
migrate create -ext sql -dir migrations create_users_table
migrate -path migrations -database "mysql://..." up
migrate -path migrations -database "mysql://..." down 1
# No built-in seeder support
```

#### 4. **Smart Auto-Naming**

**Artisan:**
```bash
artisan make:migration products
# Creates: 2026_01_16_170530_create_products_table

artisan make:migration products add_price_column
# Creates: 2026_01_16_170545_add_price_column
```

#### 5. **Built-in Seeder System**

**Artisan:**
```bash
artisan make:seeder products
# Edit SQL file
artisan db:seed
```

**golang-migrate:**
No built-in seeder support. Need separate tools or custom scripts.

## 🔧 Programmatic Usage

### Auto-Migration on Application Startup

You can integrate go-artisan into your Go application to automatically run migrations when your app starts:

```go
package main

import (
    "database/sql"
    "log"
    
    _ "github.com/go-sql-driver/mysql"
    "github.com/hymns/go-artisan/migration"
)

func main() {
    // Connect to database
    db, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/dbname?parseTime=true")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Auto-run migrations on startup (safe - idempotent)
    m := migration.New(db)
    if err := m.AutoMigrate("./database/migrations"); err != nil {
        log.Fatal(err)
    }
    
    // Auto-run seeders on startup (safe - with tracking)
    s := seeder.New(db)
    if err := s.AutoSeed("./database/seeders"); err != nil {
        log.Fatal(err)
    }
    
    // Your application code here...
}
```

**Key Features:**
- ✅ **Silent Execution** - No console output, only returns errors
- ✅ **Production Ready** - Safe to run on every app startup
- ✅ **Idempotent** - Only runs pending migrations and seeders
- ✅ **Fast** - Skips already migrated/seeded files
- ✅ **Seeder Tracking** - Tracks seeded files to prevent duplicates

> ✅ **Safe:** Both `AutoMigrate` and `AutoSeed` use tracking tables to prevent duplicates. Safe to run on every app startup!

### Migration Methods

**`AutoMigrate(path string)`** - Silent migration for production apps
```go
if err := m.AutoMigrate("./database/migrations"); err != nil {
    log.Fatal(err)
}
```

**`Migrate(path string)`** - Migration with colored console output
```go
if err := m.Migrate("./database/migrations"); err != nil {
    log.Fatal(err)
}
// Output: ✓ Migrated: 2026_01_16_170530_create_users_table
```

**`MigrateFile(filePath string)`** - Run specific migration file
```go
// Run a specific migration file
if err := m.MigrateFile("./database/migrations/2026_01_16_170530_create_users_table"); err != nil {
    log.Fatal(err)
}
// Output: ✓ Migrated: 2026_01_16_170530_create_users_table
```

**Use Cases for `MigrateFile`:**
- Testing specific migrations in development
- Running hotfix migrations in production
- Conditional migration execution based on app logic
- Debugging migration issues

### Seeder Methods

**`AutoSeed(path string)`** - Silent seeding with tracking (NEW!)
```go
// Safe to run on every app startup - tracks seeded files
s := seeder.New(db)
if err := s.AutoSeed("./database/seeders"); err != nil {
    log.Fatal(err)
}
```

**Features:**
- ✅ **Idempotent** - Skips already-seeded files
- ✅ **Tracking Table** - Uses `seeders` table to track execution
- ✅ **Silent** - No console output
- ✅ **Production Safe** - Won't duplicate data on restart

**`RunWithTracking(path string)`** - Run with tracking and colored output
```go
// Shows which seeders are skipped vs newly seeded
s := seeder.New(db)
if err := s.RunWithTracking("./database/seeders"); err != nil {
    log.Fatal(err)
}
// Output: ⚠ Already seeded: users_seeder
// Output: ✓ Seeded: products_seeder
```

**`Run(path string)`** - Run all seeders WITHOUT tracking
```go
// ⚠️ WARNING: Will duplicate data if run multiple times
s := seeder.New(db)
if err := s.Run("./database/seeders"); err != nil {
    log.Fatal(err)
}
// Output: ✓ Seeded: users_seeder
```

**`RunFile(filePath string)`** - Run specific seeder file WITHOUT tracking
```go
// ⚠️ WARNING: Will duplicate data if run multiple times
s := seeder.New(db)
if err := s.RunFile("./database/seeders/users_seeder"); err != nil {
    log.Fatal(err)
}
// Output: ✓ Seeded: users_seeder
```

**Use Cases:**
- **`AutoSeed`** - Production apps, auto-run on startup
- **`RunWithTracking`** - Development, see what's already seeded
- **`Run`** - One-time seeding, test data that can be wiped
- **`RunFile`** - Specific seeder for testing/debugging

## 📖 Advanced Usage

### Complex Migrations

```sql
--UP--
-- Create main table
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    user_id INTEGER NOT NULL,
    total DECIMAL(10,2),
    status ENUM('pending', 'completed', 'cancelled'),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create related table
CREATE TABLE order_items (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    order_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER DEFAULT 1,
    price DECIMAL(10,2),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- Add indexes
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_order ON order_items(order_id);

--DOWN--
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
```

### Seeders with Multiple Statements

```sql
-- Seeder: products_seeder

INSERT INTO products (name, price) VALUES 
    ('Product 1', 99.99),
    ('Product 2', 149.99),
    ('Product 3', 199.99);

INSERT INTO categories (name) VALUES 
    ('Electronics'),
    ('Books'),
    ('Clothing');

UPDATE products SET category_id = 1 WHERE id = 1;
```

### Environment-Specific Configurations

```bash
# Development
cp .env.example .env.dev
# Set DB_DRIVER=sqlite3, DB_NAME=./dev.db

# Testing
cp .env.example .env.test
# Set DB_DRIVER=sqlite3, DB_NAME=./test.db

# Production
cp .env.example .env.prod
# Set DB_DRIVER=postgres with production credentials
```

## 🏗️ Project Structure

```
your-project/
├── bin/
│   └── artisan              # Compiled binary (after make build)
├── database/
│   ├── migrations/          # Migration files (created via make:migration)
│   │   └── ...              # e.g., 1768501234_create_users_table
│   └── seeders/             # Seeder files (created via make:seeder)
│       └── ...              # e.g., users_seeder
├── .env                     # Database configuration (copy from .env.example)
└── Makefile                 # Build shortcuts (optional)
```

## 🛠️ Development

### Build from Source

```bash
git clone https://github.com/hymns/go-artisan.git
cd go-artisan
make deps    # Install dependencies
make build   # Build binary
make test    # Run tests
```

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 👨‍💻 Author

**Muhammad Hamizi Jaminan**

- 📧 Email: [hello@hamizi.net](mailto:hello@hamizi.net)
- 🐙 GitHub: [@hymns](https://github.com/hymns)
- 🌐 Website: [hamizi.net](https://hamizi.net)

A passionate Go developer who loves building developer tools and bringing the best of Laravel's ecosystem to the Go community.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Inspired by:
- [Laravel Artisan](https://laravel.com/docs/artisan)
- [golang-migrate](https://github.com/golang-migrate/migrate)

## 📞 Support

- 📖 [Documentation](https://github.com/hymns/go-artisan/wiki)
- 🐛 [Issue Tracker](https://github.com/hymns/go-artisan/issues)
- 💬 [Discussions](https://github.com/hymns/go-artisan/discussions)

---

Made with ❤️ for Go developers who miss Laravel's migration system.
