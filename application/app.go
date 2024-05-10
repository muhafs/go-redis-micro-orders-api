package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	return &App{
		router: loadRoutes(),
		rdb:    redis.NewClient(&redis.Options{}),
	}
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: app.router,
	}

	if err := app.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	defer func() {
		if err := app.rdb.Close(); err != nil {
			fmt.Println("failed to close redis:", err)
		}
	}()

	fmt.Println("Server is running at localhost:3000")

	ch := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return server.Shutdown(timeout)
	}
}
