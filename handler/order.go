package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/OPC-16/go-orders-api/model"
	"github.com/OPC-16/go-orders-api/repository/order"
	"github.com/go-chi/chi/v5"
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
    cursorStr := r.URL.Query().Get("cursor")
    if cursorStr == "" {
        cursorStr = "0"
    }

    const decimal = 10
    const bitSize = 64
    cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    const size = 50
    res, err := h.Repo.FindAll(r.Context(), order.FindAllPage{
        Offset: cursor,
        Size: size,
    })
    if err != nil {
        fmt.Println("failed to find all:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    //crafting our response
    var response struct {
        Items []model.Order `json:"items"`
        Next uint64         `json:"next,omitempty"`
    }
    response.Items = res.Orders
    response.Next = res.Cursor

    data, err := json.Marshal(response)
    if err != nil {
        fmt.Println("failed to marshal:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.Write(data)
}

func (h *Order) GetByID(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Get an order by ID")
    idParam := chi.URLParam(r, "id")

    const base = 10
    const bitSize = 64
    orderID, err := strconv.ParseUint(idParam, base, bitSize)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    o, err := h.Repo.FindByID(r.Context(), orderID)
    if errors.Is(err, order.ErrNotExist) {
        w.WriteHeader(http.StatusNotFound)
        return
    } else if err != nil {
        fmt.Println("failed to find by id:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    //lastly if everything went right we just need to write our order model to our response
    if err := json.NewEncoder(w).Encode(o); err != nil {
        fmt.Println("failed to marshal:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (h *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Update an order by ID")
    var body struct {
        Status string `json:"status"`
    }

    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    idParam := chi.URLParam(r, "id")

    const base = 10
    const bitSize = 64

    orderID, err := strconv.ParseUint(idParam, base, bitSize)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    theOrder, err := h.Repo.FindByID(r.Context(), orderID)
    if errors.Is(err, order.ErrNotExist) {
        w.WriteHeader(http.StatusNotFound)
        return
    } else if err != nil {
        fmt.Println("failed to find by id:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    const completedStatus = "completed"
    const shippedStatus   = "shipped"
    now := time.Now().UTC()

    switch body.Status {
    case shippedStatus:
        if theOrder.ShippedAt != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        theOrder.ShippedAt = &now
    case completedStatus:
        if theOrder.CompletedAt != nil || theOrder.ShippedAt == nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        theOrder.CompletedAt = &now
    default:
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    err = h.Repo.Update(r.Context(), theOrder)
    if err != nil {
        fmt.Println("failed to insert:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    if err := json.NewEncoder(w).Encode(theOrder); err != nil {
        fmt.Println("failed to marshal:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (h *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Delete an order by ID")
    idParam := chi.URLParam(r, "id")

    const base = 10
    const bitSize = 64
    orderID, err := strconv.ParseUint(idParam, base, bitSize)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    err = h.Repo.DeleteByID(r.Context(), orderID)
    if errors.Is(err, order.ErrNotExist) {
        w.WriteHeader(http.StatusNotFound)
        return
    } else if err != nil {
        fmt.Println("failed to find by id:", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}
