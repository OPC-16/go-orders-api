package application

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type App struct {
    router http.Handler
    rdb *redis.Client
}

//constructor for App, call loadRoutes() for router field
//which creates new chi router, uses Logger middleware for logging
//and adds basic http handler for GET method for "/" path as http.StatusOK
func New() *App {
    app := &App{
        router: loadRoutes(),
        rdb: redis.NewClient(&redis.Options{}),
    }
    return app
}

func (a *App) Start(ctx context.Context) error {
    server := &http.Server {
        Addr: ":3000",
        Handler: a.router,
    }

    //checking if redis is connected, if not we don't even start the server
    err := a.rdb.Ping(ctx).Err()
    if err != nil {
        return fmt.Errorf("failed to connect to redis: %w", err)
    }

    fmt.Println("Starting the server...")
    err = server.ListenAndServe()
    if err != nil {
        return fmt.Errorf("failed to start the server: %w", err)
    }

    return nil
}
