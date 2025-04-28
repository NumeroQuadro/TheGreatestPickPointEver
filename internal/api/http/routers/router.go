package routers

import (
	"github.com/gorilla/mux"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/http/handler"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"net/http"
)

type Router interface {
	RegisterRoutes(config config.Config)
}

type RouterImpl struct {
	Router  *mux.Router
	Handler handler.OrderHandler
}

func NewRouter(rt *mux.Router, h handler.OrderHandler) *RouterImpl {
	return &RouterImpl{Router: rt, Handler: h}
}

func (r *RouterImpl) RegisterRoutes(config config.Config) {
	ordersRouter := r.Router.PathPrefix("/orders").Subrouter()

	ordersRouter.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			r.Handler.ListOrders(w, req)
		case http.MethodPost:
			r.Handler.ConfirmOrder(w, req)
		}
	})
	ordersRouter.HandleFunc("/batch", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			r.Handler.ConfirmOrders(w, req)
		}
	})
	ordersRouter.HandleFunc("/{action}/{id:[0-9]+}", func(w http.ResponseWriter, req *http.Request) {
		r.Handler.ReturnOrder(w, req)
	}).Methods("PUT")
	ordersRouter.HandleFunc("/{action}/{id:[0-9]+}/{user_id:[0-9]+}", func(w http.ResponseWriter, req *http.Request) {
		r.Handler.ProcessOrder(config, w, req)
	}).Methods("PUT")
	ordersRouter.HandleFunc("/{id:[0-9]+}", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			r.Handler.GetOrderByID(w, req)
		}
	})
	ordersRouter.HandleFunc("/{orders}/{search:[0-9]+}", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			r.Handler.ListOrders(w, req)
		}
	})
}
