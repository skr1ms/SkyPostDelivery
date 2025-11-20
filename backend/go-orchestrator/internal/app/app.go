package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/config"
	grpcmiddleware "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/grpc/middleware"
	grpcserver "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/grpc/server"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/middleware"
	v1 "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1"
	repo "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/persistent"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/webapi"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/jwt"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/minio"
	pb "github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/pb"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/postgres"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/qr"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/rabbitmq"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/redis"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/validator"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"google.golang.org/grpc"
)

func Run(cfg *config.Config) {
	logger := logger.New(cfg.App.LogLevel)

	pg, err := postgres.New(&cfg.PG)
	if err != nil {
		logger.Error("app - Run - postgres.New", err)
	}
	defer pg.Close()

	rdb := redis.New(&cfg.Redis)
	defer rdb.Close()

	limiter := middleware.NewLimiter(rdb)

	prometheusMiddleware := ginprometheus.NewPrometheus("go-orchestrator")

	if err := runMigrations(pg); err != nil {
		logger.Error("app - Run - runMigrations", err)
	}

	qrGenerator := qr.NewQRGenerator(&cfg.QR)

	if err := ensureAdminExists(pg, cfg, qrGenerator); err != nil {
		logger.Error("app - Run - ensureAdminExists", err)
	}

	minioClient, err := minio.New(
		&cfg.MinIO,
	)
	if err != nil {
		logger.Error("app - Run - minio.New", err)
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

	rabbitmqClient, err := rabbitmq.NewClient(&cfg.RabbitMQ)
	if err != nil {
		logger.Error("app - Run - rabbitmq.NewClient", err)
	}
	defer rabbitmqClient.Close()

	ctx := context.Background()

	var pushSender webapi.PushSender
	if cfg.Firebase.CredentialsFile != "" {
		pushSender, err = webapi.NewFCMSender(ctx, cfg.Firebase.CredentialsFile, cfg.Firebase.ProjectID)
		if err != nil {
			logger.Error("app - Run - FCMSender.New", err)
			pushSender = webapi.NewNoopSender()
		}
	} else {
		pushSender = webapi.NewNoopSender()
	}

	notificationUC := usecase.NewNotificationUseCase(deviceRepo, pushSender)
	userUC := usecase.NewUserUseCase(userRepo, smsWebAPI, qrAdapter, jwtService, validator.New())
	goodUC := usecase.NewGoodUseCase(goodRepo)
	orderUC := usecase.NewOrderUseCase(orderRepo, goodRepo, droneRepo, deliveryRepo, parcelAutomatRepo, lockerRepo, internalLockerRepo, rabbitmqClient)
	droneUC := usecase.NewDroneUseCase(droneRepo)
	deliveryUC := usecase.NewDeliveryUseCase(deliveryRepo, orderRepo, lockerRepo, internalLockerRepo, rabbitmqClient, notificationUC)
	lockerUC := usecase.NewLockerUseCase(lockerRepo)
	parcelAutomatUC := usecase.NewParcelAutomatUseCase(parcelAutomatRepo, lockerRepo, internalLockerRepo, orderRepo, deliveryRepo, qrUC, orangePIAdapter)
	go orderUC.StartPendingOrdersWorker(ctx, 30*time.Second)
	logger.Info("Started pending orders worker (checking every 30s)", nil)

	go deliveryUC.StartConfirmationConsumer()
	logger.Info("Started delivery confirmation consumer", nil)

	gin.SetMode(cfg.GinMode)
	router := gin.New()

	if v, ok := binding.Validator.Engine().(*playgroundvalidator.Validate); ok {
		v.RegisterValidation("russian_phone", validator.ValidateRussianPhoneField)
		v.RegisterValidation("strong_password", validator.ValidateStrongPasswordField)
		v.RegisterValidation("custom_email", validator.ValidateEmailField)
	}

	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	prometheusMiddleware.Use(router)

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

	v1.NewRouter(router, userUC, goodUC, orderUC, droneUC, deliveryUC, lockerUC, parcelAutomatUC, qrUC, notificationUC, jwtMiddleware, limiter)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
	if err != nil {
		logger.Error("app - Run - net.Listen", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcmiddleware.PrometheusUnaryInterceptor()),
	)
	orchestratorServer := grpcserver.NewOrchestratorServer(deliveryUC, parcelAutomatUC)
	pb.RegisterOrchestratorServiceServer(grpcServer, orchestratorServer)

	go func() {
		logger.Info(fmt.Sprintf("gRPC server started on port %s", cfg.GRPC.Port), nil)
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Error("app - Run - grpcServer.Serve", err)
		}
	}()

	go func() {
		logger.Info(fmt.Sprintf("HTTP server started on port %s", cfg.HTTP.Port), nil)
		if err := router.Run(":" + cfg.HTTP.Port); err != nil {
			logger.Error("app - Run - router.Run", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	logger.Info("app - Run - shutting down", nil)
	grpcServer.GracefulStop()
}
