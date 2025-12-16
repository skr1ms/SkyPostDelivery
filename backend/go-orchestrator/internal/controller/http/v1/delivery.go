package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/request"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/response"
	_ "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase"
)

type deliveryRoutes struct {
	uc *usecase.DeliveryUseCase
}

func newDeliveryRoutes(g *gin.RouterGroup, uc *usecase.DeliveryUseCase) {
	r := &deliveryRoutes{uc: uc}

	group := g.Group("/deliveries")
	{
		group.GET("/:id", r.get)
		group.PUT("/:id/status", r.updateStatus)
		group.POST("/confirm-loaded", r.confirmGoodsLoaded)
	}
}

// @Summary      Get delivery
// @Description  Returns delivery information by ID
// @Tags         deliveries
// @Accept       json
// @Produce      json
// @Param        id path int true "Delivery ID"
// @Success      200 {object} entity.Delivery
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Security     Bearer
// @Router       /deliveries/{id} [get]
func (r *deliveryRoutes) get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid delivery ID"})
		return
	}

	delivery, err := r.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, delivery)
}

// @Summary      Update delivery status
// @Description  Updates delivery status (pending, in_progress, completed, failed)
// @Tags         deliveries
// @Accept       json
// @Produce      json
// @Param        id path int true "Delivery ID"
// @Param        request body request.UpdateDeliveryStatus true "New status"
// @Success      200 {object} response.Success
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /deliveries/{id}/status [put]
func (r *deliveryRoutes) updateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid delivery ID"})
		return
	}

	var req request.UpdateDeliveryStatus
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	if err := r.uc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success{Success: true})
}

// @Summary      Confirm goods loaded
// @Description  Confirms goods loading from cell to drone
// @Tags         deliveries
// @Accept       json
// @Produce      json
// @Param        request body request.ConfirmGoodsLoadedRequest true "Order ID and cell ID"
// @Success      200 {object} map[string]any
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /deliveries/confirm-loaded [post]
func (r *deliveryRoutes) confirmGoodsLoaded(c *gin.Context) {
	var req request.ConfirmGoodsLoadedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "invalid order_id"})
		return
	}

	lockerCellID, err := uuid.Parse(req.LockerCellID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "invalid locker_cell_id"})
		return
	}

	if err := r.uc.ConfirmGoodsLoaded(c.Request.Context(), orderID, lockerCellID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Goods loaded confirmed successfully",
	})
}
