package service

import (
	"encoding/json"
	"fmt"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"strconv"
	"time"
)

type SearchFilter struct {
	OrderID        *int64
	UserID         *int64
	Status         *string
	ExpirationTime *time.Time
	SearchTerm     *string
}

type External struct {
	OrderID        int64     `json:"-"`
	OrderIDRaw     string    `json:"order_id"`
	UserID         int64     `json:"-"`
	UserIDRaw      string    `json:"user_id"`
	ExpirationTime time.Time `json:"expiration_time"`
	Weight         int       `json:"weight"`
	Cost           int       `json:"cost"`
}

func (e *External) UnmarshalJSON(data []byte) error {
	type Alias External
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	orderID, err := strconv.ParseInt(e.OrderIDRaw, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid order_id: %w", err)
	}
	e.OrderID = orderID

	userID, err := strconv.ParseInt(e.UserIDRaw, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}
	e.UserID = userID

	return nil
}

type OrderDto struct {
	OrderID        int64         `db:"order_id"`
	UserID         int64         `db:"user_id"`
	ExpirationTime time.Time     `db:"expiration_date"`
	Status         domain.Status `db:"status"`
	Weight         int           `db:"weight"`
	Cost           int           `db:"cost"`
}

func ConvertDtoToDomainOrder(dto OrderDto) domain.Order {
	return domain.Order{
		OrderID:        dto.OrderID,
		UserID:         dto.UserID,
		ExpirationTime: dto.ExpirationTime,
		Status:         dto.Status,
		Weight:         dto.Weight,
		Cost:           dto.Cost,
	}
}
