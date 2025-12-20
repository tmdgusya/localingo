.PHONY: sqlc-generate
sqlc-generate:
	cd db && sqlc generate

.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose down -v

.PHONY: docker-logs
docker-logs:
	docker compose logs -f postgres

.PHONY: test
test:
	go test ./... -v

.PHONY: test-short
test-short:
	go test ./... -v -short

.PHONY: build
build:
	go build -o bin/localingo ./main.go

.PHONY: run
run:
	go run main.go

.PHONY: clean
clean:
	rm -rf bin/
	docker compose down -v

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  sqlc-generate  - Generate Go code from SQL queries"
	@echo "  docker-up      - Start PostgreSQL with docker-compose"
	@echo "  docker-down    - Stop and remove PostgreSQL containers"
	@echo "  docker-logs    - Show PostgreSQL logs"
	@echo "  test           - Run all tests"
	@echo "  test-short     - Run tests excluding long-running ones"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  clean          - Clean build artifacts and Docker containers"
