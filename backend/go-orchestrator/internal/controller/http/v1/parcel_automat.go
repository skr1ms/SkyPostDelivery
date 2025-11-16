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

type parcelAutomatRoutes struct {
	uc *usecase.ParcelAutomatUseCase
}

func newParcelAutomatRoutes(public *gin.RouterGroup, protected *gin.RouterGroup, uc *usecase.ParcelAutomatUseCase) {
	r := &parcelAutomatRoutes{uc: uc}

	// Публичные роуты для постамата (без авторизации)
	publicGroup := public.Group("/automats")
	{
		publicGroup.POST("/qr-scan", r.processQRScan)
		publicGroup.POST("/confirm-pickup", r.confirmPickup)
	}

	// Защищенные роуты для админки (с авторизацией)
	protectedGroup := protected.Group("/automats")
	{
		protectedGroup.POST("/", r.create)
		protectedGroup.GET("/", r.list)
		protectedGroup.GET("/working", r.listWorking)
		protectedGroup.GET("/:id", r.getByID)
		protectedGroup.PUT("/:id", r.update)
		protectedGroup.GET("/:id/cells", r.getCells)
		protectedGroup.PATCH("/:id/cells/:cellId", r.updateCell)
		protectedGroup.PATCH("/:id/status", r.updateStatus)
		protectedGroup.DELETE("/:id", r.delete)
	}
}

// @Summary      Создать постамат
// @Description  Создает новый постамат с указанными параметрами
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        request body request.CreateParcelAutomatRequest true "Данные постамата"
// @Success      201 {object} entity.ParcelAutomat
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /automats [post]
func (r *parcelAutomatRoutes) create(c *gin.Context) {
	var req request.CreateParcelAutomatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	if len(req.Cells) != req.NumberOfCells {
		c.JSON(http.StatusBadRequest, response.Error{Error: "number of cells dimensions must match number_of_cells"})
		return
	}

	automat, err := r.uc.Create(c.Request.Context(), req.City, req.Address, req.IPAddress, req.Coordinates, req.NumberOfCells, req.ArucoID, req.Cells)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, automat)
}

// @Summary      Список постаматов
// @Description  Возвращает список всех постаматов
// @Tags         automats
// @Accept       json
// @Produce      json
// @Success      200 {array} entity.ParcelAutomat
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /automats [get]
func (r *parcelAutomatRoutes) list(c *gin.Context) {
	automats, err := r.uc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, automats)
}

// @Summary      Список работающих постаматов
// @Description  Возвращает список постаматов со статусом "работает"
// @Tags         automats
// @Accept       json
// @Produce      json
// @Success      200 {array} entity.ParcelAutomat
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /automats/working [get]
func (r *parcelAutomatRoutes) listWorking(c *gin.Context) {
	automats, err := r.uc.ListWorking(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, automats)
}

// @Summary      Получить постамат
// @Description  Возвращает информацию о постамате по ID
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "ID постамата"
// @Success      200 {object} entity.ParcelAutomat
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Security     Bearer
// @Router       /automats/{id} [get]
func (r *parcelAutomatRoutes) getByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid automat ID"})
		return
	}

	automat, err := r.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, automat)
}

// @Summary      Обновить постамат
// @Description  Обновляет информацию о постамате (город, адрес, IP, координаты)
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "ID постамата"
// @Param        request body request.UpdateParcelAutomatRequest true "Данные для обновления"
// @Success      200 {object} entity.ParcelAutomat
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /automats/{id} [put]
func (r *parcelAutomatRoutes) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid automat ID"})
		return
	}

	var req request.UpdateParcelAutomatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	automat, err := r.uc.Update(c.Request.Context(), id, req.City, req.Address, req.IPAddress, req.Coordinates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, automat)
}

// @Summary      Получить ячейки постамата
// @Description  Возвращает список всех ячеек указанного постамата
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "ID постамата"
// @Success      200 {array} entity.LockerCell
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /automats/{id}/cells [get]
func (r *parcelAutomatRoutes) getCells(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid automat ID"})
		return
	}

	cells, err := r.uc.GetAutomatCells(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, cells)
}

// @Summary      Обновить ячейку
// @Description  Обновляет размеры ячейки постамата
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "ID постамата"
// @Param        cellId path string true "ID ячейки"
// @Param        request body request.UpdateCellRequest true "Новые размеры ячейки"
// @Success      200 {object} entity.LockerCell
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /automats/{id}/cells/{cellId} [patch]
func (r *parcelAutomatRoutes) updateCell(c *gin.Context) {
	_, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid automat ID"})
		return
	}

	cellID, err := uuid.Parse(c.Param("cellId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid cell ID"})
		return
	}

	var req request.UpdateCellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	cell, err := r.uc.UpdateCell(c.Request.Context(), cellID, req.Height, req.Length, req.Width)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, cell)
}

// @Summary      Обновить статус постамата
// @Description  Включает или выключает постамат
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "ID постамата"
// @Param        request body request.UpdateParcelAutomatStatusRequest true "Новый статус"
// @Success      200 {object} entity.ParcelAutomat
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /automats/{id}/status [patch]
func (r *parcelAutomatRoutes) updateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid automat ID"})
		return
	}

	var req request.UpdateParcelAutomatStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	automat, err := r.uc.UpdateStatus(c.Request.Context(), id, req.IsWorking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, automat)
}

// @Summary      Удалить постамат
// @Description  Удаляет постамат по ID
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "ID постамата"
// @Success      200 {object} map[string]string
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /automats/{id} [delete]
func (r *parcelAutomatRoutes) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid automat ID"})
		return
	}

	if err := r.uc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Parcel automat deleted successfully"})
}

// @Summary      Сканирование QR-кода
// @Description  Обрабатывает QR-код пользователя и возвращает ID ячеек с его заказами
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        request body request.QRScanRequest true "QR данные"
// @Success      200 {object} map[string]any
// @Failure      400 {object} response.Error
// @Router       /automats/qr-scan [post]
func (r *parcelAutomatRoutes) processQRScan(c *gin.Context) {
	var req request.QRScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	automatID, err := uuid.Parse(req.ParcelAutomatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid parcel automat ID"})
		return
	}

	cellIDs, err := r.uc.ProcessQRScan(c.Request.Context(), req.QRData, automatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	cellIDStrings := make([]string, 0, len(cellIDs))
	for _, id := range cellIDs {
		cellIDStrings = append(cellIDStrings, id.String())
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "QR code processed successfully",
		"cell_ids": cellIDStrings,
	})
}

// @Summary      Подтверждение получения товара
// @Description  Подтверждает, что пользователь забрал товар из ячейки
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        request body request.ConfirmPickupRequest true "ID ячеек"
// @Success      200 {object} map[string]any
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Router       /automats/confirm-pickup [post]
func (r *parcelAutomatRoutes) confirmPickup(c *gin.Context) {
	var req request.ConfirmPickupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	cellIDs := make([]uuid.UUID, 0, len(req.CellIDs))
	for _, idStr := range req.CellIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		cellIDs = append(cellIDs, id)
	}

	if err := r.uc.ConfirmPickup(c.Request.Context(), cellIDs); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pickup confirmed successfully",
	})
}
