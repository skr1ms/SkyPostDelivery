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

// @Summary      Список товаров
// @Description  Возвращает список всех товаров
// @Tags         goods
// @Accept       json
// @Produce      json
// @Success      200 {array} entity.Good
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /goods [get]
func (r *goodRoutes) list(c *gin.Context) {
	goods, err := r.uc.ListGoods(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, goods)
}

// @Summary      Получить товар
// @Description  Возвращает информацию о товаре по ID
// @Tags         goods
// @Accept       json
// @Produce      json
// @Param        id path string true "ID товара"
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

	good, err := r.uc.GetGood(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, good)
}

// @Summary      Создать товары
// @Description  Создает группу товаров с указанными параметрами
// @Tags         goods
// @Accept       json
// @Produce      json
// @Param        request body request.CreateGoodRequest true "Параметры товаров"
// @Success      201 {array} entity.Good
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

	goods, err := r.uc.CreateGoods(c.Request.Context(), req.Name, req.Weight, req.Height, req.Length, req.Width, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, goods)
}

// @Summary      Обновить товар
// @Description  Обновляет параметры товара (название, вес, габариты)
// @Tags         goods
// @Accept       json
// @Produce      json
// @Param        id path string true "ID товара"
// @Param        request body request.UpdateGoodRequest true "Новые параметры"
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

	good, err := r.uc.UpdateGood(c.Request.Context(), id, req.Name, req.Weight, req.Height, req.Length, req.Width)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, good)
}

// @Summary      Удалить товар
// @Description  Удаляет товар по ID
// @Tags         goods
// @Accept       json
// @Produce      json
// @Param        id path string true "ID товара"
// @Success      200 {object} map[string]string
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /goods/{id} [delete]
func (r *goodRoutes) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid good ID"})
		return
	}

	if err := r.uc.DeleteGood(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Good deleted successfully"})
}
