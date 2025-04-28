package handler

import (
	"github.com/gorilla/mux"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service"
	"log"
	"net/http"
)

type OrdersListResponse struct {
	Orders []domain.Order
}

type OrderListResponse struct {
	Order domain.Order
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	vars := mux.Vars(r)

	userID := vars["user_id"]
	status := vars["status"]
	lastID := vars["last_id"]
	limit := query.Get("limit")
	searchTerm := query.Get("search")

	if status != "" {
		if domain.IsStatusValid(status) {
			http.Error(w, "Invalid package type", http.StatusBadRequest)

			return
		}
	}

	if userID != "" {
		err := h.listOrdersByUserID(w, r, userID, lastID, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // (BadRequest or InternalError) -> TODO: process error in repo and switch case error.Is
		}

		return
	}

	var searchFilter *service.SearchFilter
	if searchTerm != "" {
		sf := service.SearchFilter{SearchTerm: &searchTerm}
		searchFilter = &sf
	}

	lastIDInt, limitInt, err := h.getPaginationVars(lastID, limit)
	if err != nil {
		orders := h.service.GetOrders(r.Context(), nil, nil, searchFilter)

		response := OrdersListResponse{
			Orders: orders,
		}
		err = h.writeResponseToHeader(response, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if len(orders) == 0 {
			log.Println("ListOrders: no orders found")
			response := OrdersListResponse{
				Orders: orders,
			}
			err = h.writeResponseToHeader(response, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		return
	}
	orders := h.service.GetOrders(r.Context(), &lastIDInt, &limitInt, searchFilter)

	response := OrdersListResponse{
		Orders: orders,
	}
	err = h.writeResponseToHeader(response, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
