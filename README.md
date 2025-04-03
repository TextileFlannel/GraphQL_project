# GraphQL_project

Структура проекта:

```
.
├── cmd/
│   └── server/
│       └── server.go
│
├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── graph/
│   │   ├── model/
│   │   │   ├── models_gen.go
│   │   │   └── models.go
│   │   │
│   │   ├── generated.go
│   │   ├── resolver.go   
│   │   ├── schema.graphqls
│   │   └── schema.resolvers.go
│   │
│   ├── service/
│   │   ├── service_test.go
│   │   └── service.go
│   │
│   └── storage/
│       ├── innem_test.go
│       ├── innem.go
│       ├── postgres_test.go
│       ├── postgres.go
│       └── storage.go
│
├── migrations/
│   ├── 20250402203731_tables.sql
│   └── migrations.go
│
├── .env
├── docker-compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── gqlgen.yml
├── Makefile
└── README.md
```

# Клонируем проект

```
git clone https://github.com/TextileFlannel/GraphQL_project.git
```

# Запуск:

## Через Docker:

```
make docker-compose-up
```

## Остановка:

```
make docker-compose-down
```

# Локально:

# Сборка проекта:

```
make build
```

## In Memory:

```
make run
```

##  Postgres:

```
make run-postgres
```

## Применение миграций:

```
make migrate-up
```

## Запуск тестов:

```
make tests
```

# Примеры запросов:

## Через curl

Создание поста:
```
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "mutation { createPost(input: {title: \"1\", content: \"2\", commentable: true, author: \"danil\"}) { id } }"}' \
  http://localhost:8080/query
```
Пример ответа:
```
{"data":{"createPost":{"id":"7a482ad0-10ff-4204-80c1-58c02b05e64d"}}}
```

Добавление комментария под постом:
```
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "mutation { createComment(input: {content: \"wqe\", author: \"asdd\", postId: \"7a482ad0-10ff-4204-80c1-58c02b05e64d\"}) { id } }"}' \
  http://localhost:8080/query
```
Пример ответа:
```
{"data":{"createComment":{"id":"a3df2b58-aa05-47e6-b212-e91b8e87d00e"}}}
```

Вывод постов и комментариев с пагинацией
```
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "{posts (limit: 1) { id comments { id comments (limit: 1, offset: 0) { id } } } }"}' \
  http://localhost:8080/query
```
Пример ответа:
```
{"data":{"posts":[{"id":"7a482ad0-10ff-4204-80c1-58c02b05e64d","comments":[{"id":"a3df2b58-aa05-47e6-b212-e91b8e87d00e","comments":null}]}]}}
```

## Через GraphQL playground

Создание поста:
```
mutation {
  createPost(input: {title:"1", content:"2", commentable: true, author:"danil"}) {
		id
  }
}
```
Пример ответа:
```
{
  "data": {
    "createPost": {
      "id": "684f5bfd-56d8-4c28-b232-c5a6997bb8c1"
    }
  }
}
```

Добавление комментария под постом:
```
mutation {
  createComment(input: {content: "wqe", author: "asdd", postId: "684f5bfd-56d8-4c28-b232-c5a6997bb8c1"}) {
		id
  }
}
```
Пример ответа:
```
{
  "data": {
    "createComment": {
      "id": "86bc5828-efcb-4f2a-a71e-9a58d1755bb9"
    }
  }
}
```

Вывод постов и комментариев с пагинацией
```
{
  posts (limit: 1) {
    id
    comments {
			id
      comments (limit: 1, offset: 0) {
        id
      }
    }
  }
}
```
Пример ответа:
```
{
  "data": {
    "posts": [
      {
        "id": "684f5bfd-56d8-4c28-b232-c5a6997bb8c1",
        "comments": [
          {
            "id": "86bc5828-efcb-4f2a-a71e-9a58d1755bb9",
            "comments": null
          }
        ]
      }
    ]
  }
}
```

Подписка на посты:
```
subscription {
  commentAdded(postID: "684f5bfd-56d8-4c28-b232-c5a6997bb8c1") {
    id
  }
}
```
Пример ответа:
```
{
  "data": {
    "commentAdded": {
      "id": "4ae6bdb7-9bf9-44ec-a4cd-c2fea6db77be"
    }
  }
}
```