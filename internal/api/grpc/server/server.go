package server

import (
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/cache"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository/postgresql"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/tx_manager"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/workers"
	"google.golang.org/grpc/reflection"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	orderpb "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/generated"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/interceptors"
	grpcservice "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/service"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/logger"
	service "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service"
)

func Run(config config.Config, mng *tx_manager.TxManager, workersManager *workers.WorkerManager) {
	tracer := otel.Tracer("order-service")

	interceptor := interceptors.MetricsAndLoggingInterceptor(logger.ZapLogger, tracer)

	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptor)),
	)

	reflection.Register(s)

	client := cache.NewCacheClient(&config, 500)
	orderRepo := postgresql.NewOrdersRepo(mng, client)
	orderService := service.NewOrderServiceImpl(orderRepo, *mng, workersManager)
	orderServer := grpcservice.NewOrderServiceServer(orderService, config)

	orderpb.RegisterOrderServiceServer(s, orderServer)

	lis, err := net.Listen("tcp", config.GRPCListenAddress)
	if err != nil {
		logger.ZapLogger.Fatal("Failed to listen", zap.Error(err))
	}

	logger.ZapLogger.Info("gRPC server is listening...", zap.String("addr", lis.Addr().String()))
	if err := s.Serve(lis); err != nil {
		logger.ZapLogger.Fatal("Failed to serve", zap.Error(err))
	}
}
