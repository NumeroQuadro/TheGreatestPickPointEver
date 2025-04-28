package postgresql

import (
	"context"
	"fmt"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/tx_manager"
)

type OutboxRepository interface {
	Create(ctx context.Context, entryID int64, taskType domain.TaskType) (int64, error)
	FetchAndMarkProcessing(ctx context.Context, limit int) ([]domain.Task, error)
	DeleteSuccessful(ctx context.Context, tasks []domain.Task, failedIDs []int64) error
}

type OutboxRepositoryImpl struct {
	tx *tx_manager.TxManager
}

func NewOutboxRepositoryImpl(tx *tx_manager.TxManager) *OutboxRepositoryImpl {
	return &OutboxRepositoryImpl{
		tx: tx,
	}
}
func (r *OutboxRepositoryImpl) Create(ctx context.Context, entryID int64, taskType domain.TaskType) (int64, error) {
	var id int64
	query := `
		INSERT INTO outbox (entry_id, task_type, task_status, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING task_id;
	`
	err := r.tx.GetQueryEngine(ctx).ExecQueryRow(ctx, query, entryID, taskType, domain.Created).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create outbox task: %w", err)
	}

	return id, nil
}

func (r *OutboxRepositoryImpl) FetchAndMarkProcessing(ctx context.Context, limit int) ([]domain.Task, error) {
	const q = `
WITH cte AS (
    SELECT task_id
      FROM outbox
     WHERE (task_status = 'CREATED' OR task_status = 'FAILED')
       AND attempts_count < 3
       AND next_attempt_at <= NOW()
     ORDER BY created_at
     LIMIT $1
     FOR UPDATE SKIP LOCKED
)
UPDATE outbox
   SET task_status = 'PROCESSING',
       updated_at  = NOW()
 WHERE task_id IN (SELECT task_id FROM cte)
RETURNING task_id, task_status, task_type, entry_id,
          attempts_count, next_attempt_at
`
	var tasks []domain.Task
	if err := r.tx.GetQueryEngine(ctx).
		Select(ctx, &tasks, q, limit); err != nil {
		return nil, fmt.Errorf("fetch tasks: %w", err)
	}

	return tasks, nil
}

func (r *OutboxRepositoryImpl) DeleteSuccessful(ctx context.Context, tasks []domain.Task, failedIDs []int64) error {
	var successIDs []int64
	for _, t := range tasks {
		var failed bool
		for _, fid := range failedIDs {
			if t.TaskID == fid {
				failed = true

				break
			}
		}
		if !failed {
			successIDs = append(successIDs, t.TaskID)
		}
	}

	if len(successIDs) > 0 {
		delQuery := `DELETE FROM outbox WHERE task_id = ANY($1)`
		if _, err := r.tx.GetQueryEngine(ctx).Exec(ctx, delQuery, successIDs); err != nil {
			return fmt.Errorf("delete successful tasks: %w", err)
		}
	}
	if len(failedIDs) > 0 {
		const updQuery = `
UPDATE outbox
   SET attempts_count = attempts_count + 1,
       task_status = CASE
         WHEN attempts_count + 1 >= 3 THEN 'NO_ATTEMPTS_LEFT'
         ELSE 'FAILED'
       END,
       next_attempt_at = CASE
         WHEN attempts_count + 1 < 3 THEN NOW() + INTERVAL '2 seconds'
         ELSE next_attempt_at
       END,
       finished_at = CASE
         WHEN attempts_count + 1 >= 3 THEN NOW()
         ELSE finished_at
       END,
       updated_at = NOW()
 WHERE task_id = ANY($1)
`
		if _, err := r.tx.GetQueryEngine(ctx).Exec(ctx, updQuery, failedIDs); err != nil {
			return fmt.Errorf("update failed tasks: %w", err)
		}
	}

	return nil
}
