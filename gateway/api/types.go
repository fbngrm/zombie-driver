package api

type Location struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"latitude"`
	Long float64 `json:"longitude"`
}
