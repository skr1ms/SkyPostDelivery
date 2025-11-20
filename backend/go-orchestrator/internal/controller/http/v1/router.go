package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/middleware"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(
	router *gin.Engine,
	userUC *usecase.UserUseCase,
	goodUC *usecase.GoodUseCase,
	orderUC *usecase.OrderUseCase,
	droneUC *usecase.DroneUseCase,
	deliveryUC *usecase.DeliveryUseCase,
	lockerUC *usecase.LockerUseCase,
	parcelAutomatUC *usecase.ParcelAutomatUseCase,
	qrUC *usecase.QRUseCase,
	notificationUC *usecase.NotificationUseCase,
	jwtMiddleware *middleware.JWTMiddleware,
	limiter *middleware.Limiter,
) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		protected := v1.Group("")
		protected.Use(jwtMiddleware.RequireAuth())

		newUserRoutes(v1, userUC, jwtMiddleware.JWTService, notificationUC, protected, limiter.MiddleWare(middleware.UserPeriod, middleware.UserRateLimit))
		newQRRoutes(v1, qrUC, jwtMiddleware, limiter.MiddleWare(middleware.QrPeriod, middleware.QrRateLimit))
		newLockerRoutes(v1, lockerUC)
		newGoodRoutes(protected, goodUC)
		newOrderRoutes(protected, orderUC, limiter.MiddleWare(middleware.OrderPeriod, middleware.OrderRateLimit))
		newDeliveryRoutes(protected, deliveryUC)
		newDroneRoutes(protected, droneUC)
		newParcelAutomatRoutes(v1, protected, parcelAutomatUC)
		newMonitoringRoutes(protected, droneUC, parcelAutomatUC, deliveryUC, orderUC)
	}
}
