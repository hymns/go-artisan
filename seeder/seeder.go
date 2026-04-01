package seeder

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
)

type Seeder struct {
	DB     *sql.DB
	Driver string
}

func New(db *sql.DB) *Seeder {
	return &Seeder{
		DB:     db,
		Driver: getDBDriver(db),
	}
}

func getDBDriver(db *sql.DB) string {
	var driver string
	if err := db.QueryRow("SELECT 1").Scan(&driver); err == nil {
		return "mysql"
	}
	// Try PostgreSQL specific query
	if err := db.QueryRow("SELECT version()").Scan(&driver); err == nil {
		if strings.Contains(strings.ToLower(driver), "postgres") {
			return "postgres"
		}
	}
	return "mysql" // default
}

func (s *Seeder) EnsureSeedersTable() error {
	var query string

	switch s.Driver {
	case "postgres":
		query = `CREATE TABLE IF NOT EXISTS seeders (
			id SERIAL PRIMARY KEY,
			seeder VARCHAR(255) NOT NULL UNIQUE,
			seeded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	case "sqlite", "sqlite3":
		query = `CREATE TABLE IF NOT EXISTS seeders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			seeder VARCHAR(255) NOT NULL UNIQUE,
			seeded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	case "sqlserver", "mssql":
		query = `IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='seeders' AND xtype='U')
			CREATE TABLE seeders (
				id INT IDENTITY(1,1) PRIMARY KEY,
				seeder VARCHAR(255) NOT NULL UNIQUE,
				seeded_at DATETIME DEFAULT GETDATE()
			)`
	default: // mysql
		query = `CREATE TABLE IF NOT EXISTS seeders (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			seeder VARCHAR(255) NOT NULL UNIQUE,
			seeded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	}

	_, err := s.DB.Exec(query)
	return err
}

func (s *Seeder) getSeeded() ([]string, error) {
	rows, err := s.DB.Query("SELECT seeder FROM seeders ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seeded []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		seeded = append(seeded, name)
	}

	return seeded, rows.Err()
}

func (s *Seeder) recordSeeder(name string) error {
	query := "INSERT INTO seeders (seeder) VALUES (?)"
	if s.Driver == "postgres" {
		query = "INSERT INTO seeders (seeder) VALUES ($1)"
	} else if s.Driver == "sqlserver" || s.Driver == "mssql" {
		query = "INSERT INTO seeders (seeder) VALUES (@p1)"
	}

	_, err := s.DB.Exec(query, name)
	return err
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *Seeder) RunFile(filePath string) error {
	name := filepath.Base(filePath)

	// Read and parse SQL file
	statements, err := s.parseSeederSQL(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse seeder %s: %w", name, err)
	}

	// Start transaction for atomic seeding
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for seeder %s: %w", name, err)
	}

	// Execute each SQL statement within transaction
	for _, stmt := range statements {
		if stmt == "" {
			continue
		}
		if _, err := tx.Exec(stmt); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to run seeder %s: %w", name, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit seeder %s: %w", name, err)
	}

	color.Green("✓ Seeded: %s", name)
	return nil
}

func (s *Seeder) AutoSeed(seedersPath string) error {
	if err := s.EnsureSeedersTable(); err != nil {
		return fmt.Errorf("failed to ensure seeders table: %w", err)
	}

	seeded, err := s.getSeeded()
	if err != nil {
		return fmt.Errorf("failed to get seeded list: %w", err)
	}

	files, err := s.getSeederFiles(seedersPath)
	if err != nil {
		return fmt.Errorf("failed to get seeder files: %w", err)
	}

	executed := 0
	for _, file := range files {
		name := filepath.Base(file)

		// Skip if already seeded
		if contains(seeded, name) {
			continue
		}

		// Read and parse SQL file
		statements, err := s.parseSeederSQL(file)
		if err != nil {
			return fmt.Errorf("failed to parse seeder %s: %w", name, err)
		}

		// Start transaction for atomic seeding
		tx, err := s.DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for seeder %s: %w", name, err)
		}

		// Execute each SQL statement within transaction
		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to run seeder %s: %w", name, err)
			}
		}

		// Record seeder within same transaction
		var recordQuery string
		if s.Driver == "postgres" {
			recordQuery = "INSERT INTO seeders (seeder) VALUES ($1)"
		} else if s.Driver == "sqlserver" || s.Driver == "mssql" {
			recordQuery = "INSERT INTO seeders (seeder) VALUES (@p1)"
		} else {
			recordQuery = "INSERT INTO seeders (seeder) VALUES (?)"
		}

		if _, err := tx.Exec(recordQuery, name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record seeder %s: %w", name, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit seeder %s: %w", name, err)
		}

		executed++
	}

	if executed == 0 {
		// Silent - no output for AutoSeed
	}

	return nil
}

func (s *Seeder) Run(seedersPath string) error {
	files, err := s.getSeederFiles(seedersPath)
	if err != nil {
		return fmt.Errorf("failed to get seeder files: %w", err)
	}

	for _, file := range files {
		name := filepath.Base(file)

		// Read and parse SQL file
		statements, err := s.parseSeederSQL(file)
		if err != nil {
			return fmt.Errorf("failed to parse seeder %s: %w", name, err)
		}

		// Start transaction for atomic seeding
		tx, err := s.DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for seeder %s: %w", name, err)
		}

		// Execute each SQL statement within transaction
		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to run seeder %s: %w", name, err)
			}
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit seeder %s: %w", name, err)
		}

		color.Green("✓ Seeded: %s", name)
	}

	return nil
}

func (s *Seeder) RunWithTracking(seedersPath string) error {
	if err := s.EnsureSeedersTable(); err != nil {
		return fmt.Errorf("failed to ensure seeders table: %w", err)
	}

	seeded, err := s.getSeeded()
	if err != nil {
		return fmt.Errorf("failed to get seeded list: %w", err)
	}

	files, err := s.getSeederFiles(seedersPath)
	if err != nil {
		return fmt.Errorf("failed to get seeder files: %w", err)
	}

	executed := 0
	for _, file := range files {
		name := filepath.Base(file)

		// Skip if already seeded (silent)
		if contains(seeded, name) {
			continue
		}

		// Read and parse SQL file
		statements, err := s.parseSeederSQL(file)
		if err != nil {
			return fmt.Errorf("failed to parse seeder %s: %w", name, err)
		}

		// Start transaction for atomic seeding
		tx, err := s.DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for seeder %s: %w", name, err)
		}

		// Execute each SQL statement within transaction
		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to run seeder %s: %w", name, err)
			}
		}

		// Record seeder within same transaction
		var recordQuery string
		if s.Driver == "postgres" {
			recordQuery = "INSERT INTO seeders (seeder) VALUES ($1)"
		} else if s.Driver == "sqlserver" || s.Driver == "mssql" {
			recordQuery = "INSERT INTO seeders (seeder) VALUES (@p1)"
		} else {
			recordQuery = "INSERT INTO seeders (seeder) VALUES (?)"
		}

		if _, err := tx.Exec(recordQuery, name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record seeder %s: %w", name, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit seeder %s: %w", name, err)
		}

		color.Green("✓ Seeded: %s", name)
		executed++
	}

	if executed == 0 {
		color.Cyan("Nothing to seed.")
	}

	return nil
}

func (s *Seeder) getSeederFiles(path string) ([]string, error) {
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
		// Skip hidden files and .go files (old format)
		if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".go") {
			continue
		}

		files = append(files, filepath.Join(path, name))
	}

	sort.Strings(files)
	return files, nil
}

type SeederStatus struct {
	Name   string
	Seeded bool
}

func (s *Seeder) Status(seedersPath string) ([]SeederStatus, error) {
	if err := s.EnsureSeedersTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure seeders table: %w", err)
	}

	seeded, err := s.getSeeded()
	if err != nil {
		return nil, fmt.Errorf("failed to get seeded list: %w", err)
	}

	files, err := s.getSeederFiles(seedersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get seeder files: %w", err)
	}

	var statuses []SeederStatus
	for _, file := range files {
		name := filepath.Base(file)
		status := SeederStatus{
			Name:   name,
			Seeded: contains(seeded, name),
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (s *Seeder) MakeSeeder(seederName, seedersPath string) error {
	filename := seederName
	filepath := filepath.Join(seedersPath, filename)

	template := s.getSeederTemplate(seederName)

	if err := os.MkdirAll(seedersPath, 0755); err != nil {
		return fmt.Errorf("failed to create seeders directory: %w", err)
	}

	if err := os.WriteFile(filepath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write seeder file: %w", err)
	}

	color.Green("✓ Seeder created: %s", filename)
	return nil
}

func (s *Seeder) getSeederTemplate(seederName string) string {
	return fmt.Sprintf(`-- Seeder: %s

-- Add your INSERT statements here
-- Example:
-- INSERT INTO users (name, email, password) VALUES 
--   ('John Doe', 'john@example.com', 'hashed_password'),
--   ('Jane Smith', 'jane@example.com', 'hashed_password');

`, seederName)
}

func (s *Seeder) parseSeederSQL(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	text := string(content)

	// Remove comment lines starting with -- and stop at --DOWN-- marker
	lines := strings.Split(text, "\n")
	var sqlLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Stop processing at --DOWN-- marker (seeders don't need down migrations)
		if trimmed == "--DOWN--" {
			break
		}

		// Skip empty lines and comment lines (but not --DOWN-- since we break above)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		sqlLines = append(sqlLines, line)
	}

	sql := strings.Join(sqlLines, "\n")
	sql = strings.TrimSpace(sql)

	// Smart split by semicolon (respect quoted strings)
	statements := s.splitSQLStatements(sql)

	// Normalize each statement (escape newlines in quoted strings)
	var result []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			// Normalize newlines and quotes within string literals
			stmt = s.normalizeSQL(stmt)
			result = append(result, stmt)
		}
	}

	return result, nil
}

// splitSQLStatements splits SQL by semicolon but respects quoted strings
func (s *Seeder) splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for i := 0; i < len(sql); i++ {
		char := sql[i]

		// Handle escape sequences
		if escaped {
			current.WriteByte(char)
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			current.WriteByte(char)
			continue
		}

		// Toggle quote states
		if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			current.WriteByte(char)
			continue
		}

		if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			current.WriteByte(char)
			continue
		}

		// Split on semicolon only if not inside quotes
		if char == ';' && !inSingleQuote && !inDoubleQuote {
			stmt := current.String()
			if strings.TrimSpace(stmt) != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
			continue
		}

		current.WriteByte(char)
	}

	// Add remaining content
	if current.Len() > 0 {
		stmt := current.String()
		if strings.TrimSpace(stmt) != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// normalizeSQL escapes newlines and special characters within quoted strings
func (s *Seeder) normalizeSQL(sql string) string {
	var result strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for i := 0; i < len(sql); i++ {
		char := sql[i]

		// Handle escape sequences
		if escaped {
			result.WriteByte(char)
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			result.WriteByte(char)
			continue
		}

		// Toggle quote states
		if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			result.WriteByte(char)
			continue
		}

		if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			result.WriteByte(char)
			continue
		}

		// Escape newlines within quoted strings
		if (inSingleQuote || inDoubleQuote) && char == '\n' {
			result.WriteString("\\n")
			continue
		}

		// Escape carriage returns within quoted strings
		if (inSingleQuote || inDoubleQuote) && char == '\r' {
			result.WriteString("\\r")
			continue
		}

		// Escape tabs within quoted strings
		if (inSingleQuote || inDoubleQuote) && char == '\t' {
			result.WriteString("\\t")
			continue
		}

		result.WriteByte(char)
	}

	return result.String()
}
