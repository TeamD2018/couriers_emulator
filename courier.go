package main

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Location struct {
	Point *GeoPoint `json:"point"`
}

type Courier struct {
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Phone    *string   `json:"phone,omitempty"`
	Location *Location `json:"location,omitempty"`
}
