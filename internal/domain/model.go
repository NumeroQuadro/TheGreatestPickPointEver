package domain

import (
	"time"
)

type Order struct {
	OrderID        int64     `db:"order_id"`
	UserID         int64     `db:"user_id"`
	ExpirationTime time.Time `db:"expiration_date"`
	Status         Status    `db:"status"`
	Weight         int       `db:"weight"`
	Cost           int       `db:"cost"`
	LastChangedAt  time.Time `db:"last_changed_at"`
}

func NewOrder(
	orderID int64,
	userID int64,
	expirationTime time.Time,
	status Status,
	weight int,
	cost int,
) (Order, error) {
	if weight < 0 || cost < 0 {
		return Order{}, ErrOrderFieldsAreIncorrect
	}

	return Order{
		OrderID:        orderID,
		UserID:         userID,
		ExpirationTime: expirationTime,
		Status:         status,
		Weight:         weight,
		Cost:           cost,
		LastChangedAt:  time.Now(),
	}, nil
}
