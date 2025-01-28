package entity

import "time"

type User struct {
	ID      int
	Balance float64
}

type Transaction struct {
	ID        int
	UserID    int
	Amount    float64
	Type      string
	CreatedAt time.Time
}
