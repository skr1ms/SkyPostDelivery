package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/controller/http/middleware"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
)

func NewRouter(
	router *gin.Engine,
	cellManager *usecase.CellManagerUseCase,
	qrScanner *usecase.QRScannerUseCase,
	log logger.Interface,
) {
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS())

	router.GET("/", serviceInfo)
	router.GET("/health", healthCheck)

	api := router.Group("/api")
	{
		cells := api.Group("/cells")
		{
			newCellRoutes(cells, cellManager, log)
		}

		qr := api.Group("/qr", middleware.RequireInitialized(cellManager.IsInitialized))
		{
			newQRRoutes(qr, qrScanner, log)
		}
	}
}
