package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/request"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/response"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase"
)

type parcelAutomatRoutes struct {
	uc *usecase.ParcelAutomatUseCase
}

func newParcelAutomatRoutes(public *gin.RouterGroup, protected *gin.RouterGroup, uc *usecase.ParcelAutomatUseCase) {
	r := &parcelAutomatRoutes{uc: uc}

	publicGroup := public.Group("/automats")
	{
		publicGroup.POST("/qr-scan", r.processQRScan)
		publicGroup.POST("/confirm-pickup", r.confirmPickup)
	}

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

// @Summary      Create parcel automat
// @Description  Creates a new parcel automat with specified parameters
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        request body request.CreateParcelAutomatRequest true "Parcel automat data"
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

	automat := &entity.ParcelAutomat{
		City:          req.City,
		Address:       req.Address,
		IPAddress:     req.IPAddress,
		Coordinates:   req.Coordinates,
		NumberOfCells: req.NumberOfCells,
		ArucoID:       req.ArucoID,
	}

	createdAutomat, err := r.uc.Create(c.Request.Context(), automat, req.Cells)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, createdAutomat)
}

// @Summary      List parcel automats
// @Description  Returns list of all parcel automats
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
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, automats)
}

// @Summary      List working parcel automats
// @Description  Returns list of parcel automats with "working" status
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
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, automats)
}

// @Summary      Get parcel automat
// @Description  Returns parcel automat information by ID
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "Parcel automat ID"
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
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, automat)
}

// @Summary      Update parcel automat
// @Description  Updates parcel automat information (city, address, IP, coordinates)
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "Parcel automat ID"
// @Param        request body request.UpdateParcelAutomatRequest true "Update data"
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

	automat := &entity.ParcelAutomat{
		ID:          id,
		City:        req.City,
		Address:     req.Address,
		IPAddress:   req.IPAddress,
		Coordinates: req.Coordinates,
	}

	updatedAutomat, err := r.uc.Update(c.Request.Context(), automat)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, updatedAutomat)
}

// @Summary      Get parcel automat cells
// @Description  Returns list of all cells of the specified parcel automat
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "Parcel automat ID"
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
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, cells)
}

// @Summary      Update cell
// @Description  Updates parcel automat cell dimensions
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "Parcel automat ID"
// @Param        cellId path string true "Cell ID"
// @Param        request body request.UpdateCellRequest true "New cell dimensions"
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
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, cell)
}

// @Summary      Update parcel automat status
// @Description  Enables or disables parcel automat
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "Parcel automat ID"
// @Param        request body request.UpdateParcelAutomatStatusRequest true "New status"
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

	automat := &entity.ParcelAutomat{
		ID:        id,
		IsWorking: req.IsWorking,
	}

	updatedAutomat, err := r.uc.UpdateStatus(c.Request.Context(), automat)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, updatedAutomat)
}

// @Summary      Delete parcel automat
// @Description  Deletes parcel automat by ID
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        id path string true "Parcel automat ID"
// @Success      204
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
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
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      QR code scanning
// @Description  Processes user QR code and returns cell IDs with their orders
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        request body request.QRScanRequest true "QR data"
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
		handleError(c, err)
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

// @Summary      Confirm goods pickup
// @Description  Confirms that user picked up goods from cell
// @Tags         automats
// @Accept       json
// @Produce      json
// @Param        request body request.ConfirmPickupRequest true "Cell IDs"
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
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pickup confirmed successfully",
	})
}
