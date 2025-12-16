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

type droneRoutes struct {
	uc *usecase.DroneUseCase
}

func newDroneRoutes(g *gin.RouterGroup, uc *usecase.DroneUseCase) {
	r := &droneRoutes{uc: uc}

	group := g.Group("/drones")
	{
		group.POST("/", r.create)
		group.GET("/", r.list)
		group.GET("/:id", r.getDroneByID)
		group.GET("/:id/status", r.getStatus)
		group.PUT("/:id", r.update)
		group.PATCH("/:id/status", r.updateStatus)
		group.DELETE("/:id", r.delete)
	}
}

// @Summary      List drones
// @Description  Returns list of all available drones
// @Tags         drones
// @Accept       json
// @Produce      json
// @Success      200 {array} entity.Drone
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /drones [get]
func (r *droneRoutes) list(c *gin.Context) {
	drones, err := r.uc.List(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, drones)
}

// @Summary      Drone status
// @Description  Returns current drone status (battery level, position, etc.)
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path string true "Drone ID"
// @Success      200 {object} map[string]any
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /drones/{id}/status [get]
func (r *droneRoutes) getStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid drone ID"})
		return
	}

	status, err := r.uc.GetStatus(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}

// @Summary      Get drone by ID
// @Description  Returns drone information by ID
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path string true "Drone ID"
// @Success      200 {object} entity.Drone
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /drones/{id} [get]
func (r *droneRoutes) getDroneByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid drone ID"})
		return
	}

	drone, err := r.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, drone)
}

// @Summary      Create drone
// @Description  Creates a new drone with specified model
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        request body request.CreateDroneRequest true "Drone model"
// @Success      201 {object} entity.Drone
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /drones [post]
func (r *droneRoutes) create(c *gin.Context) {
	var req request.CreateDroneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	drone, err := r.uc.Create(c.Request.Context(), req.Model, req.IPAddress)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, drone)
}

// @Summary      Update drone
// @Description  Updates drone model
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path string true "Drone ID"
// @Param        request body request.UpdateDroneRequest true "New model"
// @Success      200 {object} entity.Drone
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /drones/{id} [put]
func (r *droneRoutes) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid drone ID"})
		return
	}

	var req request.UpdateDroneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	drone, err := r.uc.Update(c.Request.Context(), id, req.Model, req.IPAddress, "")
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, drone)
}

// @Summary      Update drone status
// @Description  Changes drone status (idle, busy, error)
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path string true "Drone ID"
// @Param        request body request.UpdateDroneStatusRequest true "New status"
// @Success      204
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /drones/{id}/status [patch]
func (r *droneRoutes) updateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid drone ID"})
		return
	}

	var req request.UpdateDroneStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	err = r.uc.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Delete drone
// @Description  Deletes drone by ID
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path string true "Drone ID"
// @Success      204
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Failure      409 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /drones/{id} [delete]
func (r *droneRoutes) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid drone ID"})
		return
	}

	if err := r.uc.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
