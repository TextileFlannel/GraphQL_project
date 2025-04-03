package main

import (
	"flag"
	"fmt"
	"os"
	"context"
	"time"
	"graphql_project/internal/config"
	"graphql_project/migrations"
	"graphql_project/internal/graph"
	"graphql_project/internal/service"
	"graphql_project/internal/storage"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	// Загрузка конфигурации
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Парсинг флагов
	var storageType string
	flag.StringVar(&storageType, "storage", cfg.StorageType, "storage type (inmem|postgres)")
	flag.Parse()
	
	// Инициализация хранилища
	var store storage.Storage
	switch storageType {
	case "inmem":
		store = storage.NewInMemStorage()
		log.Println("Using in-memory storage")

	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

		store, err = storage.NewPostgresStorage(dsn)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		log.Println("Connected to PostgreSQL")

		if err := migrations.RunMigrations(dsn); err != nil {
			log.Fatalf("Migrations failed: %v", err)
		}

	default:
		log.Fatalf("Unknown storage type: %s", storageType)
	}

	// Инициализация сервиса
	svc := service.NewService(store)

	//Создание GraphQL резольвера
	resolver := graph.NewResolver(svc)

	// Настройки GraphQL сервера
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(&transport.Websocket{})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
	}

	serverErrors := make(chan error, 1)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Server starting on http://localhost:%s", cfg.HTTPPort)
		serverErrors <- server.ListenAndServe()
	}()

	// Основной цикл обработки событий
	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)

	case <-shutdown:
		log.Println("Starting graceful shutdown")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			if err := server.Close(); err != nil {
				log.Fatalf("Force shutdown error: %v", err)
			}
		}

		// Закрытие соединения с хранилищем данных
		if closer, ok := store.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				log.Printf("Storage close error: %v", err)
			}
		}

		log.Println("Server stopped")
	}
}
