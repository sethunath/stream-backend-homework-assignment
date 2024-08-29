package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/GetStream/stream-backend-homework-assignment/postgres"
	"github.com/GetStream/stream-backend-homework-assignment/redis"
)

const connStr = "postgres://message-api:message-api@localhost:5432/message-api?sslmode=disable"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		stop()
	}()

	addr := flag.String("addr", "localhost:8080", "HTTP network address")
	connStr := flag.String("connection-string", connStr, "Postgres connection string")
	redisAddr := flag.String("redis-address", "localhost:6379", "Redis endpoint")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	pg, err := postgres.Connect(ctx, *connStr)
	if err != nil {
		logger.Error("Could not connect to PostgreSQL", "error", err.Error())
		os.Exit(1)
	}

	redis, err := redis.Connect(ctx, *redisAddr)
	if err != nil {
		logger.Error("Could not connect to Redis", "error", err.Error())
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		logger.Error("Could not listen", "error", err)
		os.Exit(1)
	}

	api := &api.API{
		Logger: logger,
		DB:     pg,
		Cache:  redis,
	}

	srv := &http.Server{
		Handler: api,
	}

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	logger.Info("Ready to accept traffic", "address", *addr)
	if err := srv.Serve(lis); err != nil && err != http.ErrServerClosed {
		logger.Error("Could not start server", "error", err)
		os.Exit(1)
	}
}
