package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muhafs/orders-api/model"
	"github.com/muhafs/orders-api/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println("failed to extract request:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()

	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	if err := o.Repo.Insert(r.Context(), order); err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(order)
	if err != nil {
		fmt.Println("failed to encode data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
	w.WriteHeader(http.StatusCreated)
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const (
		decimal = 10
		bitSize = 64
	)
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		fmt.Println("invalid cursor:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	res, err := o.Repo.List(r.Context(), order.ListPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		fmt.Println("failed to list data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}
	response.Items = res.Orders
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to encode data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (o *Order) Find(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const (
		decimal = 10
		bitSize = 64
	)
	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		fmt.Println("invalid id:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := o.Repo.Find(r.Context(), orderID)
	if errors.Is(err, order.ErrNotFound) {
		fmt.Println("order not found:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		fmt.Println("failed encode data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) Update(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println("failed to extract request:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	const (
		decimal = 10
		bitSize = 64
	)
	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		fmt.Println("invalid id:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := o.Repo.Find(r.Context(), orderID)
	if errors.Is(err, order.ErrNotFound) {
		fmt.Println("order not found:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const (
		completedStatus = "completed"
		shippedStatus   = "shipped"
	)
	now := time.Now().UTC()

	switch body.Status {
	case shippedStatus:
		if res.ShippedAt != nil {
			fmt.Println("shipping date already created")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		res.ShippedAt = &now

	case completedStatus:
		if res.CompletedAt != nil || res.ShippedAt == nil {
			fmt.Println("order has arrived")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		res.CompletedAt = &now
	default:
		fmt.Println("order status undefined")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = o.Repo.Update(r.Context(), res); err != nil {
		fmt.Println("failed to update data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(res); err != nil {
		fmt.Println("failed encode data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) Delete(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const (
		decimal = 10
		bitSize = 64
	)
	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		fmt.Println("invalid id:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = o.Repo.Delete(r.Context(), orderID)
	if errors.Is(err, order.ErrNotFound) {
		fmt.Println("order not found:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to delete:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
