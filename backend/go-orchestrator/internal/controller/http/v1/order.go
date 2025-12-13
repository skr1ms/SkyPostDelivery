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

type orderRoutes struct {
	uc *usecase.OrderUseCase
}

func newOrderRoutes(g *gin.RouterGroup, uc *usecase.OrderUseCase, orderRateLimiter gin.HandlerFunc) {
	r := &orderRoutes{uc: uc}

	group := g.Group("/orders")
	{
		group.POST("/", orderRateLimiter, r.create)
		group.POST("/batch", orderRateLimiter, r.createMultiple)
		group.POST("/:id/return", r.returnOrder)
		group.GET("/:id", r.get)
		group.GET("/user/:userId", r.getUserOrders)
	}
}

// @Summary      Создать заказ
// @Description  Создает новый заказ на доставку товара (user_id извлекается из JWT токена)
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request body request.CreateOrder true "Данные заказа"
// @Success      201 {object} entity.Order
// @Failure      400 {object} response.Error
// @Failure      401 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /orders [post]
func (r *orderRoutes) create(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid user ID"})
		return
	}

	var req request.CreateOrder
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	order, err := r.uc.CreateOrder(c.Request.Context(), userID, req.GoodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// @Summary      Создать несколько заказов
// @Description  Создает несколько заказов для разных товаров (user_id извлекается из JWT токена)
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request body request.CreateMultipleOrders true "Данные для создания нескольких заказов"
// @Success      201 {array} entity.Order
// @Failure      400 {object} response.Error
// @Failure      401 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /orders/batch [post]
func (r *orderRoutes) createMultiple(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid user ID"})
		return
	}

	var req request.CreateMultipleOrders
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	orders, err := r.uc.CreateMultipleOrders(c.Request.Context(), userID, req.GoodIDs)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, orders)
}

// @Summary      Получить заказ
// @Description  Возвращает информацию о заказе по ID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id path int true "ID заказа"
// @Success      200 {object} entity.Order
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Security     Bearer
// @Router       /orders/{id} [get]
func (r *orderRoutes) get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid order ID"})
		return
	}

	order, err := r.uc.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// @Summary      Отменить заказ
// @Description  Отменяет заказ, освобождает ячейку и возвращает дрон на базу (метка 131)
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id path string true "ID заказа (UUID)"
// @Success      200 {object} response.Success
// @Failure      400 {object} response.Error
// @Failure      401 {object} response.Error
// @Failure      403 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /orders/{id}/return [post]
func (r *orderRoutes) returnOrder(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid user ID"})
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid order ID"})
		return
	}

	if err := r.uc.ReturnOrder(c.Request.Context(), orderID, userID); err != nil {
		if err.Error() == "order usecase - ReturnOrder: order does not belong to user" {
			c.JSON(http.StatusForbidden, response.Error{Error: "You don't have permission to return this order"})
			return
		}
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Success{Success: true})
}

// @Summary      Получить заказы пользователя
// @Description  Возвращает список всех заказов пользователя
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        userId path int true "ID пользователя"
// @Success      200 {array} entity.Order
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /orders/user/{userId} [get]
func (r *orderRoutes) getUserOrders(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid user ID"})
		return
	}

	result, err := r.uc.GetUserOrdersWithGoods(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	ordersWithGoods := make([]response.OrderWithGood, 0, len(result))
	for _, item := range result {
		ordersWithGoods = append(ordersWithGoods, response.OrderWithGood{
			ID:              item.Order.ID,
			UserID:          item.Order.UserID,
			GoodID:          item.Order.GoodID,
			ParcelAutomatID: item.Order.ParcelAutomatID,
			Status:          item.Order.Status,
			CreatedAt:       item.Order.CreatedAt,
			Good:            item.Good,
		})
	}

	c.JSON(http.StatusOK, ordersWithGoods)
}
