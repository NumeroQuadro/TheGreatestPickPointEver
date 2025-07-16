package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	tech_monitoring "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/monitoring"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/tracing"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/logger"
	domain_monitoring "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/monitoring"

	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/server"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/db"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/kafka_broker"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository/postgresql"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/tx_manager"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/workers"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %+v\n", cfg)
	config.ApplyEnvironmentVariables(cfg)

	ctx := context.Background()
	shutdown := tracing.InitTracer()
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConn, err := db.Open(ctx, *cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	kfCfg := kafka_broker.Config{
		Brokers:    cfg.Kafka.Brokers,
		Topic:      cfg.Kafka.Topic,
		GroupID:    cfg.Kafka.GroupID,
		Partition:  cfg.Kafka.Partition,
		BufferSize: cfg.Kafka.BufferSize,
		ReadTO:     time.Duration(cfg.Kafka.ReadTimeoutSeconds) * time.Second,
		WriteTO:    time.Duration(cfg.Kafka.WriteTimeoutSeconds) * time.Second,
	}
	kafkaClient, err := kafka_broker.NewClient(kfCfg)
	if err != nil {
		log.Fatalf("Kafka failed: %v", err)
	}

	txManager := tx_manager.NewTxManager(dbConn)

	auditRepo := postgresql.NewAuditRepositoryImpl(dbConn)
	orderStatusAuditRepo := postgresql.NewOrderStatusAuditRepositoryImpl(dbConn)
	outboxRepo := postgresql.NewOutboxRepositoryImpl(txManager)
	workersManager := workers.NewWorkerManager(kafkaClient, auditRepo, orderStatusAuditRepo, outboxRepo, 5, cancel, *cfg)
	workersManager.Start(ctx)

	tech_monitoring.RegisterBusinessMetrics()
	domain_monitoring.RegisterBusinessMetrics()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		tech_monitoring.StartMetricsServer(cfg)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Run(*cfg, txManager, workersManager)
	}()

	wg.Wait()

	defer logger.ZapLogger.Sync()
}

