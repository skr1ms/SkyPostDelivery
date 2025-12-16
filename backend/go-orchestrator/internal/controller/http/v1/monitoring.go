package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/response"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase"
)

type monitoringRoutes struct {
	droneUC         *usecase.DroneUseCase
	parcelAutomatUC *usecase.ParcelAutomatUseCase
	deliveryUC      *usecase.DeliveryUseCase
	orderUC         *usecase.OrderUseCase
}

func newMonitoringRoutes(g *gin.RouterGroup, droneUC *usecase.DroneUseCase, parcelAutomatUC *usecase.ParcelAutomatUseCase, deliveryUC *usecase.DeliveryUseCase, orderUC *usecase.OrderUseCase) {
	r := &monitoringRoutes{
		droneUC:         droneUC,
		parcelAutomatUC: parcelAutomatUC,
		deliveryUC:      deliveryUC,
		orderUC:         orderUC,
	}

	group := g.Group("/monitoring")
	{
		group.GET("/system-status", r.getSystemStatus)
	}
}

// @Summary      System status
// @Description  Returns current system state: list of drones, parcel automats and active deliveries
// @Tags         monitoring
// @Accept       json
// @Produce      json
// @Success      200 {object} response.SystemStatus
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /monitoring/system-status [get]
func (r *monitoringRoutes) getSystemStatus(c *gin.Context) {
	ctx := c.Request.Context()

	drones, err := r.droneUC.List(ctx)
	if err != nil {
		handleError(c, err)
		return
	}

	automats, err := r.parcelAutomatUC.List(ctx)
	if err != nil {
		handleError(c, err)
		return
	}

	automatsWithCells := make([]response.AutomatWithCells, 0, len(automats))
	for _, automat := range automats {
		cells, err := r.parcelAutomatUC.GetAutomatCells(ctx, automat.ID)
		if err != nil {
			continue
		}
		automatsWithCells = append(automatsWithCells, response.AutomatWithCells{
			ParcelAutomat: automat,
			Cells:         cells,
		})
	}

	activeDeliveries, err := r.deliveryUC.ListByStatus(ctx, "in_transit")
	if err != nil {
		activeDeliveries = []*entity.Delivery{}
	}

	pendingDeliveries, err := r.deliveryUC.ListByStatus(ctx, "pending")
	if err != nil {
		pendingDeliveries = []*entity.Delivery{}
	}

	allActiveDeliveries := append(activeDeliveries, pendingDeliveries...)

	c.JSON(http.StatusOK, response.SystemStatus{
		Drones:           drones,
		Automats:         automatsWithCells,
		ActiveDeliveries: allActiveDeliveries,
	})
}
