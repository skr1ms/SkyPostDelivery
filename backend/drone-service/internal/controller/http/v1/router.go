package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/v1/response"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/websocket"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(
	router *gin.Engine,
	droneWSHandler *websocket.DroneWebSocketHandler,
	adminWSHandler *websocket.AdminWebSocketHandler,
	videoHandler *websocket.VideoHandler,
	droneManager *usecase.DroneManagerUseCase,
) {
	droneHandler := NewDroneHandler(droneManager)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", healthCheck)
	router.GET("/status", droneHandler.GetStatus)

	api := router.Group("/v1/api")
	{
		drones := api.Group("/drones")
		{
			drones.POST("/:drone_id/command", droneHandler.SendCommand)
		}
	}

	ws := router.Group("/ws")
	{
		ws.GET("/drone", droneWSHandler.HandleDroneConnection)
		ws.GET("/drone/:drone_id", droneWSHandler.HandleDroneConnection)
		ws.GET("/drone/:drone_id/video", videoHandler.HandleAdminVideoConnection)
		ws.GET("/admin", adminWSHandler.HandleAdminConnection)
	}
}

// @Summary      Health Check
// @Description  Service health check
// @Tags         monitoring
// @Accept       json
// @Produce      json
// @Success      200 {object} response.Health
// @Router       /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, response.Health{
		Status:  "healthy",
		Service: "drone-service",
	})
}
