package repository

import (
	"fmt"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"time"
)

type Filter struct {
	OrderID        *int64
	UserID         *int64
	ExpirationTime *time.Time
	Status         *domain.Status
	Weight         *int
	Cost           *int
	SearchTerm     *string
}

func BuildSQLQuery(filter Filter) (string, []interface{}) {
	baseQuery := `
        SELECT order_id, user_id, expiration_date, status, last_changed_at, weight, cost
        FROM orders
        WHERE 1=1
    `
	var values []interface{}

	argPos := 1
	if filter.UserID != nil {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argPos)
		values = append(values, *filter.UserID)
		argPos++
	}
	if filter.ExpirationTime != nil {
		baseQuery += fmt.Sprintf(" AND expiration_date = $%d", argPos)
		values = append(values, *filter.ExpirationTime)
		argPos++
	}
	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND status = $%d", argPos)
		values = append(values, *filter.Status)
		argPos++
	}
	if filter.Cost != nil {
		baseQuery += fmt.Sprintf(" AND cost = $%d", argPos)
		values = append(values, *filter.Cost)
		argPos++
	}
	if filter.Weight != nil {
		baseQuery += fmt.Sprintf(" AND weight = $%d", argPos)
		values = append(values, *filter.Weight)
		argPos++
	}
	if filter.SearchTerm != nil {
		baseQuery += fmt.Sprintf(" AND (CAST(order_id AS TEXT) LIKE $%d OR status LIKE $%d)", argPos, argPos)
		values = append(values, *filter.SearchTerm)
	}

	return baseQuery, values
}

func (filter *Filter) GetFilterStringView() string {
	filterString := ""
	if filter.Status != nil {
		filterString += "status,"
	}
	if filter.UserID != nil {
		filterString += "user_id,"
	}
	if filter.OrderID != nil {
		filterString += "order_id,"
	}
	if filter.Cost != nil {
		filterString += "cost,"
	}
	if filter.Weight != nil {
		filterString += "weight,"
	}
	if filter.ExpirationTime != nil {
		filterString += "expiration_time,"
	}
	if filter.SearchTerm != nil {
		filterString += "search_term,"
	}

	return filterString
}
