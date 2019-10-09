package server

type Location struct {
	ID   string  `json:"id"`
	Lat  float32 `json:"latitude"`
	Long float32 `json:"longitude"`
}

type LocationUpdate struct {
	UpdatedAt string  `json:"updated_at"`
	Lat       float32 `json:"latitude"`
	Long      float32 `json:"longitude"`
}
