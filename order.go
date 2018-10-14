package main

type Order struct {
	Destination Location `json:"destination"`
	Source      Location `json:"source"`
}

