package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hymns/go-artisan/migration"
	"github.com/hymns/go-artisan/seeder"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	loadEnvFile()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	db, err := connectDB()
	if err != nil {
		color.Red("✗ Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	switch command {
	case "migrate", "db:migrate":
		handleMigrate(db, args)
	case "migrate:rollback", "db:rollback":
		handleMigrateRollback(db, args)
	case "migrate:fresh":
		handleMigrateFresh(db, args)
	case "migrate:status":
		handleMigrateStatus(db)
	case "migrate:dry-run", "migrate:dryrun":
		handleMigrateDryRun(db)
	case "db:seed":
		handleSeed(db, args)
	case "make:migration":
		handleMakeMigration(args)
	case "make:seeder":
		handleMakeSeeder(args)
	case "about":
		printAbout()
	case "help", "--help", "-h":
		printUsage()
	default:
		color.Red("✗ Unknown command: %s", command)
		printUsage()
		os.Exit(1)
	}
}

func loadEnvFile() {
	cwd, err := os.Getwd()
	if err != nil {
		color.Yellow("Warning: Could not get current directory")
		return
	}

	envPath := filepath.Join(cwd, ".env")
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			color.Yellow("Warning: Failed to load .env file")
		}
	} else {
		color.Yellow("Warning: .env file not found in current directory")
	}
}

func connectDB() (*sql.DB, error) {
	dbDriver := getEnv("DB_DRIVER", "mysql")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "database")
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "")

	var dsn string
	switch dbDriver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	case "sqlserver", "mssql":
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", dbUser, dbPass, dbHost, dbPort, dbName)
	case "sqlite", "sqlite3":
		// For SQLite, dbName is the file path
		dsn = dbName
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", dbDriver)
	}

	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func handleMigrate(db *sql.DB, args []string) {
	m := migration.New(db)
	migrationsPath := getEnv("MIGRATIONS_PATH", "./database/migrations")

	// Parse flags
	var specificPath string
	runSeed := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "--path=") {
			specificPath = strings.TrimPrefix(arg, "--path=")
		} else if arg == "--seed" {
			runSeed = true
		}
	}

	// Run specific migration file if --path provided
	if specificPath != "" {
		if err := m.MigrateFile(specificPath); err != nil {
			color.Red("✗ Migration failed: %v", err)
			os.Exit(1)
		}
	} else {
		// Run all pending migrations
		if err := m.Migrate(migrationsPath); err != nil {
			color.Red("✗ Migration failed: %v", err)
			os.Exit(1)
		}
	}

	if runSeed {
		fmt.Println()
		color.Cyan("Running seeders...")
		handleSeed(db, []string{})
	}
}

func handleMigrateRollback(db *sql.DB, args []string) {
	m := migration.New(db)
	migrationsPath := getEnv("MIGRATIONS_PATH", "./database/migrations")

	// Parse --step flag, default to 1
	steps := 1
	for _, arg := range args {
		if strings.HasPrefix(arg, "--step=") {
			stepStr := strings.TrimPrefix(arg, "--step=")
			if s, err := fmt.Sscanf(stepStr, "%d", &steps); err != nil || s != 1 {
				color.Red("✗ Invalid --step value: %s", stepStr)
				os.Exit(1)
			}
		}
	}

	// Rollback N steps
	for i := 0; i < steps; i++ {
		if err := m.Rollback(migrationsPath); err != nil {
			color.Red("✗ Rollback failed: %v", err)
			os.Exit(1)
		}
	}
}

func handleMigrateFresh(db *sql.DB, args []string) {
	m := migration.New(db)
	migrationsPath := getEnv("MIGRATIONS_PATH", "./database/migrations")

	// Parse --seed flag
	runSeed := false
	for _, arg := range args {
		if arg == "--seed" {
			runSeed = true
			break
		}
	}

	color.Cyan("Rolling back all migrations...")

	// Rollback all migrations
	for {
		batch, err := m.GetLastBatch()
		if err != nil {
			color.Red("✗ Failed to get last batch: %v", err)
			os.Exit(1)
		}

		if batch == 0 {
			break
		}

		if err := m.Rollback(migrationsPath); err != nil {
			color.Red("✗ Rollback failed: %v", err)
			os.Exit(1)
		}
	}

	color.Green("✓ All migrations rolled back")

	// Re-run all migrations
	fmt.Println()
	color.Cyan("Running migrations...")
	if err := m.Migrate(migrationsPath); err != nil {
		color.Red("✗ Migration failed: %v", err)
		os.Exit(1)
	}

	// Run seeders if --seed flag provided
	if runSeed {
		fmt.Println()
		color.Cyan("Running seeders...")
		handleSeed(db, []string{})
	}
}

func handleMigrateStatus(db *sql.DB) {
	m := migration.New(db)
	migrationsPath := getEnv("MIGRATIONS_PATH", "./database/migrations")

	statuses, err := m.Status(migrationsPath)
	if err != nil {
		color.Red("✗ Failed to get migration status: %v", err)
		os.Exit(1)
	}

	if len(statuses) == 0 {
		color.Cyan("No migrations found.")
		return
	}

	color.Cyan("\nMigration Status:\n")
	color.White("%-50s %-10s %s\n", "Migration", "Batch", "Ran")
	color.White("%s\n", strings.Repeat("-", 70))

	for _, status := range statuses {
		if status.Migrated {
			fmt.Printf("%-50s %-10d ", status.Name, status.Batch)
			color.Green("YES\n")
		} else {
			fmt.Printf("%-50s %-10s ", status.Name, "-")
			color.Yellow("NO\n")
		}
	}
}

