package main

import (
	"avito-pvz/database"
	"avito-pvz/database/migrations"
	"avito-pvz/internal/config"
	"avito-pvz/internal/handler"
	"avito-pvz/internal/logger"
	"avito-pvz/internal/metrics"
	"avito-pvz/internal/repository"
	"avito-pvz/internal/routes"
	"avito-pvz/internal/routes/middleware"

	"avito-pvz/internal/service"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	log := logger.InitLogger(cfg)
	metrics.StartMetricsServer()
	log.Info().Msg("Starting metrics server on :9000")

	db, err := database.ConnectDB(cfg, log)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer database.CloseDB(db, log)
	err = migrations.RunMigration(context.Background(), db, log)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	r := gin.Default()
	r.Use(middleware.MetricsMiddleware())

	serverAdress := cfg.AppHost + ":" + cfg.AppPort

	userRepository := repository.NewUserRepository(db, log)
	userService := service.NewUserService(userRepository, cfg, log)
	userHandler := handler.NewUserHandler(userService, log)

	pvzRepository := repository.NewPvzRepository(db, log)
	pvzService := service.NewPvzService(pvzRepository, log)
	pvzHandler := handler.NewPvzHandler(pvzService, log)

	receptionRepository := repository.NewReceptionRepository(db, log)
	receptionService := service.NewReceptionService(receptionRepository, log)
	receptionHandler := handler.NewReceptionHandler(receptionService, log)

	productRepository := repository.NewProductRepository(db, log)
	productService := service.NewProductService(receptionRepository, productRepository, log)
	productHandler := handler.NewProductHandler(productService, log)

	routes.SetupRoutes(
		cfg,
		r,
		userHandler,
		pvzHandler,
		receptionHandler,
		productHandler,
		middleware.AuthMiddleware(cfg, log),
		middleware.ModeratorOnly,
		middleware.EmployeeOnly,
	)

	srv := &http.Server{
		Addr:    cfg.AppHost + ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Failed to start server")
		}
	}()
	log.Info().Msgf("Starting server on %s", serverAdress)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error shutting down server")
	} else {
		log.Info().Msg("Server stopped gracefully")
	}
}
