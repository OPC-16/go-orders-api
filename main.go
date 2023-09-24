package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
    router := chi.NewRouter()
    router.Use(middleware.Logger)

    router.Get("/hello", basicHandler)

    server := &http.Server {
        Addr: ":3000",
        Handler: router,
    }

    err := server.ListenAndServe()
    if err != nil {
        log.Fatal("Error:", err)
    }
}

func basicHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello from orders-api\n"))
}
