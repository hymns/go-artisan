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
	DB *sql.DB
}

func New(db *sql.DB) *Seeder {
	return &Seeder{DB: db}
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

		// Execute each SQL statement
		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			if _, err := s.DB.Exec(stmt); err != nil {
				return fmt.Errorf("failed to run seeder %s: %w", name, err)
			}
		}

		color.Green("✓ Seeded: %s", name)
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

	// Remove comment lines starting with --
	lines := strings.Split(text, "\n")
	var sqlLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comment lines
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		sqlLines = append(sqlLines, line)
	}

	sql := strings.Join(sqlLines, "\n")
	sql = strings.TrimSpace(sql)

	// Split by semicolon to get individual statements
	statements := strings.Split(sql, ";")

	// Trim each statement
	var result []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result, nil
}
