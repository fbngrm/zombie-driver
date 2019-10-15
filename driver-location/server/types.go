package server

type Location struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"latitude"`
	Long float64 `json:"longitude"`
}

type LocationUpdate struct {
	UpdatedAt string  `json:"updated_at"`
	Lat       float64 `json:"latitude"`
	Long      float64 `json:"longitude"`
}
