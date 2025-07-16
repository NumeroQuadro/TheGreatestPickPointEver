package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service"
)

type CreateOrderRequest struct {
	OrderID          int64     `json:"order_id"`
	UserID           int64     `json:"user_id"`
	ExpirationTime   time.Time `json:"expiration_time"`
	Status           string    `json:"status"`
	Weight           int       `json:"weight"`
	Cost             int       `json:"cost"`
	PackageType      string    `json:"package_type"`
	IsAdditionalFilm bool      `json:"is_additional_film"`
}

type CreateOrderResponse struct {
	OrderID int64
}

func (h *OrderHandler) ConfirmOrder(w http.ResponseWriter, req *http.Request) {
	body, err := h.getRequestBody(w, req, false)
	if err != nil || string(body) == "" {
		http.Error(w, "body is incorrect", http.StatusBadRequest)

		return
	}

	var oc CreateOrderRequest
	if err := json.Unmarshal(body, &oc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if !h.isOrderValid(oc) {
		http.Error(w, "order is not valid", http.StatusBadRequest)

		return
	}

	packageTypeFromString, err := domain.GetPackageTypeFromString(oc.PackageType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	dto := service.OrderDto{
		OrderID:        oc.OrderID,
		UserID:         oc.UserID,
		ExpirationTime: oc.ExpirationTime,
		Weight:         oc.Weight,
		Cost:           oc.Cost,
	}

	order, err := h.service.AddOrder(req.Context(), dto, packageTypeFromString, oc.IsAdditionalFilm)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)

			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	}

	response := CreateOrderResponse{
		OrderID: order.OrderID,
	}
	_ = h.writeResponseToHeader(response, w)
}
