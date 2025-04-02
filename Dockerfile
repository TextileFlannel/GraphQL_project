FROM golang:1.23.7-alpine AS builder

WORKDIR /app

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /gql-proj ./cmd/server

FROM alpine:latest

RUN apk add --no-cache bash postgresql-client

COPY --from=builder /gql-proj /gql-proj
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY ./migrations /migrations
COPY .env .

EXPOSE 8080

CMD ["sh", "-c", "goose -dir /migrations postgres \"$DATABASE_URL\" up && /gql-proj"]