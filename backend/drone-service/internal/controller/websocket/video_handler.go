package websocket

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/v1/response"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/minio"
)

type VideoHandler struct {
	adminVideoConnections map[string]map[*websocket.Conn]bool
	minioClient           *minio.Client
	frameCounters         map[string]int
	mu                    sync.RWMutex
	upgrader              websocket.Upgrader
}

func NewVideoHandler(minioClient *minio.Client) *VideoHandler {
	return &VideoHandler{
		adminVideoConnections: make(map[string]map[*websocket.Conn]bool),
		minioClient:           minioClient,
		frameCounters:         make(map[string]int),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// @Summary      WebSocket для видео потока дрона
// @Description  Устанавливает WebSocket соединение для получения видео потока от конкретного дрона
// @Description  Администратор подключается к этому эндпоинту для просмотра видео в реальном времени
// @Tags         websocket
// @Accept       json
// @Produce      octet-stream
// @Param        drone_id path string true "ID дрона"
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
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	h.mu.Lock()
	if h.adminVideoConnections[droneID] == nil {
		h.adminVideoConnections[droneID] = make(map[*websocket.Conn]bool)
	}
	h.adminVideoConnections[droneID][conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.adminVideoConnections[droneID], conn)
		if len(h.adminVideoConnections[droneID]) == 0 {
			delete(h.adminVideoConnections, droneID)
		}
		h.mu.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
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
