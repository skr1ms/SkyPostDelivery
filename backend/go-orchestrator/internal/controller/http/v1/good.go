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

type goodRoutes struct {
	uc *usecase.GoodUseCase
}

func newGoodRoutes(g *gin.RouterGroup, uc *usecase.GoodUseCase) {
	r := &goodRoutes{uc: uc}

	group := g.Group("/goods")
	{
		group.POST("/", r.create)
		group.GET("/", r.list)
		group.GET("/:id", r.get)
		group.PATCH("/:id", r.update)
		group.DELETE("/:id", r.delete)
	}
}

// @Summary      List goods
// @Description  Returns list of all goods
// @Tags         goods
// @Accept       json
// @Produce      json
// @Success      200 {array} entity.Good
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /goods [get]
func (r *goodRoutes) list(c *gin.Context) {
	goods, err := r.uc.List(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, goods)
}

// @Summary      Get good
// @Description  Returns good information by ID
// @Tags         goods
// @Accept       json
// @Produce      json
// @Param        id path string true "Good ID"
// @Success      200 {object} entity.Good
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Security     Bearer
// @Router       /goods/{id} [get]
func (r *goodRoutes) get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid good ID"})
		return
	}

	good, err := r.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, good)
}

// @Summary      Create goods
// @Description  Creates a good with specified parameters
// @Tags         goods
// @Accept       json
// @Produce      json
// @Param        request body request.CreateGoodRequest true "Good parameters"
// @Success      201 {object} entity.Good
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /goods [post]
func (r *goodRoutes) create(c *gin.Context) {
	var req request.CreateGoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	good, err := r.uc.CreateWithQuantity(c.Request.Context(), req.Name, req.Weight, req.Height, req.Length, req.Width, req.Quantity)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, good)
}

// @Summary      Update good
// @Description  Updates good parameters (name, weight, dimensions)
// @Tags         goods
// @Accept       json
// @Produce      json
// @Param        id path string true "Good ID"
// @Param        request body request.UpdateGoodRequest true "New parameters"
// @Success      200 {object} entity.Good
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /goods/{id} [patch]
func (r *goodRoutes) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid good ID"})
		return
	}

	var req request.UpdateGoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	good, err := r.uc.Update(c.Request.Context(), id, req.Name, req.Weight, req.Height, req.Length, req.Width)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, good)
}

// @Summary      Delete good
// @Description  Deletes good by ID
// @Tags         goods
// @Accept       json
// @Produce      json
// @Param        id path string true "Good ID"
// @Success      204
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /goods/{id} [delete]
func (r *goodRoutes) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid good ID"})
		return
	}

	if err := r.uc.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
