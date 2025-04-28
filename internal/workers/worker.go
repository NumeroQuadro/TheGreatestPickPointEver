package workers

import (
	"context"
	"fmt"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/kafka_broker"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/logger"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository/postgresql"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

type WorkerDb struct {
	ar    postgresql.AuditRepository
	osar  postgresql.OrderStatusAuditRepository
	ob    postgresql.OutboxRepository
	batch []interface{}
	timer *time.Timer
}

func NewWorkerDb(
	ob postgresql.OutboxRepository,
	ar postgresql.AuditRepository,
	osar postgresql.OrderStatusAuditRepository,
) *WorkerDb {
	return &WorkerDb{
		batch: make([]interface{}, 0),
		ob:    ob,
		ar:    ar,
		osar:  osar,
		timer: time.NewTimer(500 * time.Millisecond),
	}
}

type WorkerStdOut struct {
	FilterWord string
	batch      []interface{}
	timer      *time.Timer
}

func NewWorkerStdOut(filterWord string) *WorkerStdOut {
	return &WorkerStdOut{
		FilterWord: filterWord,
		batch:      make([]interface{}, 0),
		timer:      time.NewTimer(500 * time.Millisecond),
	}
}

func (wso *WorkerStdOut) Work(wg *sync.WaitGroup, ctx context.Context, jobs <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})

	wg.Add(1)
	go func() {
		defer close(out)
		defer wg.Done()
		defer wso.timer.Stop()

		for {
			select {
			case <-ctx.Done():
				if len(wso.batch) > 0 {
					if err := wso.processBatch(ctx, wso.batch); err != nil {
						fmt.Printf("stdout worker failed to process batch: %v\n", err)
					}
				}

				return
			case job, ok := <-jobs:
				if !ok {
					if len(wso.batch) > 0 {
						if err := wso.processBatch(ctx, wso.batch); err != nil {
							fmt.Printf("stdout worker failed to process final batch: %v\n", err)
						}
					}

					return
				}
				wso.batch = append(wso.batch, job)
				if len(wso.batch) >= 5 {
					fmt.Printf("stdout worker len(batch) >=5 --> processing batch: %v\n", len(wso.batch))
					if err := wso.processBatch(ctx, wso.batch); err != nil {
						fmt.Printf("stdout worker failed to process batch: %v\n", err)
					} else {
						for _, processedJob := range wso.batch {
							out <- processedJob
						}
					}
					wso.batch = wso.batch[:0]
					if !wso.timer.Stop() {
						<-wso.timer.C
					}
					wso.timer.Reset(500 * time.Millisecond)
				}
			case <-wso.timer.C:
				if len(wso.batch) > 0 {
					fmt.Printf("<-timer.C --> trying to process batch: %v\n", len(wso.batch))
					if err := wso.processBatch(ctx, wso.batch); err != nil {
						fmt.Printf("stdout worker failed to process batch on timer: %v\n", err)
					} else {
						for _, processedJob := range wso.batch {
							out <- processedJob
						}
					}
					wso.batch = wso.batch[:0]
				}
				wso.timer.Reset(500 * time.Millisecond)
			}
		}
	}()

	return out
}

func (wso *WorkerStdOut) processBatch(_ context.Context, batch []interface{}) error {
	for _, job := range batch {
		if wso.FilterWord != "" {
			logString := FormatAudit(job)
			if !strings.Contains(strings.ToLower(logString), strings.ToLower(wso.FilterWord)) {
				continue
			}
		}
		logger.ZapLogger.Info("worker job processed", zap.String("worker", FormatAudit(job)))
	}

	return nil
}

func FormatAudit(job interface{}) string {
	auditJob, ok := job.(domain.AuditLogRecord)
	if ok {
		return formatAuditLog(auditJob)
	}
	orderStatusAudit, ok := job.(domain.AuditOrderInfo)
	if ok {
		return formatOrderStatusLog(orderStatusAudit)
	}
	logger.ZapLogger.Debug("job cannot be used as an audit record")

	return ""
}

func formatOrderStatusLog(log domain.AuditOrderInfo) string {
	return fmt.Sprintf(`
Audit OrderStatusLog Entry:
---------------
OrderID: %d
Previous Status: %s
Current Status: %s
`,
		log.OrderID,
		log.PreviousStatus,
		log.CurrentStatus,
	)
}

