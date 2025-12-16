package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
)

type cellRoutes struct {
	cellManager *usecase.CellManagerUseCase
	logger      logger.Interface
}

func newCellRoutes(group *gin.RouterGroup, cellManager *usecase.CellManagerUseCase, log logger.Interface) {
	r := &cellRoutes{
		cellManager: cellManager,
		logger:      log,
	}

	group.POST("/sync", r.syncCells)
	group.GET("/mapping", r.getMapping)
	group.POST("/:number/open", r.openCell)
	group.POST("/internal/:number/open", r.openInternalDoor)
	group.POST("/prepare", r.prepareCell)
	group.GET("/count", r.getCellsCount)
}

func (r *cellRoutes) syncCells(c *gin.Context) {
	var req entity.SyncCellsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Error("Failed to bind sync cells request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.cellManager.SyncCells(c.Request.Context(), &req); err != nil {
		r.logger.Error("Failed to sync cells", err)
		handleError(c, err)
		return
	}

	mapping := r.cellManager.GetMapping()

	c.JSON(http.StatusOK, entity.SyncCellsResponse{
		Message:            "Cells synchronized successfully",
		CellsCount:         len(mapping.ExternalCells),
		InternalCellsCount: len(mapping.InternalCells),
		ParcelAutomatID:    mapping.ParcelAutomatID.String(),
	})
}

func (r *cellRoutes) getMapping(c *gin.Context) {
	mapping := r.cellManager.GetMapping()

	externalMapping := make(map[string]entity.CellInfo)
	for num, uuid := range mapping.ExternalCells {
		externalMapping[strconv.Itoa(num)] = entity.CellInfo{
			CellUUID:        uuid.String(),
			ParcelAutomatID: mapping.ParcelAutomatID.String(),
		}
	}

	internalMapping := make(map[string]entity.CellInfo)
	for num, uuid := range mapping.InternalCells {
		internalMapping[strconv.Itoa(num)] = entity.CellInfo{
			CellUUID:        uuid.String(),
			ParcelAutomatID: mapping.ParcelAutomatID.String(),
		}
	}

	c.JSON(http.StatusOK, entity.CellMappingResponse{
		Mapping:            externalMapping,
		CellsCount:         len(externalMapping),
		InternalMapping:    internalMapping,
		InternalCellsCount: len(internalMapping),
		ParcelAutomatID:    mapping.ParcelAutomatID.String(),
		Initialized:        mapping.Initialized,
	})
}

func (r *cellRoutes) openCell(c *gin.Context) {
	cellNumber, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		r.logger.Error("Invalid cell number", err, map[string]any{
			"param": c.Param("number"),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cell number"})
		return
	}

	var req entity.OpenCellRequest
	_ = c.ShouldBindJSON(&req)

	resp, err := r.cellManager.OpenCell(c.Request.Context(), cellNumber, req.OrderNumber)
	if err != nil {
		r.logger.Error("Failed to open cell", err, map[string]any{
			"cell_number": cellNumber,
		})
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (r *cellRoutes) openInternalDoor(c *gin.Context) {
	doorNumber, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		r.logger.Error("Invalid door number", err, map[string]any{
			"param": c.Param("number"),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid door number"})
		return
	}

	resp, err := r.cellManager.OpenInternalDoor(c.Request.Context(), doorNumber)
	if err != nil {
		r.logger.Error("Failed to open internal door", err, map[string]any{
			"door_number": doorNumber,
		})
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (r *cellRoutes) prepareCell(c *gin.Context) {
	var req entity.PrepareCellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Error("Failed to bind prepare cell request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := uuid.Parse(req.CellID); err != nil {
		r.logger.Error("Invalid cell UUID", err, map[string]any{
			"cell_id": req.CellID,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cell UUID"})
		return
	}

	resp, err := r.cellManager.PrepareCell(c.Request.Context(), req.CellID)
	if err != nil {
		r.logger.Error("Failed to prepare cell", err, map[string]any{
			"cell_id": req.CellID,
		})
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (r *cellRoutes) getCellsCount(c *gin.Context) {
	response, err := r.cellManager.GetCellsCount()
	if err != nil {
		r.logger.Error("Failed to get cells count", err)
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}
