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
	"github.com/skr1ms/SkyPostDelivery/locker-agent/config"
	v1 "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/controller/http/v1"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/hardware"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/repo/inmemory"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/orchestrator"
)

func Run(cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.New(cfg.Log.Level)

	cellRepo := inmemory.NewCellMappingRepo()

	arduino, err := hardware.NewArduinoController(
		cfg.Arduino.Port,
		cfg.Arduino.Baudrate,
		cfg.Arduino.MockMode,
		log,
	)
	if err != nil {
		log.Fatal("app - Run - hardware.NewArduinoController", err)
	}
	defer func() {
		if err := arduino.Close(); err != nil {
			log.Error("Failed to close arduino", err)
		}
	}()

	display, err := hardware.NewDisplayController(
		cfg.Display.Port,
		cfg.Display.Baudrate,
		cfg.Display.MockMode,
		log,
	)
	if err != nil {
		log.Fatal("app - Run - hardware.NewDisplayController", err)
	}
	defer func() {
		if err := display.Close(); err != nil {
			log.Error("Failed to close display", err)
		}
	}()

	qrCamera, err := hardware.NewQRCamera(
		cfg.Camera.Index,
		cfg.Camera.Width,
		cfg.Camera.Height,
		cfg.Camera.FPS,
		cfg.Camera.ScanIntervalMS,
		cfg.Camera.MockMode,
		log,
	)
	if err != nil {
		log.Fatal("app - Run - hardware.NewQRCamera", err)
	}
	defer func() {
		if err := qrCamera.Close(); err != nil {
			log.Error("Failed to close QR camera", err)
		}
	}()

	orchestratorClient := orchestrator.NewClient(
		cfg.Orchestrator.URL,
		cfg.Orchestrator.Timeout,
		cfg.Orchestrator.RetryCount,
		log,
	)

	cellManager := usecase.NewCellManagerUseCase(cellRepo, arduino, display, log)

	qrScanner := usecase.NewQRScannerUseCase(
		cellManager,
		orchestratorClient,
		qrCamera,
		display,
		cellRepo,
		log,
	)

	if !cellManager.IsInitialized() {
		log.Warn("Cell mapping not initialized. Waiting for sync from orchestrator via POST /api/cells/sync", nil)
	} else {
		log.Info("Cell mapping already initialized", nil)
	}

	qrScanner.StartScanner(ctx)

	if display != nil {
		_ = display.ShowWelcome()
	}

	gin.SetMode(cfg.Service.GinMode)
	router := gin.New()

	v1.NewRouter(router, cellManager, qrScanner, log)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Info(fmt.Sprintf("HTTP server started on port %s", cfg.HTTP.Port), nil)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("app - Run - httpServer.ListenAndServe", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	log.Info("app - Run - shutting down", nil)

	cancel()

	qrScanner.StopScanner()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("app - Run - httpServer.Shutdown", err)
	}

	log.Info("app - Run - graceful shutdown completed", nil)
}
