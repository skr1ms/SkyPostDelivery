package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase"
)

type AdminWebSocketHandler struct {
	connectedAdmins   map[string]*websocket.Conn
	mu                sync.RWMutex
	upgrader          websocket.Upgrader
	droneManager      *usecase.DroneManagerUseCase
	droneRepo         repo.DroneRepo
	broadcastInterval int
	broadcastStarted  bool
}

func NewAdminWebSocketHandler(droneManager *usecase.DroneManagerUseCase, droneRepo repo.DroneRepo, broadcastInterval int) *AdminWebSocketHandler {
	return &AdminWebSocketHandler{
		connectedAdmins:   make(map[string]*websocket.Conn),
		droneManager:      droneManager,
		droneRepo:         droneRepo,
		broadcastInterval: broadcastInterval,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// @Summary      WebSocket подключение администратора
// @Description  Устанавливает WebSocket соединение для мониторинга дронов администратором
// @Description  Администратор получает обновления статуса всех дронов и доставок в реальном времени
// @Tags         websocket
// @Accept       json
// @Produce      json
// @Param        admin_id query string false "ID администратора"
// @Success      101 {string} string "Switching Protocols"
// @Router       /ws/admin [get]
func (h *AdminWebSocketHandler) HandleAdminConnection(c *gin.Context) {
	adminID := c.Query("admin_id")
	if adminID == "" {
		adminID = "admin"
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade admin connection: %v", err)
		return
	}
	defer conn.Close()

	h.mu.Lock()
	h.connectedAdmins[adminID] = conn
	if !h.broadcastStarted {
		h.broadcastStarted = true
		go h.startBroadcastLoop()
	}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.connectedAdmins, adminID)
		h.mu.Unlock()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from admin %s: %v", adminID, err)
			break
		}

		if err := h.processAdminMessage(context.Background(), adminID, conn, message); err != nil {
			log.Printf("Error processing message from admin %s: %v", adminID, err)
		}
	}
}

func (h *AdminWebSocketHandler) processAdminMessage(_ context.Context, _ string, conn *websocket.Conn, message []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		return err
	}

	msgType, ok := data["type"].(string)
	if !ok {
		return nil
	}

	if msgType == "ping" {
		response := map[string]interface{}{
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

		message := map[string]interface{}{
			"type":      "drones_status",
			"timestamp": time.Now().Format(time.RFC3339),
			"drones":    dronesStatus,
		}

		h.broadcast(message)
	}
}

func (h *AdminWebSocketHandler) getAllDronesStatus(ctx context.Context) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	drones := h.droneManager.GetAllDrones()
	for _, droneID := range drones {
		state, err := h.droneRepo.GetDroneState(ctx, droneID)
		if err != nil {
			log.Printf("Failed to get state for drone %s: %v", droneID, err)
			continue
		}

		droneStatus := map[string]interface{}{
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

func (h *AdminWebSocketHandler) broadcast(message map[string]interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conn := range h.connectedAdmins {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error broadcasting to admin: %v", err)
		}
	}
}

func (h *AdminWebSocketHandler) BroadcastToAdmins(message map[string]interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := map[string]interface{}{
		"type":      "drone_update",
		"timestamp": time.Now().Format(time.RFC3339),
		"payload":   message,
	}

	for _, conn := range h.connectedAdmins {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Error broadcasting to admin: %v", err)
		}
	}
}
