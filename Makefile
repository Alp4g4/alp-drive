MIGRATIONS := internal/infrastructure/sqlite/migrations
GOOSE := go tool goose -dir $(MIGRATIONS)

export GOOSE_DRIVER := sqlite3
export GOOSE_DBSTRING := file:data/alp-drive.db?_pragma=foreign_keys(1)&_time_format=sqlite

.PHONY: migrate-create migrate-up migrate-down migrate-status migrate-check
.PHONY: templ-generate css-build assets build run check

migrate-create: ## make migrate-create name=create_folder
	@test -n "$(name)" || (echo "usage: make migrate-create name=<verb>_<object>" && exit 1)
	$(GOOSE) create $(name) sql

migrate-up: ## apply every pending migration
	mkdir -p data
	$(GOOSE) up

migrate-down: ## roll back the last migration
	$(GOOSE) down

migrate-status: ## list applied and pending migrations
	$(GOOSE) status

migrate-check: ## up, down to 0, up: what CI runs on a fresh database
	$(GOOSE) up
	$(GOOSE) reset
	$(GOOSE) up

templ-generate: ## regenerate the *_templ.go files from .templ sources
	go tool templ generate

css-build: ## rebuild internal/delivery/web/static/app.css from the Tailwind theme
	npx @tailwindcss/cli -i assets/css/app.css -o internal/delivery/web/static/app.css --minify

assets: templ-generate css-build

build: assets
	go build -o bin/alp-drive ./cmd/alp-drive

run: migrate-up
	go run ./cmd/alp-drive

check:
	gofmt -l .
	go vet ./...
	bash scripts/check-deps.sh
	go test ./...