func formatAuditLog(auditJob domain.AuditLogRecord) string {
	const noneString = "none"
	requestBody := noneString
	if len(auditJob.RequestBody) > 0 {
		requestBody = string(auditJob.RequestBody)
	}

	responseBody := noneString
	if len(auditJob.ResponseBody) > 0 {
		responseBody = string(auditJob.ResponseBody)
	}

	headers := make([]string, 0)
	for k, v := range auditJob.RequestHeader {
		headers = append(headers, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
	}

	return fmt.Sprintf(`
Audit LogAudit Entry:
---------------
Method: %s
Path: %s
Status Code: %d

Headers:
%s

Request RequestBody: %s

Response RequestBody: %s
`,
		auditJob.Method,
		auditJob.Path,
		auditJob.StatusCode,
		strings.Join(headers, "\n"),
		requestBody,
		responseBody)
}

func (w *WorkerDb) Work(wg *sync.WaitGroup, ctx context.Context, jobs <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})

	wg.Add(1)
	go func() {
		defer close(out)
		defer wg.Done()
		defer w.timer.Stop()

		for {
			select {
			case <-ctx.Done():
				if len(w.batch) > 0 {
					if err := w.processBatch(ctx, w.batch); err != nil {
						logger.ZapLogger.Debug("dbworker failed to process batch", zap.String("dbworker", err.Error()))
					}
				}

				return
			case job, ok := <-jobs:
				if !ok {
					if len(w.batch) > 0 {
						if err := w.processBatch(ctx, w.batch); err != nil {
							logger.ZapLogger.Debug("dbworker failed to process final batch", zap.String("dbworker", err.Error()))
						}
					}

					return
				}
				w.batch = append(w.batch, job)
				if len(w.batch) >= 5 {
					if err := w.processBatch(ctx, w.batch); err != nil {
						logger.ZapLogger.Debug("dbworker failed to process batch", zap.String("dbworker", err.Error()))
					} else {
						for _, processedJob := range w.batch {
							out <- processedJob
						}
					}
					w.batch = w.batch[:0]
					if !w.timer.Stop() {
						<-w.timer.C
					}
					w.timer.Reset(500 * time.Millisecond)
				}
			case <-w.timer.C:
				if len(w.batch) > 0 {
					if err := w.processBatch(ctx, w.batch); err != nil {
						logger.ZapLogger.Debug("dbworker failed to process batch on timer", zap.String("dbworker", err.Error()))
					} else {
						for _, processedJob := range w.batch {
							out <- processedJob
						}
					}
					w.batch = w.batch[:0]
				}
				w.timer.Reset(500 * time.Millisecond)
			}
		}
	}()

	return out
}

func (w *WorkerDb) processBatch(ctx context.Context, batch []interface{}) error {
	for _, job := range batch {
		auditRecord, ok := job.(domain.AuditLogRecord)
		if ok {
			// add to audit table
			entryID, err := w.ar.Create(ctx, auditRecord)
			if err != nil {
				logger.ZapLogger.Debug("dbworker cannot create new audit record", zap.String("dbworker", err.Error()))

				continue
			}

			// add to outbox
			taskStatus := domain.AuditLog
			_, err = w.ob.Create(ctx, entryID, taskStatus)
			if err != nil {
				return err
			}

			continue
		}

		orderStatusLog, ok := job.(domain.AuditOrderInfo)
		if ok {
			// add to audit table
			entryID, err := w.osar.Create(ctx, orderStatusLog)
			if err != nil {
				logger.ZapLogger.Debug("dbworker cannot create new audit record", zap.String("dbworker", err.Error()))

				continue
			}

			// add to outbox
			taskType := domain.OrderStatusLog
			_, err = w.ob.Create(ctx, entryID, taskType)
			if err != nil {
				logger.ZapLogger.Error("outbox cannot create an entry", zap.String("outbox", err.Error()))

				return err
			}

			continue
		}

		return fmt.Errorf("worker: invalid job type, expected AuditLogRecord or AuditOrderInfo")
	}

	return nil
}

type OutboxWorker struct {
	client   *kafka_broker.Client
	wg       *sync.WaitGroup
	repo     postgresql.OutboxRepository
	interval time.Duration
}

func NewOutboxWorker(client *kafka_broker.Client, wg *sync.WaitGroup, repo postgresql.OutboxRepository, interval time.Duration) *OutboxWorker {
	return &OutboxWorker{
		client:   client,
		wg:       wg,
		repo:     repo,
		interval: interval,
	}
}

func (ow *OutboxWorker) ProcessOutbox(ctx context.Context) {
	ow.wg.Add(1)
	go func() {
		ticker := time.NewTicker(ow.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tasks, err := ow.repo.FetchAndMarkProcessing(ctx, 10)
				if err != nil {
					logger.ZapLogger.Error("fetch tasks failed", zap.String("obworker", err.Error()))

					continue
				}
				if len(tasks) == 0 {
					continue
				}

				var failedIDs []int64
				for _, t := range tasks {
					key := fmt.Sprint(t.TaskID)
					if err := ow.client.Publish(ctx, key, t); err != nil {
						logger.ZapLogger.Error("failed publishing task_id", zap.String("obworker", fmt.Sprintf("task: %d, error: %v", t.TaskID, err)))

						failedIDs = append(failedIDs, t.TaskID)
					}
				}

				if err := ow.repo.DeleteSuccessful(ctx, tasks, failedIDs); err != nil {
					logger.ZapLogger.Error("failed cleanup tasks", zap.String("obworker", err.Error()))
				}
			}
		}
	}()
}
