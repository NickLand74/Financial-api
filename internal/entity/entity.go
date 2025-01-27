package entity

type User struct {
	ID      int
	Balance float64
}

type Transaction struct {
	ID     int
	UserID int
	Amount float64
	Type   string // "topup" or "transfer"
}
