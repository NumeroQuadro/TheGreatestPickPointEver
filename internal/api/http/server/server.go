package server

import (
	"context"
	"github.com/gorilla/mux"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/http/handler"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/http/middleware"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/http/routers"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/cache"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/db"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository/postgresql"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/tx_manager"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/workers"
	"net/http"
	"time"
)

func NewHTTPServer(ctx context.Context, dbConn db.DB, config config.Config, mng *tx_manager.TxManager, workersManager *workers.WorkerManager) *http.Server {
	baseRouter := mux.NewRouter().StrictSlash(true)

	client := cache.NewCacheClient(&config, 500)
	orderRepo := postgresql.NewOrdersRepo(mng, client)
	orderService := service.NewOrderServiceImpl(orderRepo, *mng, workersManager)
	orderHandler := handler.NewOrderHandler(orderService, workersManager)

	client.StartPeriodicUpdate(ctx, time.Duration(config.Interval), orderRepo)

	routerImpl := routers.NewRouter(baseRouter, *orderHandler)
	routerImpl.RegisterRoutes(config)

	finalHandler := middleware.AuthMiddleware(config, middleware.AuditMiddleware(workersManager, routerImpl.Router))

	workersManager.Start(ctx)

	return &http.Server{
		Addr:              config.ListenAddress,
		Handler:           finalHandler,
		ReadHeaderTimeout: 15 * time.Second,
	}
}
