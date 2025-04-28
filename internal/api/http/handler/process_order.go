package handler

import (
	"github.com/gorilla/mux"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"net/http"
	"strconv"
)

func (h *OrderHandler) ProcessOrder(config config.Config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]
	userID := vars["user_id"]
	action := vars["action"]

	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		http.Error(w, "invalid order ID format", http.StatusBadRequest)

		return
	}
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		http.Error(w, "invalid order ID format", http.StatusBadRequest)

		return
	}

	switch action {
	case "complete":
		err = h.service.CompleteOrder(r.Context(), orderIDInt, userIDInt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	case "refund":
		err = h.service.RefundOrder(r.Context(), orderIDInt, config.OrderExpirationDays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

	default:
		statusError := http.StatusBadRequest
		http.Error(w, "Invalid action", statusError)

		return
	}

	h.writeOkResponseToHeader(w)
}
