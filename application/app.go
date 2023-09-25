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
    rdb *redis.Client
}

//constructor for App, call App.loadRoutes()
//which creates new chi router, uses Logger middleware for logging
//and adds basic http handler for GET method for "/" path as http.StatusOK
func New() *App {
    app := &App{
        rdb: redis.NewClient(&redis.Options{}),
    }

    app.loadRoutes()

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

    defer func() {
        if err := a.rdb.Close(); err != nil {
            fmt.Println("Failed to close redis:", err)
        }
    }()

    fmt.Println("Starting the server...")

    ch := make(chan error, 1)
    go func() {
        err = server.ListenAndServe()
        if err != nil {
            ch <- fmt.Errorf("failed to start the server: %w", err)
        }
        close(ch)
    }()

    select {
    case err = <-ch:
        return err
    case <-ctx.Done():
        //we cannot use the same context as this context is already cancelled at this point of time, and if used by itself our shutdown could run indefinatly
        //so we add a timeout to it

        timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
        defer cancel()      //this closes our redis instance
        return server.Shutdown(timeout)
    }

    return nil
}
