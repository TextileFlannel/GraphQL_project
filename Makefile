APP_NAME=gql-proj

## Сборка проекта
build:
	go build -o $(APP_NAME) ./cmd/server

## Запуск проекта
run: build
	./$(APP_NAME) --storage=inmem

## Запуск проекта с PostgreSQL
run-postgres: go run
	./$(APP_NAME) --storage=postgres

## Применение миграций
migrate-up:
	goose -dir ./migrations postgres "postgres://postgres:postgres@localhost:5432/links?sslmode=disable up"

## Запуск через Docker Compose
docker-compose-up:
	docker-compose up -d --build

## Остановка Docker Compose
docker-compose-down:
	docker-compose down

## Запуск тестов
tests:
	go test ./...