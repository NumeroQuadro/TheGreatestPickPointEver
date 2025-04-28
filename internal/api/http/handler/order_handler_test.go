package handler

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service"
	mock_repository "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrderHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_repository.NewMockOrderRepository(ctrl)
	handler := NewOrderHandler(&service.OrderServiceImpl{repo: mockRepo})

	t.Run("validateRequestBody", func(t *testing.T) {
		t.Parallel()

		t.Run("empty body", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			w := httptest.NewRecorder()

			body, _, failed := handler.validateRequestBody(w, req, true)

			require.True(t, failed)
			require.Nil(t, body)
			require.Equal(t, http.StatusBadRequest, w.Code)
			require.Equal(t, "body is empty\n", w.Body.String())
		})

		t.Run("valid body", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`{"key":"value"}`)))
			w := httptest.NewRecorder()

			body, _, failed := handler.validateRequestBody(w, req, false)

			require.False(t, failed)
			require.Equal(t, `{"key":"value"}`, string(body))
			require.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("read error", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", &errorReader{})
			w := httptest.NewRecorder()

			body, _, failed := handler.validateRequestBody(w, req, false)

			require.True(t, failed)
			require.Nil(t, body)
			require.Equal(t, http.StatusPreconditionRequired, w.Code)
			require.Contains(t, w.Body.String(), "mock read error")
		})
	})

	t.Run("isOrderValid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			req  CreateOrderRequest
			want bool
		}{
			{name: "valid", req: CreateOrderRequest{Status: "confirmed", UserID: 1, Cost: 100, Weight: 10}, want: true},
			{name: "empty status", req: CreateOrderRequest{Status: "", UserID: 1, Cost: 100, Weight: 10}, want: false},
			{name: "invalid userID", req: CreateOrderRequest{Status: "confirmed", UserID: 0, Cost: 100, Weight: 10}, want: false},
			{name: "invalid userID", req: CreateOrderRequest{Status: "confirmed", UserID: -1, Cost: 100, Weight: 10}, want: false},
			{name: "invalid cost", req: CreateOrderRequest{Status: "confirmed", UserID: 1, Cost: 0, Weight: 10}, want: false},
			{name: "invalid weight", req: CreateOrderRequest{Status: "confirmed", UserID: 1, Cost: 100, Weight: 0}, want: false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := handler.isOrderValid(tt.req)
				require.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("getPaginationVars", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			lastId   string
			limit    string
			wantLast int64
			wantLim  int
			wantErr  bool
		}{
			{name: "valid", lastId: "10", limit: "5", wantLast: 10, wantLim: 5, wantErr: false},
			{name: "invalid lastId", lastId: "abc", limit: "5", wantErr: true},
			{name: "invalid limit", lastId: "10", limit: "xyz", wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				lastId, limit, err := handler.getPaginationVars(tt.lastId, tt.limit)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tt.wantLast, lastId)
				require.Equal(t, tt.wantLim, limit)
			})
		}
	})

	t.Run("listOrdersByUserID", func(t *testing.T) {
		t.Parallel()

		t.Run("invalid userId", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/orders?userId=abc", nil)
			w := httptest.NewRecorder()

			handler.listOrdersByUserID(w, req, "abc", "", "")

			require.Equal(t, http.StatusBadRequest, w.Code)
			require.Contains(t, w.Body.String(), "invalid syntax")
		})

		t.Run("invalid pagination", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/orders?userId=1&lastId=abc&limit=5", nil)
			w := httptest.NewRecorder()

			handler.listOrdersByUserID(w, req, "1", "abc", "5")

			require.Equal(t, http.StatusBadRequest, w.Code)
			require.Contains(t, w.Body.String(), "last_id is not valid")
		})

		t.Run("service error", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/orders?userId=1", nil)
			w := httptest.NewRecorder()

			filter := repository.Filter{UserID: int64Ptr(1)}
			mockRepo.EXPECT().FindAll(req.Context(), filter, nil, nil).
				Return(nil, domain.ErrOrderNotFound)
			handler.listOrdersByUserID(w, req, "1", "", "")

			require.Equal(t, http.StatusInternalServerError, w.Code)
			require.Contains(t, w.Body.String(), "order not found")
		})
	})
}

type errorReader struct{}

func (e *errorReader) Read([]byte) (n int, err error) {
	return 0, fmt.Errorf("mock read error")
}

func int64Ptr(i int64) *int64 { return &i }
