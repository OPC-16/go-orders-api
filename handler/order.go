package handler

import (
	"math/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/OPC-16/go-orders-api/model"
	"github.com/OPC-16/go-orders-api/repository/order"
	"github.com/google/uuid"
)

type Order struct {
    Repo *order.RedisRepo
}

/* These are handler methods which get called on specific routes, these routes are defined in application/routes.go */

func (h *Order) Create(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Create an order")
    //body has an anonymous type associated with it and it represents our expected POST data we receive from the Client
    var body struct {
        CustomerID uuid.UUID        `json:"customer_id"`
        LineItems  []model.LineItem `json:"line_items"`
    }

    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    now := time.Now().UTC()

    order := model.Order {
        OrderID: rand.Uint64(),
        CustomerID: body.CustomerID,
        LineItems: body.LineItems,
        CreatedAt: &now,
    }

    err := h.Repo.Insert(r.Context(), order)
    if err != nil {
        fmt.Println("failed to insert:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    //returning our generated model.Order to the client in json format
    res, err := json.Marshal(order)
    if err != nil {
        fmt.Println("failed to marshal:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.Write(res)
    w.WriteHeader(http.StatusCreated)
}

func (h *Order) List(w http.ResponseWriter, r *http.Request) {
    fmt.Println("List all orders")
}

func (h *Order) GetByID(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Get an order by ID")
}

func (h *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Update an order by ID")
}

func (h *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Delete an order by ID")
}
