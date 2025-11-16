package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/skr1ms/SkyPostDelivery/drone-service/config"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/middleware"
	v1 "github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/v1"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/websocket"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase/repo"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/minio"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/postgres"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/rabbitmq"
)

func Run(cfg *config.Config) {
	ctx := context.Background()

	pg, err := postgres.New(cfg.PG.URL)
	if err != nil {
		log.Fatalf("app - Run - postgres.New: %v", err)
	}
	defer pg.Close()

	rabbitmqClient, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("app - Run - rabbitmq.NewClient: %v", err)
	}
	defer rabbitmqClient.Close()

	minioClient, err := minio.New(cfg.MinIO)
	if err != nil {
		log.Fatalf("app - Run - minio.New: %v", err)
	}

	stateRepo := repo.NewPostgresStateRepo(pg)
	droneManager := usecase.NewDroneManagerUseCase(stateRepo)

	videoHandler := websocket.NewVideoHandler(minioClient)
	droneWSHandler := websocket.NewDroneWebSocketHandler(stateRepo, stateRepo, droneManager, videoHandler)
	adminWSHandler := websocket.NewAdminWebSocketHandler(droneManager, stateRepo, cfg.WebSocket.BroadcastInterval)

	deliveryUseCase := usecase.NewDeliveryUseCase(
		stateRepo,
		stateRepo,
		droneManager,
		droneWSHandler,
		nil,
		rabbitmqClient,
	)

	droneWSHandler.SetDeliveryUseCase(deliveryUseCase)

	deliveryWorker := rabbitmq.NewDeliveryWorker(rabbitmqClient, deliveryUseCase)
	if err := deliveryWorker.Start(ctx); err != nil {
		log.Fatalf("app - Run - deliveryWorker.Start: %v", err)
	}
	log.Println("Delivery worker started successfully")

	gin.SetMode(gin.DebugMode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.PrometheusMiddleware())
	router.Use(middleware.CORS())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1.NewRouter(router, droneWSHandler, adminWSHandler, videoHandler, droneManager)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("HTTP server started on port %s", cfg.HTTP.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("app - Run - httpServer.ListenAndServe: %v", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	log.Println("app - Run - shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("app - Run - httpServer.Shutdown: %v", err)
	}
}
