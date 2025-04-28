package handler

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)

		return
	}

	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		http.Error(w, "invalid order ID format", http.StatusBadRequest)

		return
	}

	order, err := h.service.GetOrderByID(r.Context(), orderIDInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	response := OrderListResponse{
		Order: order,
	}

	_ = h.writeResponseToHeader(response, w)
}
