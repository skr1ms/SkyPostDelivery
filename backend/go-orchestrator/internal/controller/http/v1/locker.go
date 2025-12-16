package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/response"
	_ "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase"
)

type lockerRoutes struct {
	uc *usecase.LockerUseCase
}

func newLockerRoutes(g *gin.RouterGroup, uc *usecase.LockerUseCase) {
	r := &lockerRoutes{uc: uc}

	group := g.Group("/locker")
	{
		group.GET("/cells/:id", r.getCell)
	}
}

// @Summary      Get parcel automat cell
// @Description  Returns parcel automat cell information by ID
// @Tags         locker
// @Accept       json
// @Produce      json
// @Param        id path int true "Cell ID"
// @Success      200 {object} entity.LockerCell
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Security     Bearer
// @Router       /locker/cells/{id} [get]
func (r *lockerRoutes) getCell(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid cell ID"})
		return
	}

	cell, err := r.uc.GetCell(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, cell)
}