func handleMigrateDryRun(db *sql.DB) {
	m := migration.New(db)
	migrationsPath := getEnv("MIGRATIONS_PATH", "./database/migrations")

	if err := m.DryRun(migrationsPath); err != nil {
		color.Red("✗ Dry run failed: %v", err)
		os.Exit(1)
	}
}

func handleSeed(db *sql.DB, args []string) {
	s := seeder.New(db)
	seedersPath := getEnv("SEEDERS_PATH", "./database/seeders")

	// Parse --path flag
	var specificPath string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--path=") {
			specificPath = strings.TrimPrefix(arg, "--path=")
			break
		}
	}

	// Run specific seeder file if --path provided
	if specificPath != "" {
		if err := s.RunFile(specificPath); err != nil {
			color.Red("✗ Seeding failed: %v", err)
			os.Exit(1)
		}
	} else {
		// Run all seeders
		if err := s.Run(seedersPath); err != nil {
			color.Red("✗ Seeding failed: %v", err)
			os.Exit(1)
		}
	}
}

func handleMakeMigration(args []string) {
	var tableName, migrationName string

	// Parse arguments and flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "--table=") {
			tableName = strings.TrimPrefix(arg, "--table=")
		} else if i == 0 && !strings.HasPrefix(arg, "--") {
			tableName = arg
		} else if i == 1 && !strings.HasPrefix(arg, "--") {
			migrationName = arg
		}
	}

	if tableName == "" {
		color.Red("✗ Usage: artisan make:migration <table_name> [migration_name]")
		color.Red("✗    or: artisan make:migration --table=<table_name> [migration_name]")
		os.Exit(1)
	}

	// Auto-generate migration name if not provided
	if migrationName == "" {
		migrationName = fmt.Sprintf("create_%s_table", tableName)
	}

	migrationsPath := getEnv("MIGRATIONS_PATH", "./database/migrations")

	m := migration.New(nil)
	if err := m.MakeMigration(tableName, migrationName, migrationsPath); err != nil {
		color.Red("✗ Failed to create migration: %v", err)
		os.Exit(1)
	}
}

func handleMakeSeeder(args []string) {
	var seederName string

	// Parse arguments and flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "--seeder=") {
			seederName = strings.TrimPrefix(arg, "--seeder=")
		} else if i == 0 && !strings.HasPrefix(arg, "--") {
			seederName = arg
		}
	}

	if seederName == "" {
		color.Red("✗ Usage: artisan make:seeder <seeder_name>")
		color.Red("✗    or: artisan make:seeder --seeder=<seeder_name>")
		os.Exit(1)
	}

	// Auto-append _seeder suffix if not present
	if !strings.HasSuffix(seederName, "_seeder") {
		seederName = seederName + "_seeder"
	}

	seedersPath := getEnv("SEEDERS_PATH", "./database/seeders")

	s := seeder.New(nil)
	if err := s.MakeSeeder(seederName, seedersPath); err != nil {
		color.Red("✗ Failed to create seeder: %v", err)
		os.Exit(1)
	}
}

func printAbout() {
	color.Cyan("\n╔════════════════════════════════════════════════════════════════╗\n")
	color.Cyan("║                                                                ║\n")
	color.Cyan("║  ")
	color.White("Artisan - Database Migration Tool for Go")
	color.Cyan("                   ║\n")
	color.Cyan("║                                                                ║\n")
	color.Cyan("╚════════════════════════════════════════════════════════════════╝\n\n")

	color.Green("Version: ")
	color.White("1.3.0\n")

	color.Green("Author:  ")
	color.White("Muhammad Hamizi Jaminan\n")

	color.Green("Email:   ")
	color.Cyan("hello@hamizi.net\n")

	color.Green("GitHub:  ")
	color.Cyan("https://github.com/hymns/go-artisan\n")

	fmt.Println()
	color.Yellow("A Laravel-inspired database migration tool for Go developers.\n")
	color.White("Supports MySQL, PostgreSQL, SQL Server, and SQLite with batch tracking,\n")
	color.White("multi-statement migrations, and Laravel-style commands.\n")
	fmt.Println()

	color.Cyan("Made with ❤️  for Go developers who miss Laravel's migration system.\n\n")
}

func printUsage() {
	color.Cyan("\nArtisan - Database Migration Tool\n")
	color.White("Usage: artisan [command]\n\n")

	commands := []struct {
		name        string
		description string
	}{
		{"migrate", "Run database migrations"},
		{"migrate --path=<file>", "Run specific migration file"},
		{"migrate --seed", "Run migrations and seeders"},
		{"migrate:rollback", "Rollback migrations (default: 1 step)"},
		{"migrate:rollback --step=N", "Rollback N steps"},
		{"migrate:fresh", "Rollback all, then re-run migrations"},
		{"migrate:fresh --seed", "Rollback all, migrate, then seed"},
		{"migrate:status", "Show migration status (pending/migrated)"},
		{"migrate:dry-run", "Preview pending migrations without running"},
		{"db:seed", "Run database seeders"},
		{"db:seed --path=<file>", "Run specific seeder file"},
		{"", ""},
		{"make:migration <table>", "Create migration (auto-name: create_<table>_table)"},
		{"make:migration <table> <name>", "Create migration with custom name"},
		{"make:migration --table=<table>", "Create migration using flag"},
		{"", ""},
		{"make:seeder <name>", "Create seeder (auto-append: _seeder)"},
		{"make:seeder --seeder=<name>", "Create seeder using flag"},
		{"", ""},
		{"about", "Show information about Artisan"},
		{"help", "Show this help message"},
	}

	color.Cyan("Available commands:\n")
	for _, cmd := range commands {
		if cmd.name == "" {
			fmt.Println()
			continue
		}
		color.White("  %-40s %s\n", cmd.name, cmd.description)
	}
	fmt.Println()
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
