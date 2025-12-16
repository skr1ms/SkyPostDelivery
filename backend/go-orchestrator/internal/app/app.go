package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/migrator"
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
	logger := logger.New(cfg.LogLevel)

	pg, err := postgres.New(&cfg.PG)
	if err != nil {
		logger.Error("app - Run - postgres.New", err, nil)
		return
	}
	defer pg.Close()

	rdb := redis.New(&cfg.Redis)
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.Error("failed to close redis client", err, nil)
		}
	}()

	limiter := middleware.NewLimiter(rdb, logger)

	prometheusMiddleware := ginprometheus.NewPrometheus("go-orchestrator")

	if err := migrator.Run(&cfg.PG); err != nil {
		logger.Error("app - Run - migrator.Run", err, nil)
	}

	qrGenerator := qr.NewQRGenerator(&cfg.QR)

	if err := ensureAdminExists(pg, cfg, qrGenerator, logger); err != nil {
		logger.Error("app - Run - ensureAdminExists", err, nil)
	}

	minioClient, err := minio.New(&cfg.MinIO, logger)
	if err != nil {
		logger.Error("app - Run - minio.New", err, nil)
	}

	jwtService := jwt.NewJWTService(
		cfg.AccessSecret,
		cfg.RefreshSecret,
		cfg.AccessTTL,
		cfg.RefreshTTL,
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
	qrUC := usecase.NewQRUseCase(qrGenerator, userRepo, minioClient, logger)
	orangePIAdapter := webapi.NewOrangePIAdapter()

	smsWebAPI := webapi.NewSMSAeroAPI(cfg.SMSAero.Email, cfg.APIKey, cfg.BaseURL)

	rabbitmqClient, err := rabbitmq.NewClient(&cfg.RabbitMQ, logger)
	if err != nil {
		logger.Error("app - Run - rabbitmq.NewClient", err, nil)
		return
	}
	defer func() {
		if err := rabbitmqClient.Close(); err != nil {
			logger.Error("failed to close rabbitmq client", err, nil)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var pushSender webapi.PushSender
	if cfg.CredentialsFile != "" {
		pushSender, err = webapi.NewFCMSender(ctx, cfg.CredentialsFile, cfg.ProjectID)
		if err != nil {
			logger.Error("app - Run - FCMSender.New", err, nil)
			pushSender = webapi.NewNoopSender()
		}
	} else {
		pushSender = webapi.NewNoopSender()
	}

	notificationUC := usecase.NewNotificationUseCase(deviceRepo, pushSender, logger)
	userUC := usecase.NewUserUseCase(userRepo, smsWebAPI, qrAdapter, jwtService, validator.New(), logger)
	goodUC := usecase.NewGoodUseCase(goodRepo, logger)
	orderUC := usecase.NewOrderUseCase(orderRepo, goodRepo, droneRepo, deliveryRepo, parcelAutomatRepo, lockerRepo, internalLockerRepo, rabbitmqClient, logger)
	droneUC := usecase.NewDroneUseCase(droneRepo, logger)
	deliveryUC := usecase.NewDeliveryUseCase(deliveryRepo, orderRepo, lockerRepo, internalLockerRepo, rabbitmqClient, notificationUC, logger)
	lockerUC := usecase.NewLockerUseCase(lockerRepo, logger)
	parcelAutomatUC := usecase.NewParcelAutomatUseCase(parcelAutomatRepo, lockerRepo, internalLockerRepo, orderRepo, deliveryRepo, qrUC, orangePIAdapter, logger)

	go orderUC.StartPendingOrdersWorker(ctx, 30*time.Second)
	logger.Info("Started pending orders worker (checking every 30s)", nil, nil)

	go deliveryUC.StartConfirmationConsumer(ctx)
	logger.Info("Started delivery confirmation consumer", nil, nil)

	gin.SetMode(cfg.GinMode)
	router := gin.New()

	if v, ok := binding.Validator.Engine().(*playgroundvalidator.Validate); ok {
		_ = v.RegisterValidation("russian_phone", validator.ValidateRussianPhoneField)
		_ = v.RegisterValidation("strong_password", validator.ValidateStrongPasswordField)
		_ = v.RegisterValidation("custom_email", validator.ValidateEmailField)
	}

	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	prometheusMiddleware.Use(router)

	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)

	v1.NewRouter(router, userUC, goodUC, orderUC, droneUC, deliveryUC, lockerUC, parcelAutomatUC, qrUC, notificationUC, jwtMiddleware, limiter)

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: router,
	}

	go func() {
		logger.Info(fmt.Sprintf("HTTP server started on port %s", cfg.HTTP.Port), nil, nil)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("app - Run - httpServer.ListenAndServe", err, nil)
		}
	}()

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
	if err != nil {
		logger.Error("app - Run - net.Listen", err, nil)
		return
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcmiddleware.PrometheusUnaryInterceptor()),
	)
	orchestratorServer := grpcserver.NewOrchestratorServer(deliveryUC, parcelAutomatUC)
	pb.RegisterOrchestratorServiceServer(grpcServer, orchestratorServer)

	go func() {
		logger.Info(fmt.Sprintf("gRPC server started on port %s", cfg.GRPC.Port), nil, nil)
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Error("app - Run - grpcServer.Serve", err, nil)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	logger.Info("app - Run - shutting down", nil, nil)

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("app - Run - httpServer.Shutdown", err, nil)
	} else {
		logger.Info("HTTP server stopped gracefully", nil, nil)
	}

	grpcServer.GracefulStop()
	logger.Info("gRPC server stopped gracefully", nil, nil)
}
