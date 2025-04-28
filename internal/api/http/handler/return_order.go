package handler

import (
	"errors"
	"github.com/gorilla/mux"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"net/http"
	"strconv"
)

func (h *OrderHandler) ReturnOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]
	action := vars["action"]

	if action != "return" {
		http.Error(w, "action not allowed", http.StatusMethodNotAllowed)

		return
	}

	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.service.ReturnOrder(r.Context(), orderIDInt)
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)

		return
	default:
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	}

	h.writeOkResponseToHeader(w)
}
