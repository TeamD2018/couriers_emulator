package main

type Order struct {
	ID          string   `json:"id"`
	Destination Location `json:"destination"`
	Source      Location `json:"source"`
	OrderNumber int      `json:"order_number"`
}
