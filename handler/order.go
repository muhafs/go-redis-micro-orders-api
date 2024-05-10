package handler

import (
	"fmt"
	"net/http"
)

type Order struct {
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("create an order")
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("list of orders")
}

func (o *Order) Find(w http.ResponseWriter, r *http.Request) {
	fmt.Println("find an order")
}

func (o *Order) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("update an order")
}

func (o *Order) Delete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("delete an order")
}
