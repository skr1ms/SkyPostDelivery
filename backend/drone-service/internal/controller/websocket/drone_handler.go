package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase"
)

type DroneWebSocketHandler struct {
	droneMessageUseCase *usecase.DroneMessageUseCase
	connectedDrones     map[string]*websocket.Conn
	ipToID              map[string]string
	mu                  sync.RWMutex
	upgrader            websocket.Upgrader
}

func NewDroneWebSocketHandler(
	droneMessageUseCase *usecase.DroneMessageUseCase,
) *DroneWebSocketHandler {
	return &DroneWebSocketHandler{
		droneMessageUseCase: droneMessageUseCase,
		connectedDrones:     make(map[string]*websocket.Conn),
		ipToID:              make(map[string]string),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// @Summary      WebSocket подключение дрона
// @Description  Устанавливает WebSocket соединение для дрона. Первое сообщение должно быть регистрацией с ip_address
// @Description  Поддерживаемые типы сообщений: heartbeat, status_update, delivery_update, video_frame, arrived_at_destination, cargo_dropped
// @Tags         websocket
// @Accept       json
// @Produce      json
// @Param        drone_id path string false "ID дрона (опционально)"
// @Success      101 {string} string "Switching Protocols"
// @Router       /ws/drone [get]
// @Router       /ws/drone/{drone_id} [get]
func (h *DroneWebSocketHandler) HandleDroneConnection(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	ctx := c.Request.Context()

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Failed to read registration message: %v", err)
		return
	}

	var regData map[string]interface{}
	if err := json.Unmarshal(message, &regData); err != nil {
		log.Printf("Failed to unmarshal registration message: %v", err)
		return
	}

	if regData["type"] != "register" {
		_ = conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "First message must be registration with ip_address",
		})
		return
	}

	ipAddress, ok := regData["ip_address"].(string)
	if !ok || ipAddress == "" {
		_ = conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "ip_address is required for registration",
		})
		return
	}

	droneID, err := h.droneMessageUseCase.RegisterDrone(ctx, ipAddress)
	if err != nil {
		_ = conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": err.Error(),
		})
		return
	}

	h.mu.Lock()
	h.connectedDrones[droneID] = conn
	h.ipToID[ipAddress] = droneID
	h.mu.Unlock()

	_ = conn.WriteJSON(map[string]interface{}{
		"type":      "registered",
		"drone_id":  droneID,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	defer func() {
		h.mu.Lock()
		delete(h.connectedDrones, droneID)
		for ip, id := range h.ipToID {
			if id == droneID {
				delete(h.ipToID, ip)
				break
			}
		}
		h.mu.Unlock()
		_ = h.droneMessageUseCase.UnregisterDrone(ctx, droneID)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Connection closed for drone %s: %v", droneID, err)
			break
		}

		if err := h.processDroneMessage(ctx, droneID, message); err != nil {
			log.Printf("Error processing message from drone %s: %v", droneID, err)
		}
	}
}

func (h *DroneWebSocketHandler) processDroneMessage(ctx context.Context, droneID string, message []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		return err
	}

	messageType, ok := data["type"].(string)
	if !ok {
		return nil
	}

	switch messageType {
	case "heartbeat":
		return h.handleHeartbeat(ctx, droneID, data)
	case "status_update":
		return h.handleStatusUpdate(ctx, droneID, data)
	case "delivery_update":
		return h.handleDeliveryUpdate(ctx, droneID, data)
	case "arrived_at_destination":
		return h.handleArrivedAtDestination(ctx, droneID, data)
	case "cargo_dropped":
		return h.handleCargoDropped(ctx, data)
	case "video_frame":
		return h.handleVideoFrame(ctx, droneID, data)
	}

	return nil
}

func (h *DroneWebSocketHandler) handleHeartbeat(ctx context.Context, droneID string, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		payload = data
	}

	return h.droneMessageUseCase.ProcessHeartbeat(ctx, droneID, payload)
}

func (h *DroneWebSocketHandler) handleStatusUpdate(ctx context.Context, droneID string, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil
	}

	return h.droneMessageUseCase.ProcessStatusUpdate(ctx, droneID, payload)
}

func (h *DroneWebSocketHandler) handleDeliveryUpdate(ctx context.Context, droneID string, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil
	}

	return h.droneMessageUseCase.ProcessDeliveryUpdate(ctx, droneID, payload)
}

func (h *DroneWebSocketHandler) handleArrivedAtDestination(ctx context.Context, droneID string, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil
	}

	return h.droneMessageUseCase.ProcessArrivedAtDestination(ctx, droneID, payload)
}

func (h *DroneWebSocketHandler) handleCargoDropped(ctx context.Context, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil
	}

	return h.droneMessageUseCase.ProcessCargoDropped(ctx, payload)
}

func (h *DroneWebSocketHandler) handleVideoFrame(ctx context.Context, droneID string, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		payload = data
	}

	frameData, _ := payload["frame"].(string)
	deliveryID, _ := payload["delivery_id"].(string)

	if frameData != "" {
		return h.droneMessageUseCase.ProcessVideoFrame(ctx, droneID, []byte(frameData), deliveryID)
	}

	return nil
}

func (h *DroneWebSocketHandler) SendToDrone(ctx context.Context, droneID string, message map[string]interface{}) error {
	h.mu.RLock()
	conn, exists := h.connectedDrones[droneID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("drone %s is not connected", droneID)
	}

	return conn.WriteJSON(message)
}
