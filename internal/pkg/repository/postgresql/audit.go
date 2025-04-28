package postgresql

import (
	"context"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/db"
)

type AuditRepository interface {
	Create(ctx context.Context, job domain.AuditLogRecord) (int64, error)
}

type AuditRepositoryImpl struct {
	db db.DB
}

func NewAuditRepositoryImpl(database db.DB) *AuditRepositoryImpl {
	return &AuditRepositoryImpl{db: database}
}

func (a *AuditRepositoryImpl) Create(ctx context.Context, job domain.AuditLogRecord) (int64, error) {
	var entryID int64

	query := `
		INSERT INTO audit_logs (
			method, path,
			request_header, request_body, query_params,
			status_code, response_body
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) returning entry_id;
	`

	err := a.db.ExecQueryRow(ctx,
		query,
		job.Method,
		job.Path,
		job.RequestHeader,
		job.RequestBody,
		job.QueryParams,
		job.StatusCode,
		job.ResponseBody,
	).Scan(&entryID)

	if err != nil {
		return 0, err
	}

	return entryID, nil
}
