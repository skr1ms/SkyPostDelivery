package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

type AdminWebSocketHandler struct {
	connectedAdmins   map[string]*SafeConn
	mu                sync.RWMutex
	upgrader          websocket.Upgrader
	droneManager      *usecase.DroneManagerUseCase
	droneRepo         repo.DroneRepo
	broadcastInterval int
	broadcastOnce     sync.Once
	logger            logger.Interface
}

func NewAdminWebSocketHandler(droneManager *usecase.DroneManagerUseCase, droneRepo repo.DroneRepo, broadcastInterval int, log logger.Interface) *AdminWebSocketHandler {
	return &AdminWebSocketHandler{
		connectedAdmins:   make(map[string]*SafeConn),
		droneManager:      droneManager,
		droneRepo:         droneRepo,
		broadcastInterval: broadcastInterval,
		logger:            log,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// @Summary      Admin WebSocket connection
// @Description  Establishes WebSocket connection for drone monitoring by administrator
// @Description  Administrator receives real-time updates of all drones and deliveries status
// @Tags         websocket
// @Accept       json
// @Produce      json
// @Param        admin_id query string false "Administrator ID"
// @Success      101 {string} string "Switching Protocols"
// @Router       /ws/admin [get]
func (h *AdminWebSocketHandler) HandleAdminConnection(c *gin.Context) {
	adminID := c.Query("admin_id")
	if adminID == "" {
		adminID = "admin"
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade admin connection", err, nil)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	safeConn := &SafeConn{Conn: conn}

	h.mu.Lock()
	h.connectedAdmins[adminID] = safeConn
	h.mu.Unlock()

	h.broadcastOnce.Do(func() {
		go h.startBroadcastLoop()
	})

	defer func() {
		h.mu.Lock()
		delete(h.connectedAdmins, adminID)
		h.mu.Unlock()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			h.logger.Error("Error reading message from admin", err, map[string]any{"admin_id": adminID})
			break
		}

		if err := h.processAdminMessage(context.Background(), adminID, safeConn, message); err != nil {
			h.logger.Error("Error processing message from admin", err, map[string]any{"admin_id": adminID})
		}
	}
}

func (h *AdminWebSocketHandler) processAdminMessage(_ context.Context, _ string, conn *SafeConn, message []byte) error {
	var data map[string]any
	if err := json.Unmarshal(message, &data); err != nil {
		return err
	}

	msgType, ok := data["type"].(string)
	if !ok {
		return nil
	}

	if msgType == "ping" {
		response := map[string]any{
			"type":      "pong",
			"timestamp": time.Now().Format(time.RFC3339),
		}
		return conn.WriteJSON(response)
	}

	return nil
}

func (h *AdminWebSocketHandler) startBroadcastLoop() {
	ticker := time.NewTicker(time.Duration(h.broadcastInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		dronesStatus := h.getAllDronesStatus(ctx)

		message := map[string]any{
			"type":      "drones_status",
			"timestamp": time.Now().Format(time.RFC3339),
			"drones":    dronesStatus,
		}

		h.broadcast(message)
	}
}

func (h *AdminWebSocketHandler) getAllDronesStatus(ctx context.Context) []map[string]any {
	result := make([]map[string]any, 0)

	drones := h.droneManager.GetAllDrones()
	for _, droneID := range drones {
		state, err := h.droneRepo.GetDroneState(ctx, droneID)
		if err != nil {
			h.logger.Error("Failed to get state for drone", err, map[string]any{"drone_id": droneID})
			continue
		}

		droneStatus := map[string]any{
			"drone_id":            droneID,
			"status":              state.Status,
			"battery_level":       state.BatteryLevel,
			"position":            state.CurrentPosition,
			"speed":               state.Speed,
			"current_delivery_id": state.CurrentDeliveryID,
			"error_message":       state.ErrorMessage,
			"last_updated":        state.LastUpdated.Format(time.RFC3339),
		}

		result = append(result, droneStatus)
	}

	return result
}

func (h *AdminWebSocketHandler) broadcast(message map[string]any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for adminID, conn := range h.connectedAdmins {
		if err := conn.WriteJSON(message); err != nil {
			h.logger.Warn("Error broadcasting to admin", err, map[string]any{"admin_id": adminID})
		}
	}
}

func (h *AdminWebSocketHandler) BroadcastToAdmins(message map[string]any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := map[string]any{
		"type":      "drone_update",
		"timestamp": time.Now().Format(time.RFC3339),
		"payload":   message,
	}

	for adminID, conn := range h.connectedAdmins {
		if err := conn.WriteJSON(msg); err != nil {
			h.logger.Warn("Error broadcasting to admin", err, map[string]any{"admin_id": adminID})
		}
	}
}
