package middleware

import (
	"bytes"
	"fmt"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/http/handler"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/workers"
	"io"
	"net/http"
)

// AuthMiddleware checks basic auth and set StatusUnauthorized if basic auth failed
func AuthMiddleware(config config.Config, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userName, password, ok := req.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}
		if userName != config.TestCredentials.Username || password != config.TestCredentials.Password {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		fmt.Println("\033[35m"+userName, password+"\033[0m")

		handler.ServeHTTP(w, req)
	}
}

func AuditMiddleware(wm *workers.WorkerManager, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var d domain.AuditLogData
		arw := handler.AuditResponseWriter{
			ResponseWriter: w,
			AuditData:      &d,
		}
		d.Request = req

		body, err := io.ReadAll(req.Body)
		d.RequestBody = body
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)

			return
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body))

		h.ServeHTTP(&arw, req)

		r := domain.NewAuditLogRecord(&d)
		wm.LogAudit(r)
	}
}
