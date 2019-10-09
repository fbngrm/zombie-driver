package server

type LocationUpdate struct {
	UpdatedAt string  `json:"updated_at"`
	Lat       float64 `json:"latitude"`
	Long      float64 `json:"longitude"`
}

type ZombieDriver struct {
	ID     int64 `json:"id"`
	Zombie bool  `json:"zombie"`
}
