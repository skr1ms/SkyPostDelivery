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
		logger.Error("app - Run - postgres.New", err, nil)
	}
	defer pg.Close()

	rabbitmqClient, err := rabbitmq.NewClient(&cfg.RabbitMQ, logger)
	if err != nil {
		logger.Error("app - Run - rabbitmq.NewClient", err, nil)
	}
	defer func() {
		if err := rabbitmqClient.Close(); err != nil {
			logger.Error("failed to close rabbitmq client", err, nil)
		}
	}()

	minioClient, err := minio.New(&cfg.MinIO)
	if err != nil {
		logger.Error("app - Run - minio.New", err, nil)
	}

	orchestratorGRPCClient, err := grpc.NewOrchestratorGRPCClient(&cfg.OrchestratorGRPC, logger)
	if err != nil {
		logger.Error("app - Run - grpc.NewOrchestratorGRPCClient", err, nil)
	}
	defer func() {
		if orchestratorGRPCClient != nil {
			if err := orchestratorGRPCClient.Close(); err != nil {
				logger.Error("failed to close orchestrator gRPC client", err, nil)
			}
		}
	}()

	droneRepo := repo.NewDroneRepo(pg)
	deliveryRepo := repo.NewDeliveryRepo(pg)
	droneManager := usecase.NewDroneManagerUseCase(droneRepo, logger)

	videoHandler := websocket.NewVideoHandler(minioClient, logger)

	droneConnectionUseCase := usecase.NewDroneConnectionUseCase(droneRepo, droneManager, logger)
	droneTelemetryUseCase := usecase.NewDroneTelemetryUseCase(droneRepo, logger)
	droneCommandUseCase := usecase.NewDroneCommandUseCase(nil, logger)

	tempHandler := websocket.NewDroneWebSocketHandler(
		droneConnectionUseCase,
		droneTelemetryUseCase,
		nil,
		droneCommandUseCase,
		logger,
	)

	deliveryUseCase := usecase.NewDeliveryUseCase(
		droneRepo,
		deliveryRepo,
		droneManager,
		tempHandler,
		orchestratorGRPCClient,
		rabbitmqClient,
		logger,
	)

	droneDeliveryUseCase := usecase.NewDroneDeliveryUseCase(
		deliveryRepo,
		droneManager,
		deliveryUseCase,
		videoHandler,
		logger,
	)

	droneWSHandler := websocket.NewDroneWebSocketHandler(
		droneConnectionUseCase,
		droneTelemetryUseCase,
		droneDeliveryUseCase,
		droneCommandUseCase,
		logger,
	)

	droneCommandUseCase = usecase.NewDroneCommandUseCase(droneWSHandler, logger)

	deliveryUseCase = usecase.NewDeliveryUseCase(
		droneRepo,
		deliveryRepo,
		droneManager,
		droneWSHandler,
		orchestratorGRPCClient,
		rabbitmqClient,
		logger,
	)

	droneDeliveryUseCase = usecase.NewDroneDeliveryUseCase(
		deliveryRepo,
		droneManager,
		deliveryUseCase,
		videoHandler,
		logger,
	)

	droneWSHandler = websocket.NewDroneWebSocketHandler(
		droneConnectionUseCase,
		droneTelemetryUseCase,
		droneDeliveryUseCase,
		droneCommandUseCase,
		logger,
	)

	adminWSHandler := websocket.NewAdminWebSocketHandler(droneManager, droneRepo, cfg.BroadcastInterval, logger)

	deliveryWorker := rabbitmq.NewDeliveryWorker(rabbitmqClient, deliveryUseCase, logger)
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
