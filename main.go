package main

import (
	"context"
	"log"

	"github.com/OPC-16/go-orders-api/application"
)

func main() {
    app := application.New()
    err := app.Start(context.TODO())
    if err != nil {
        log.Fatal("failed to start the app:", err)
    }
} 
