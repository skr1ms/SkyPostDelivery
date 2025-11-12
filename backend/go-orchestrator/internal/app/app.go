package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/skr1ms/hitech-ekb/config"
	grpcmiddleware "github.com/skr1ms/hitech-ekb/internal/controller/grpc/middleware"
	grpcserver "github.com/skr1ms/hitech-ekb/internal/controller/grpc/server"
	"github.com/skr1ms/hitech-ekb/internal/controller/http/middleware"
	v1 "github.com/skr1ms/hitech-ekb/internal/controller/http/v1"
	"github.com/skr1ms/hitech-ekb/internal/usecase"
	"github.com/skr1ms/hitech-ekb/internal/usecase/repo"
	"github.com/skr1ms/hitech-ekb/internal/usecase/webapi"
	"github.com/skr1ms/hitech-ekb/pkg/jwt"
	"github.com/skr1ms/hitech-ekb/pkg/minio"
	pb "github.com/skr1ms/hitech-ekb/pkg/pb"
	"github.com/skr1ms/hitech-ekb/pkg/postgres"
	"github.com/skr1ms/hitech-ekb/pkg/push"
	"github.com/skr1ms/hitech-ekb/pkg/qr"
	"github.com/skr1ms/hitech-ekb/pkg/rabbitmq"
	"google.golang.org/grpc"
)

func Run(cfg *config.Config) {
	pg, err := postgres.New(cfg.PG.URL)
	if err != nil {
		log.Fatalf("app - Run - postgres.New: %v", err)
	}
	defer pg.Close()

	if err := runMigrations(pg); err != nil {
		log.Fatalf("app - Run - runMigrations: %v", err)
	}

	qrGenerator := qr.NewQRGenerator(cfg.QR.HMACSecret)

	if err := ensureAdminExists(pg, cfg, qrGenerator); err != nil {
		log.Fatalf("app - Run - ensureAdminExists: %v", err)
	}

	minioClient, err := minio.New(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.PublicURL,
		cfg.MinIO.UseSSL,
		cfg.MinIO.BucketQR,
		cfg.MinIO.BucketRecords,
	)
	if err != nil {
		log.Fatalf("app - Run - minio.New: %v", err)
	}

	jwtService := jwt.NewJWTService(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)

	userRepo := repo.NewUserRepo(pg)
	goodRepo := repo.NewGoodRepo(pg)
	orderRepo := repo.NewOrderRepo(pg)
	droneRepo := repo.NewDroneRepo(pg)
	lockerRepo := repo.NewLockerRepo(pg)
	internalLockerRepo := repo.NewInternalLockerRepo(pg)
	deliveryRepo := repo.NewDeliveryRepo(pg)
	parcelAutomatRepo := repo.NewParcelAutomatRepo(pg)
	deviceRepo := repo.NewDeviceRepo(pg)

	qrAdapter := webapi.NewQRAdapter(qrGenerator)
	qrUC := usecase.NewQRUseCase(qrGenerator, userRepo, minioClient)
	orangePIAdapter := webapi.NewOrangePIAdapter()

	smsWebAPI := webapi.NewSMSAeroAPI(cfg.SMSAero.Email, cfg.SMSAero.APIKey, cfg.SMSAero.BaseURL)

	rabbitmqClient, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	log.Println("Connected to RabbitMQ successfully")

	ctx := context.Background()

	var pushSender push.Sender
	if cfg.Firebase.CredentialsFile != "" {
		pushSender, err = push.NewFCMSender(ctx, cfg.Firebase.CredentialsFile)
		if err != nil {
			log.Printf("app - Run - push.NewFCMSender: %v", err)
			pushSender = push.NewNoopSender()
		}
	} else {
		pushSender = push.NewNoopSender()
	}

	notificationUC := usecase.NewNotificationUseCase(deviceRepo, pushSender)
	userUC := usecase.NewUserUseCase(userRepo, smsWebAPI, qrAdapter, jwtService)
	goodUC := usecase.NewGoodUseCase(goodRepo)
	orderUC := usecase.NewOrderUseCase(orderRepo, goodRepo, droneRepo, deliveryRepo, parcelAutomatRepo, lockerRepo, internalLockerRepo, rabbitmqClient)
	droneUC := usecase.NewDroneUseCase(droneRepo)
	deliveryUC := usecase.NewDeliveryUseCase(deliveryRepo, orderRepo, lockerRepo, internalLockerRepo, rabbitmqClient, notificationUC)
	lockerUC := usecase.NewLockerUseCase(lockerRepo)
	parcelAutomatUC := usecase.NewParcelAutomatUseCase(parcelAutomatRepo, lockerRepo, internalLockerRepo, orderRepo, deliveryRepo, qrUC, orangePIAdapter)
	go orderUC.StartPendingOrdersWorker(ctx, 30*time.Second)
	log.Println("Started pending orders worker (checking every 30s)")

	go deliveryUC.StartConfirmationConsumer()
	log.Println("Started delivery confirmation consumer")

	gin.SetMode(gin.DebugMode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.PrometheusMiddleware())

	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)

	allowedOrigins := []string{
		"http://localhost",
		"http://localhost:80",
		"http://localhost:3000",
		"http://localhost:5173",
	}
	if cfg.AdminPanelURL.URL != "" {
		allowedOrigins = append(allowedOrigins, cfg.AdminPanelURL.URL)
		if cfg.AdminPanelURL.URL[:5] == "https" {
			allowedOrigins = append(allowedOrigins, "http"+cfg.AdminPanelURL.URL[5:])
		} else {
			allowedOrigins = append(allowedOrigins, "https"+cfg.AdminPanelURL.URL[4:])
		}
	}

	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}
	router.Use(cors.New(corsConfig))

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1.NewRouter(router, userUC, goodUC, orderUC, droneUC, deliveryUC, lockerUC, parcelAutomatUC, qrUC, notificationUC, jwtMiddleware)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("app - Run - net.Listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcmiddleware.PrometheusUnaryInterceptor()),
	)
	orchestratorServer := grpcserver.NewOrchestratorServer(deliveryUC, parcelAutomatUC)
	pb.RegisterOrchestratorServiceServer(grpcServer, orchestratorServer)

	go func() {
		log.Printf("gRPC server started on port %s", cfg.GRPC.Port)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("app - Run - grpcServer.Serve: %v", err)
		}
	}()

	go func() {
		log.Printf("HTTP server started on port %s", cfg.HTTP.Port)
		if err := router.Run(":" + cfg.HTTP.Port); err != nil {
			log.Fatalf("app - Run - router.Run: %v", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	log.Println("app - Run - shutting down")
	grpcServer.GracefulStop()
}
