package workers

import (
	"context"
	"fmt"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/kafka_broker"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/logger"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository/postgresql"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type WorkerManager struct {
	input        chan interface{}
	dbWorker     *WorkerDb
	stdOutWorker *WorkerStdOut
	outboxWorker *OutboxWorker
	wg           *sync.WaitGroup
	cancel       context.CancelFunc
}

func NewWorkerManager(
	client *kafka_broker.Client,
	ar postgresql.AuditRepository,
	osar postgresql.OrderStatusAuditRepository,
	or postgresql.OutboxRepository,
	bufferSize int,
	cancel func(),
	cfg config.Config,
) *WorkerManager {
	wg := sync.WaitGroup{}

	return &WorkerManager{
		input:        make(chan interface{}, bufferSize),
		dbWorker:     NewWorkerDb(or, ar, osar),
		stdOutWorker: NewWorkerStdOut(cfg.FilterWord),
		outboxWorker: NewOutboxWorker(client, &wg, or, time.Duration(5)*time.Second),
		cancel:       cancel,
		wg:           &wg,
	}
}

func (wm *WorkerManager) Start(ctx context.Context) <-chan interface{} {
	dbOut := wm.dbWorker.Work(wm.wg, ctx, wm.input)
	stdOutOut := wm.stdOutWorker.Work(wm.wg, ctx, dbOut)
	wm.outboxWorker.ProcessOutbox(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	wm.wg.Add(1)
	go func() {
		defer wm.wg.Done()
		<-sigChan
		logger.ZapLogger.Debug("Received shutdown signal. Initiating graceful shutdown...")
		signal.Stop(sigChan)
		wm.cancel()
	}()

	return stdOutOut
}

func (wm *WorkerManager) LogAudit(record interface{}) {
	fmt.Printf("new messaged logged")
	select {
	case wm.input <- record:
	default:
		logger.ZapLogger.Debug("Channel is full, skipping this job")
	}
}

func (wm *WorkerManager) Shutdown() {
	close(wm.input)
	wm.wg.Wait()

	logger.ZapLogger.Debug("WorkerManager shutdown initiated.")
}
