# Multi-Database Support

Artisan supports multiple database drivers out of the box.

## Supported Databases

- ✅ **MySQL** (default)
- ✅ **PostgreSQL**
- ✅ **SQLite**

## Configuration

Update your `.env` file with the appropriate database configuration:

### MySQL

```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_NAME=your_database
DB_USER=root
DB_PASS=your_password
```

### PostgreSQL

```env
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=your_database
DB_USER=postgres
DB_PASS=your_password
```

### SQLite

```env
DB_DRIVER=sqlite3
DB_NAME=./database.db
```

Note: For SQLite, `DB_NAME` is the path to the database file. Other connection parameters (HOST, PORT, USER, PASS) are not needed.

## Driver-Specific SQL

Artisan automatically generates database-specific SQL syntax:

### MySQL
- Primary Key: `INTEGER PRIMARY KEY AUTO_INCREMENT`
- Timestamps: `TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`

### PostgreSQL
- Primary Key: `SERIAL PRIMARY KEY`
- Timestamps: `TIMESTAMP DEFAULT CURRENT_TIMESTAMP`

### SQLite
- Primary Key: `INTEGER PRIMARY KEY AUTOINCREMENT`
- Timestamps: `TIMESTAMP DEFAULT CURRENT_TIMESTAMP`

## Example Migration

When you create a migration, the SQL will be generated based on your current `DB_DRIVER`:

```bash
# With DB_DRIVER=mysql
./bin/artisan make:migration users

# Generated SQL:
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

```bash
# With DB_DRIVER=postgres
./bin/artisan make:migration users

# Generated SQL:
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Switching Databases

To switch databases:

1. Update your `.env` file with new database configuration
2. Existing migrations will work as-is (they contain raw SQL)
3. New migrations will be generated with the new database syntax

## Testing with SQLite

SQLite is great for testing and development:

```bash
# Create a test database
echo "DB_DRIVER=sqlite3" > .env.test
echo "DB_NAME=./test.db" >> .env.test

# Run migrations
./bin/artisan migrate

# Your migrations will run on SQLite!
```

## Notes

- Migration files contain raw SQL, so they are database-specific
- If you switch databases, you may need to adjust existing migration SQL
- The `migrations` table schema is automatically adjusted for each database
- All drivers use the same migration tracking system (batch-based)
