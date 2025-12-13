package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/drone-service/config"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/middleware"
	v1 "github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/v1"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/websocket"
	repo "github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo/persistent"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/grpc"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/minio"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/postgres"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/rabbitmq"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func Run(cfg *config.Config) {
	ctx := context.Background()
	logger := logger.New(cfg.LogLevel)

	prometheusMiddleware := ginprometheus.NewPrometheus("drone-service")

	pg, err := postgres.New(&cfg.PG)
	if err != nil {
		logger.Error("app - Run - postgres.New", err)
	}
	defer pg.Close()

	rabbitmqClient, err := rabbitmq.NewClient(&cfg.RabbitMQ)
	if err != nil {
		logger.Error("app - Run - rabbitmq.NewClient", err)
	}
	defer func() {
		if err := rabbitmqClient.Close(); err != nil {
			logger.Error("failed to close rabbitmq client", err)
		}
	}()

	minioClient, err := minio.New(&cfg.MinIO)
	if err != nil {
		logger.Error("app - Run - minio.New", err)
	}

	orchestratorGRPCClient, err := grpc.NewOrchestratorGRPCClient(&cfg.OrchestratorGRPC)
	if err != nil {
		logger.Error("app - Run - grpc.NewOrchestratorGRPCClient", err)
	}

	droneRepo := repo.NewDroneRepo(pg)
	deliveryRepo := repo.NewDeliveryRepo(pg)
	droneManager := usecase.NewDroneManagerUseCase(droneRepo)

	videoHandler := websocket.NewVideoHandler(minioClient)

	tempHandler := websocket.NewDroneWebSocketHandler(nil)

	deliveryUseCase := usecase.NewDeliveryUseCase(
		droneRepo,
		deliveryRepo,
		droneManager,
		tempHandler,
		orchestratorGRPCClient,
		rabbitmqClient,
	)

	droneMessageUseCase := usecase.NewDroneMessageUseCase(
		droneRepo,
		deliveryRepo,
		droneManager,
		deliveryUseCase,
		tempHandler,
		videoHandler,
	)

	droneWSHandler := websocket.NewDroneWebSocketHandler(droneMessageUseCase)

	deliveryUseCase = usecase.NewDeliveryUseCase(
		droneRepo,
		deliveryRepo,
		droneManager,
		droneWSHandler,
		orchestratorGRPCClient,
		rabbitmqClient,
	)

	droneMessageUseCase = usecase.NewDroneMessageUseCase(
		droneRepo,
		deliveryRepo,
		droneManager,
		deliveryUseCase,
		droneWSHandler,
		videoHandler,
	)

	droneWSHandler = websocket.NewDroneWebSocketHandler(droneMessageUseCase)

	adminWSHandler := websocket.NewAdminWebSocketHandler(droneManager, droneRepo, cfg.BroadcastInterval)

	deliveryWorker := rabbitmq.NewDeliveryWorker(rabbitmqClient, deliveryUseCase)
	if err := deliveryWorker.Start(ctx); err != nil {
		logger.Error("app - Run - deliveryWorker.Start", err)
	}
	logger.Info("Delivery worker started successfully", nil)

	gin.SetMode(cfg.GinMode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	prometheusMiddleware.Use(router)

	v1.NewRouter(router, droneWSHandler, adminWSHandler, videoHandler, droneManager)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info(fmt.Sprintf("HTTP server started on port %s", cfg.Port), nil)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("app - Run - httpServer.ListenAndServe", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	logger.Info("app - Run - shutting down", nil)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("app - Run - httpServer.Shutdown", err)
	}
}
