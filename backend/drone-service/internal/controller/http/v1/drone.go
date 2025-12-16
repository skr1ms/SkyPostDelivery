package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/v1/request"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/v1/response"
	_ "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase"
)

type DroneHandler struct {
	droneManager *usecase.DroneManagerUseCase
}

func NewDroneHandler(droneManager *usecase.DroneManagerUseCase) *DroneHandler {
	return &DroneHandler{
		droneManager: droneManager,
	}
}

// @Summary      Get drone status
// @Description  Returns current status of the first registered drone
// @Tags         drones
// @Accept       json
// @Produce      json
// @Success      200 {object} response.DroneStatus
// @Failure      500 {object} response.Error
// @Router       /status [get]
func (h *DroneHandler) GetStatus(c *gin.Context) {
	drones := h.droneManager.GetAllDrones()

	if len(drones) == 0 {
		c.JSON(http.StatusOK, response.DroneStatus{
			DroneID:      "unknown",
			Status:       "no_drones_registered",
			BatteryLevel: 0.0,
			Position: response.Position{
				Latitude:  0.0,
				Longitude: 0.0,
				Altitude:  0.0,
			},
			Speed:             0.0,
			CurrentDeliveryID: nil,
			ErrorMessage:      "No drones registered",
		})
		return
	}

	droneID := drones[0]
	state, err := h.droneManager.GetDroneState(c.Request.Context(), droneID)

	if err != nil {
		handleError(c, err)
		return
	}

	if state == nil {
		c.JSON(http.StatusOK, response.DroneStatus{
			DroneID:      droneID,
			Status:       "unknown",
			BatteryLevel: 0.0,
			Position: response.Position{
				Latitude:  0.0,
				Longitude: 0.0,
				Altitude:  0.0,
			},
			Speed:             0.0,
			CurrentDeliveryID: nil,
			ErrorMessage:      "Drone state not available",
		})
		return
	}

	c.JSON(http.StatusOK, response.DroneStatus{
		DroneID:      state.DroneID,
		Status:       string(state.Status),
		BatteryLevel: state.BatteryLevel,
		Position: response.Position{
			Latitude:  state.CurrentPosition.Latitude,
			Longitude: state.CurrentPosition.Longitude,
			Altitude:  state.CurrentPosition.Altitude,
		},
		Speed:             state.Speed,
		CurrentDeliveryID: state.CurrentDeliveryID,
		ErrorMessage: func() string {
			if state.ErrorMessage != nil {
				return *state.ErrorMessage
			}
			return ""
		}(),
	})
}

// @Summary      Send command to drone
// @Description  Sends control command to specific drone
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        drone_id path string true "Drone ID"
// @Param        request body request.SendCommand true "Command for drone"
// @Success      200 {object} response.CommandResponse
// @Failure      400 {object} response.Error
// @Router       /api/drones/{drone_id}/command [post]
func (h *DroneHandler) SendCommand(c *gin.Context) {
	droneID := c.Param("drone_id")

	var req request.SendCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.CommandResponse{
		Status:  "command_sent",
		Message: "Command received for drone " + droneID,
	})
}
