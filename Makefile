.PHONY: build install clean test deps migrate rollback seed migration

build:
	@echo "Building artisan..."
	@cd cmd/artisan && go build -o ../../bin/artisan
	@echo "✓ Build complete: bin/artisan"

install: build
	@echo "Installing artisan..."
	@cp bin/artisan /usr/local/bin/artisan
	@echo "✓ Installed to /usr/local/bin/artisan"

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies installed"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "✓ Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./...

migrate: build
	@./bin/artisan db:migrate

rollback: build
	@./bin/artisan db:rollback

seed: build
	@./bin/artisan db:seed

migration:
	@if [ -z "$(table)" ] || [ -z "$(name)" ]; then \
		echo "Usage: make migration table=<table_name> name=<migration_name>"; \
		exit 1; \
	fi
	@./bin/artisan make:migration $(table) $(name)
	@make build

help:
	@echo "Artisan - Database Migration Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  make build                              - Build the artisan binary"
	@echo "  make install                            - Install artisan to /usr/local/bin"
	@echo "  make deps                               - Install Go dependencies"
	@echo "  make clean                              - Remove build artifacts"
	@echo "  make test                               - Run tests"
	@echo ""
	@echo "Migration shortcuts (auto-build):"
	@echo "  make migrate                            - Run migrations"
	@echo "  make rollback                           - Rollback last batch"
	@echo "  make seed                               - Run seeders"
	@echo "  make migration table=<name> name=<name> - Create new migration"
	@echo ""
	@echo "  make help                               - Show this help message"
