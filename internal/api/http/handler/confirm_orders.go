package handler

import (
	"io"
	"net/http"
)

func (h *OrderHandler) ConfirmOrders(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)

		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not found in form data", http.StatusBadRequest)

		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)

		return
	}

	err = h.service.RetrieveOrdersFromFile(r.Context(), fileBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	h.writeOkResponseToHeader(w)
}
