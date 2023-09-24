package main

import (
	"context"
	"log"
    "os"
    "os/signal"

	"github.com/OPC-16/go-orders-api/application"
)

func main() {
    app := application.New()

    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    err := app.Start(ctx)
    if err != nil {
        log.Fatal("failed to start the app:", err)
    }
}
