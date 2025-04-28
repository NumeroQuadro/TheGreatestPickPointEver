package domain

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type AuditLogRecord struct {
	EntryID       int64             `json:"entry_id" db:"entry_id"`
	Method        string            `json:"method" db:"method"`
	Path          string            `json:"path" db:"path"`
	RequestHeader http.Header       `json:"request_header" db:"request_header"`
	RequestBody   json.RawMessage   `json:"request_body,omitempty" db:"request_body,omitempty"`
	QueryParams   map[string]string `json:"query_params,omitempty" db:"query_params,omitempty"`
	StatusCode    int               `json:"status_code" db:"status_code"`
	ResponseBody  json.RawMessage   `json:"response_body,omitempty" db:"response_body,omitempty"`
	CreatedAt     time.Time         `db:"created_at"`
}

type AuditLogData struct {
	HTTPStatus   int
	Request      *http.Request
	ResponseBody []byte
	RequestBody  []byte
}

type AuditOrderInfo struct {
	EntryID        int64     `json:"entry_id" db:"entry_id"`
	OrderID        int64     `json:"order_id" db:"order_id"`
	PreviousStatus Status    `json:"previous_status" db:"previous_status"`
	CurrentStatus  Status    `json:"current_status" db:"current_status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

func NewAuditOrderInfo(orderID int64, previousStatus Status, currentStatus Status) *AuditOrderInfo {
	return &AuditOrderInfo{
		OrderID:        orderID,
		PreviousStatus: previousStatus,
		CurrentStatus:  currentStatus,
	}
}

func NewAuditLogRecord(d *AuditLogData) AuditLogRecord {
	vars := mux.Vars(d.Request)

	validRequestBody := d.RequestBody
	if len(validRequestBody) == 0 {
		validRequestBody = []byte("null")
	}

	validResponseBody := d.ResponseBody
	if len(validResponseBody) == 0 {
		validResponseBody = []byte("null")
	}

	return AuditLogRecord{
		Method:        d.Request.Method,
		Path:          d.Request.URL.Path,
		RequestHeader: d.Request.Header,
		RequestBody:   validRequestBody,
		QueryParams:   vars,
		StatusCode:    d.HTTPStatus,
		ResponseBody:  validResponseBody,
	}
}
