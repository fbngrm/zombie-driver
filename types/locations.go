package types

type Location struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"latitude"`
	Long float64 `json:"longitude"`
}

type LocationUpdate struct {
	UpdatedAt string  `json:"updated_at"` // RFC339
	Lat       float64 `json:"latitude"`
	Long      float64 `json:"longitude"`
}

type ZombieDriver struct {
	ID     int64 `json:"id"`
	Zombie bool  `json:"zombie"`
}
