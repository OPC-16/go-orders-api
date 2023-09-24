package application

import (
	"context"
	"fmt"
	"net/http"
)

type App struct {
    router http.Handler
}

//constructor for App, call loadRoutes() for router field
//which creates new chi router, uses Logger middleware for logging
//and adds basic http handler for GET method for "/" path as http.StatusOK
func New() *App {
    app := &App{
        router: loadRoutes(),
    }
    return app
}

func (a *App) Start(ctx context.Context) error {
    server := &http.Server {
        Addr: ":3000",
        Handler: a.router,
    }

    err := server.ListenAndServe()
    if err != nil {
        return fmt.Errorf("failed to start the server: %w", err)
    }

    return nil
}
