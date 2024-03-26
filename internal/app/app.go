package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	mw "github.com/aleks0ps/gophermart/internal/app/middleware"
	"github.com/aleks0ps/gophermart/internal/app/service"
	"github.com/aleks0ps/gophermart/internal/app/storage"
)

var DSN string = "postgres://gophermart:gophermart@localhost:5432/gophermart?sslmode=disable"
var AccrualURL string = "localhost:8080"

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Init logging
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	// Flush buffer, if any
	defer logger.Sync()
	// Init logger
	sugar := logger.Sugar()
	// Init postgres
	db, err := storage.NewPGStorage(ctx, DSN, sugar)
	if err != nil {
		sugar.Fatalln(err)
		return
	}
	// Init service
	svc := &service.Service{Logger: sugar, DB: db, AccrualURL: AccrualURL}
	r := chi.NewRouter()
	r.Use(mw.DisableDefaultLogger())
	r.Use(mw.Logger(sugar))
	r.Use(mw.Gzipper())
	r.Post("/api/user/register", svc.Register)
	r.Post("/api/user/login", svc.Login)
	r.Post("/api/user/orders", svc.LoadOrder)
	r.Get("/api/user/orders", svc.GetOrders)
	http.ListenAndServe(":8088", r)
}
