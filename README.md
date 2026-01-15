# Artisan - Database Migration Tool for Go

**Artisan** is a powerful, Laravel-inspired database migration tool for Go that provides an elegant and intuitive way to manage database migrations and seeders.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 🚀 Features

- ✅ **Multi-Database Support** - MySQL, PostgreSQL, SQLite
- ✅ **Batch Tracking** - Rollback migrations by batch, not one-by-one
- ✅ **SQL-Based Migrations** - Simple SQL files, no Go code needed
- ✅ **Multi-Statement Support** - Execute multiple SQL statements per migration
- ✅ **Laravel-Style Commands** - Familiar syntax for Laravel developers
- ✅ **Auto-Naming** - Smart migration name generation
- ✅ **Built-in Seeders** - Seed your database with test data
- ✅ **Driver-Specific SQL** - Auto-generate correct SQL for your database

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

Create a `.env` file:

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

### 2. Create Your First Migration

```bash
# Auto-generate migration name
artisan make:migration users

# Output: 1768501234_create_users_table
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
# ✓ Migrated: 1768501234_create_users_table
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

# Run migrations and seeders
artisan migrate --seed

# Rollback last batch (default: 1 step)
artisan migrate:rollback

# Rollback N batches
artisan migrate:rollback --step=3

# Rollback all migrations
artisan migrate:fresh
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

Artisan supports **MySQL**, **PostgreSQL**, and **SQLite** out of the box.

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

**SQLite:**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🆚 Artisan vs golang-migrate

| Feature | Artisan | golang-migrate |
|---------|---------|----------------|
| **Batch Tracking** | ✅ Yes | ❌ No |
| **Rollback by Batch** | ✅ Yes | ❌ No (one-by-one only) |
| **Multi-Statement Support** | ✅ Yes | ⚠️ Limited |
| **Auto-Naming** | ✅ Yes | ❌ No |
| **Built-in Seeders** | ✅ Yes | ❌ No |
| **Laravel-Style Commands** | ✅ Yes | ❌ No |
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
# Creates: 1768501234_create_products_table

artisan make:migration products add_price_column
# Creates: 1768501235_add_price_column
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
│   └── artisan              # Compiled binary
├── database/
│   ├── migrations/
│   │   ├── 1768501234_create_users_table
│   │   └── 1768501235_create_posts_table
│   └── seeders/
│       ├── users_seeder
│       └── posts_seeder
├── .env                     # Database configuration
└── Makefile                 # Build shortcuts
```

## 🛠️ Development

### Build from Source

```bash
git clone https://github.com/hymns/go-artisan.git
cd goartisan
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

- 📖 [Documentation](https://github.com/hymns/artisan/wiki)
- 🐛 [Issue Tracker](https://github.com/hymns/artisan/issues)
- 💬 [Discussions](https://github.com/hymns/artisan/discussions)

---

Made with ❤️ for Go developers who miss Laravel's migration system.
