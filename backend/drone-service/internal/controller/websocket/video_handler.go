package websocket

import (
	"context"
	"encoding/base64"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/v1/response"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/minio"
)

type VideoHandler struct {
	adminVideoConnections map[string]map[*SafeConn]bool
	minioClient           *minio.Client
	frameCounters         map[string]int
	mu                    sync.RWMutex
	upgrader              websocket.Upgrader
	logger                logger.Interface
}

func NewVideoHandler(minioClient *minio.Client, log logger.Interface) *VideoHandler {
	return &VideoHandler{
		adminVideoConnections: make(map[string]map[*SafeConn]bool),
		minioClient:           minioClient,
		frameCounters:         make(map[string]int),
		logger:                log,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// @Summary      WebSocket for drone video stream
// @Description  Establishes WebSocket connection for receiving video stream from specific drone
// @Description  Administrator connects to this endpoint to view real-time video
// @Tags         websocket
// @Accept       json
// @Produce      octet-stream
// @Param        drone_id path string true "Drone ID"
// @Success      101 {string} string "Switching Protocols"
// @Failure      400 {object} response.Error "drone_id required"
// @Router       /ws/drone/{drone_id}/video [get]
func (h *VideoHandler) HandleAdminVideoConnection(c *gin.Context) {
	droneID := c.Param("drone_id")
	if droneID == "" {
		c.JSON(http.StatusBadRequest, response.Error{Error: "drone_id required"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", err, nil)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	safeConn := &SafeConn{Conn: conn}

	h.mu.Lock()
	if h.adminVideoConnections[droneID] == nil {
		h.adminVideoConnections[droneID] = make(map[*SafeConn]bool)
	}
	h.adminVideoConnections[droneID][safeConn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.adminVideoConnections[droneID], safeConn)
		if len(h.adminVideoConnections[droneID]) == 0 {
			delete(h.adminVideoConnections, droneID)
		}
		h.mu.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			h.logger.Warn("Video connection closed", err, map[string]any{"drone_id": droneID})
			break
		}
	}
}

func (h *VideoHandler) BroadcastFrameToAdmins(droneID string, frameData string, deliveryID string) {
	h.mu.RLock()
	connections, exists := h.adminVideoConnections[droneID]
	h.mu.RUnlock()

	if !exists || len(connections) == 0 {
		return
	}

	h.mu.Lock()
	for conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(frameData)); err != nil {
			h.logger.Warn("Failed to send video frame to admin", err, map[string]any{"drone_id": droneID})
			delete(connections, conn)
		}
	}
	h.mu.Unlock()

	if deliveryID != "" {
		go h.saveFrameToMinIO(droneID, deliveryID, frameData)
	}
}

func (h *VideoHandler) saveFrameToMinIO(droneID, deliveryID, frameData string) {
	h.mu.Lock()
	if _, exists := h.frameCounters[droneID]; !exists {
		h.frameCounters[droneID] = 0
	}
	h.frameCounters[droneID]++
	frameNumber := h.frameCounters[droneID]
	h.mu.Unlock()

	frameBytes, err := base64.StdEncoding.DecodeString(frameData)
	if err != nil {
		return
	}

	_, _ = h.minioClient.UploadFrame(context.Background(), droneID, deliveryID, frameBytes, frameNumber)
}

func (h *VideoHandler) HandleVideoFrame(ctx context.Context, droneID string, frameData []byte, deliveryID string) error {
	frameStr := string(frameData)

	h.BroadcastFrameToAdmins(droneID, frameStr, deliveryID)

	return nil
}
