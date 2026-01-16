package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hymns/go-artisan/migration"
	"github.com/hymns/go-artisan/seeder"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/testdb?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	m := migration.New(db)

	// Option 1: AutoMigrate - Silent migration on app startup (recommended for production)
	if err := m.AutoMigrate("./database/migrations"); err != nil {
		log.Fatal(err)
	}

	// Option 2: Migrate - With colored output (good for CLI tools)
	// if err := m.Migrate("./database/migrations"); err != nil {
	// 	log.Fatal(err)
	// }

	s := seeder.New(db)

	if err := s.Run("./database/seeders"); err != nil {
		log.Fatal(err)
	}
}
