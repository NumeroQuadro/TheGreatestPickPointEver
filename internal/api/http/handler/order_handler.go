//go:generate mockgen -source ./order_handler.go -destination=./mocks/order_handler.go -package=mock_handler
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/workers"
	"io"
	"net/http"
	"strconv"
)

type OrderHandler struct {
	service       OrderService
	workerManager *workers.WorkerManager
}

func NewOrderHandler(service *service.OrderServiceImpl, workerManager *workers.WorkerManager) *OrderHandler {
	return &OrderHandler{
		service:       service,
		workerManager: workerManager,
	}
}

type OrderService interface {
	AddOrder(ctx context.Context,
		orderDto service.OrderDto,
		packageType domain.PackageType,
		isAdditionalFilm bool) (domain.Order, error)
	RetrieveOrdersFromFile(ctx context.Context,
		data []byte) error
	CompleteOrder(ctx context.Context,
		orderID int64,
		userID int64) error
	ReturnOrder(ctx context.Context,
		orderID int64) error
	RefundOrder(ctx context.Context,
		orderID int64,
		expirationDays int) error
	GetOrdersByUserID(ctx context.Context,
		userID int64,
		limit *int,
		lastID *int64) ([]domain.Order, error)
	GetOrderByID(ctx context.Context,
		orderID int64) (domain.Order, error)
	GetOrders(ctx context.Context,
		lastID *int64,
		limit *int,
		searchFilter *service.SearchFilter) []domain.Order
	GetOrdersBySpecificStatus(ctx context.Context,
		status domain.Status) []domain.Order
	GetRefundedOrders(ctx context.Context,
		lastID *int64,
		limit *int) ([]domain.Order, error)
}

func (h *OrderHandler) getRequestBody(
	w http.ResponseWriter,
	req *http.Request,
	isGetRequest bool,
) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionRequired)

		return nil, nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(req.Body)

	return body, nil
}

func (h *OrderHandler) isOrderValid(oc CreateOrderRequest) bool {
	if oc.Status == "" {
		return false
	}
	if oc.UserID <= 0 {
		return false
	}

	if oc.Cost <= 0 {
		return false
	}

	if oc.Weight <= 0 {
		return false
	}

	return true
}

func (h *OrderHandler) getPaginationVars(lastID string, limit string) (int64, int, error) {
	lastIDInt, err := strconv.ParseInt(lastID, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("last_id is not valid")
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return 0, 0, fmt.Errorf("limit is not valid")
	}

	return lastIDInt, limitInt, nil
}

func (h *OrderHandler) listOrdersByUserID(
	w http.ResponseWriter,
	r *http.Request,
	userID,
	lastID,
	limit string,
) error {
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return err
	}

	var (
		lastIDPtr *int64
		limitPtr  *int
	)

	if lastID != "" && limit != "" {
		lastIDInt, limitInt, err := h.getPaginationVars(lastID, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return err
		}
		lastIDPtr = &lastIDInt
		limitPtr = &limitInt
	}

	orders, err := h.service.GetOrdersByUserID(r.Context(), userIDInt, limitPtr, lastIDPtr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return err
	}

	response := OrdersListResponse{Orders: orders}

	return h.writeResponseToHeader(response, w)
}

func (h *OrderHandler) writeResponseToHeader(dest interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")

	responseBody, err := json.Marshal(dest)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)

		return err
	}

	if _, err := w.Write(responseBody); err != nil {
		return err
	}

	return nil
}

func (h *OrderHandler) writeOkResponseToHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

type contextKey string

const AuditInfoKey contextKey = "orderInfo"

type AuditResponseWriter struct {
	http.ResponseWriter
	AuditData *domain.AuditLogData
}

func (arw *AuditResponseWriter) WriteHeader(statusCode int) {
	arw.AuditData.HTTPStatus = statusCode
	arw.ResponseWriter.WriteHeader(statusCode)
}

func (arw *AuditResponseWriter) Write(body []byte) (int, error) {
	arw.AuditData.ResponseBody = append(arw.AuditData.ResponseBody, body...)

	return arw.ResponseWriter.Write(body)
}
