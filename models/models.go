package models

import "time"

type Holding struct {
	CoinID   string    `bson:"coin_id"`
	CoinName string    `bson:"coin_name"`
	Quantity float64   `bson:"quantity"`
	BuyPrice float64   `bson:"buy_price"`
	AddedAt  time.Time `bson:"added_at"`
}

type Portfolio struct {
	UserEmail string    `bson:"user_email"`
	Holdings  []Holding `bson:"holdings"`
	UpdatedAt time.Time `bson:"updated_at"`
}
