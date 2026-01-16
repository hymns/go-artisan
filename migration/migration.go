package migration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Migration struct {
	DB     *sql.DB
	Driver string
}

func New(db *sql.DB) *Migration {
	return &Migration{
		DB:     db,
		Driver: getDBDriver(db),
	}
}

func getDBDriver(db *sql.DB) string {
	if db == nil {
		return os.Getenv("DB_DRIVER")
	}
	// Try to detect driver from connection
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "mysql" // default
	}
	return driver
}

func (m *Migration) EnsureMigrationsTable() error {
	var query string

	switch m.Driver {
	case "postgres":
		query = `CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			migration VARCHAR(255) NOT NULL,
			batch INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	case "sqlite", "sqlite3":
		query = `CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			migration VARCHAR(255) NOT NULL,
			batch INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	case "sqlserver", "mssql":
		query = `IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='migrations' AND xtype='U')
			CREATE TABLE migrations (
				id INT IDENTITY(1,1) PRIMARY KEY,
				migration VARCHAR(255) NOT NULL,
				batch INT NOT NULL,
				created_at DATETIME DEFAULT GETDATE()
			)`
	default: // mysql
		query = `CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			migration VARCHAR(255) NOT NULL,
			batch INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	}

	_, err := m.DB.Exec(query)
	return err
}

func (m *Migration) Migrate(migrationsPath string) error {
	if err := m.EnsureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	migrated, err := m.getMigrated()
	if err != nil {
		return fmt.Errorf("failed to get migrated list: %w", err)
	}

	batch, err := m.getNextBatch()
	if err != nil {
		return fmt.Errorf("failed to get next batch: %w", err)
	}

	files, err := m.getMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	executed := 0
	for _, file := range files {
		name := filepath.Base(file)

		if contains(migrated, name) {
			continue
		}

		// Read and parse SQL file
		statements, err := m.parseMigrationSQL(file, true) // true = UP
		if err != nil {
			return fmt.Errorf("failed to parse migration %s: %w", name, err)
		}

		// Execute each SQL statement
		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			if _, err := m.DB.Exec(stmt); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", name, err)
			}
		}

		if err := m.recordMigration(name, batch); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}

		color.Green("✓ Migrated: %s", name)
		executed++
	}

	if executed == 0 {
		color.Cyan("Nothing to migrate.")
	}

	return nil
}

func (m *Migration) Rollback(migrationsPath string) error {
	batch, err := m.getLastBatch()
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}

	if batch == 0 {
		color.Cyan("Nothing to rollback.")
		return nil
	}

	files, err := m.getBatchMigrations(batch)
	if err != nil {
		return fmt.Errorf("failed to get batch migrations: %w", err)
	}

	for i := len(files) - 1; i >= 0; i-- {
		name := files[i]
		filePath := filepath.Join(migrationsPath, name)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File doesn't exist, just remove from database
			color.Yellow("⚠ Migration file not found, removing record: %s", name)
			if err := m.deleteMigration(name); err != nil {
				return fmt.Errorf("failed to delete migration record %s: %w", name, err)
			}
			continue
		}

		// Read and parse SQL file
		statements, err := m.parseMigrationSQL(filePath, false) // false = DOWN
		if err != nil {
			return fmt.Errorf("failed to parse migration %s: %w", name, err)
		}

		// Execute each SQL statement
		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			if _, err := m.DB.Exec(stmt); err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", name, err)
			}
		}

		if err := m.deleteMigration(name); err != nil {
			return fmt.Errorf("failed to delete migration record %s: %w", name, err)
		}

		color.Green("✓ Rolled back: %s", name)
	}

	return nil
}

func (m *Migration) getMigrated() ([]string, error) {
	rows, err := m.DB.Query("SELECT migration FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrated []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrated = append(migrated, name)
	}

	return migrated, rows.Err()
}

func (m *Migration) recordMigration(name string, batch int) error {
	_, err := m.DB.Exec("INSERT INTO migrations (migration, batch) VALUES (?, ?)", name, batch)
	return err
}

func (m *Migration) getNextBatch() (int, error) {
	var batch sql.NullInt64
	err := m.DB.QueryRow("SELECT MAX(batch) FROM migrations").Scan(&batch)
	if err != nil {
		return 0, err
	}

	if !batch.Valid {
		return 1, nil
	}

	return int(batch.Int64) + 1, nil
}

func (m *Migration) getLastBatch() (int, error) {
	var batch sql.NullInt64
	err := m.DB.QueryRow("SELECT MAX(batch) FROM migrations").Scan(&batch)
	if err != nil {
		return 0, err
	}

	if !batch.Valid {
		return 0, nil
	}

	return int(batch.Int64), nil
}

func (m *Migration) GetLastBatch() (int, error) {
	return m.getLastBatch()
}

func (m *Migration) getBatchMigrations(batch int) ([]string, error) {
	rows, err := m.DB.Query("SELECT migration FROM migrations WHERE batch = ? ORDER BY id DESC", batch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrations = append(migrations, name)
	}

	return migrations, rows.Err()
}

func (m *Migration) deleteMigration(name string) error {
	_, err := m.DB.Exec("DELETE FROM migrations WHERE migration = ?", name)
	return err
}

func (m *Migration) getMigrationFiles(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip hidden files, registry.go, and .go files (old format)
		if strings.HasPrefix(name, ".") || name == "registry.go" || strings.HasSuffix(name, ".go") {
			continue
		}

		files = append(files, filepath.Join(path, name))
	}

	sort.Strings(files)
	return files, nil
}

func (m *Migration) MakeMigration(tableName, migrationName, migrationsPath string) error {
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, migrationName)
	filepath := filepath.Join(migrationsPath, filename)

	template := m.getMigrationTemplate(tableName, migrationName)

	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	if err := os.WriteFile(filepath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	color.Green("✓ Migration created: %s", filename)
	return nil
}

func (m *Migration) getMigrationTemplate(tableName, migrationName string) string {
	var upSQL, downSQL string

	switch m.Driver {
	case "postgres":
		upSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`, tableName)
		downSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)

	case "sqlite", "sqlite3":
		upSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`, tableName)
		downSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)

	case "sqlserver", "mssql":
		upSQL = fmt.Sprintf(`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='%s' AND xtype='U')
    CREATE TABLE %s (
        id INT IDENTITY(1,1) PRIMARY KEY,
        created_at DATETIME DEFAULT GETDATE(),
        updated_at DATETIME DEFAULT GETDATE()
    );`, tableName, tableName)
		downSQL = fmt.Sprintf("IF EXISTS (SELECT * FROM sysobjects WHERE name='%s' AND xtype='U') DROP TABLE %s;", tableName, tableName)

	default: // mysql
		upSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);`, tableName)
		downSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
	}

	return fmt.Sprintf(`-- Migration: %s
-- Created at: %s
-- Database: %s

--UP--
%s

--DOWN--
%s
`, migrationName, time.Now().Format("2006-01-02 15:04:05"), m.Driver, upSQL, downSQL)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (m *Migration) AutoMigrate(migrationsPath string) error {
	if err := m.EnsureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	migrated, err := m.getMigrated()
	if err != nil {
		return fmt.Errorf("failed to get migrated list: %w", err)
	}

	batch, err := m.getNextBatch()
	if err != nil {
		return fmt.Errorf("failed to get next batch: %w", err)
	}

	files, err := m.getMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	executed := 0
	for _, file := range files {
		name := filepath.Base(file)

		if contains(migrated, name) {
			continue
		}

		statements, err := m.parseMigrationSQL(file, true)
		if err != nil {
			return fmt.Errorf("failed to parse migration %s: %w", name, err)
		}

		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			if _, err := m.DB.Exec(stmt); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", name, err)
			}
		}

		if err := m.recordMigration(name, batch); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}

		executed++
	}

	return nil
}

func (m *Migration) parseMigrationSQL(filePath string, isUp bool) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	text := string(content)

	// Find --UP-- and --DOWN-- sections
	upMarker := "--UP--"
	downMarker := "--DOWN--"

	upIndex := strings.Index(text, upMarker)
	downIndex := strings.Index(text, downMarker)

	if upIndex == -1 || downIndex == -1 {
		return nil, fmt.Errorf("migration file must contain both --UP-- and --DOWN-- sections")
	}

	var sql string
	if isUp {
		// Extract SQL between --UP-- and --DOWN--
		sql = text[upIndex+len(upMarker) : downIndex]
	} else {
		// Extract SQL after --DOWN--
		sql = text[downIndex+len(downMarker):]
	}

	// Trim whitespace
	sql = strings.TrimSpace(sql)

	// Split by semicolon to get individual statements
	statements := strings.Split(sql, ";")

	// Trim each statement and filter out empty ones
	var result []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		// Skip empty statements and comment-only lines
		if stmt != "" && !strings.HasPrefix(stmt, "--") {
			result = append(result, stmt)
		}
	}

	return result, nil
}
