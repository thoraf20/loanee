.PHONY: run build migrate tidy

run:
	@./start.sh

build:
	@go build -o loanee ./cmd/main.go

migrate:
	@go run cmd/migrate.go

tidy:
	@go mod tidy && go mod verify

migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down
