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

// @Summary      Список дронов
// @Description  Возвращает список всех доступных дронов
// @Tags         drones
// @Accept       json
// @Produce      json
// @Success      200 {array} entity.Drone
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /drones [get]
func (r *droneRoutes) list(c *gin.Context) {
	drones, err := r.uc.ListDrones(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, drones)
}

// @Summary      Статус дрона
// @Description  Возвращает текущий статус дрона (уровень батареи, позиция и т.д.)
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path int true "ID дрона"
// @Success      200 {object} map[string]any
// @Failure      400 {object} response.Error
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
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// @Summary      Получить дрон по ID
// @Description  Возвращает информацию о дроне по его ID
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path string true "ID дрона"
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

	drone, err := r.uc.GetDroneByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, drone)
}

// @Summary      Создать дрон
// @Description  Создает новый дрон с указанной моделью
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        request body request.CreateDroneRequest true "Модель дрона"
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

	drone, err := r.uc.CreateDrone(c.Request.Context(), req.Model, req.IPAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, drone)
}

// @Summary      Обновить дрон
// @Description  Обновляет модель дрона
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path string true "ID дрона"
// @Param        request body request.UpdateDroneRequest true "Новая модель"
// @Success      200 {object} entity.Drone
// @Failure      400 {object} response.Error
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

	drone, err := r.uc.UpdateDroneInfo(c.Request.Context(), id, req.Model, req.IPAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, drone)
}

// @Summary      Обновить статус дрона
// @Description  Изменяет статус дрона (idle, busy, error)
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path string true "ID дрона"
// @Param        request body request.UpdateDroneStatusRequest true "Новый статус"
// @Success      200 {object} entity.Drone
// @Failure      400 {object} response.Error
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

	drone, err := r.uc.UpdateDrone(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, drone)
}

// @Summary      Удалить дрон
// @Description  Удаляет дрон по ID
// @Tags         drones
// @Accept       json
// @Produce      json
// @Param        id path string true "ID дрона"
// @Success      200 {object} map[string]string
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /drones/{id} [delete]
func (r *droneRoutes) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid drone ID"})
		return
	}

	if err := r.uc.DeleteDrone(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Drone deleted successfully"})
}
