package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

type DroneWebSocketHandler struct {
	connUC          *usecase.DroneConnectionUseCase
	telemetryUC     *usecase.DroneTelemetryUseCase
	deliveryUC      *usecase.DroneDeliveryUseCase
	commandUC       *usecase.DroneCommandUseCase
	connectedDrones map[string]*SafeConn
	ipToID          map[string]string
	mu              sync.RWMutex
	upgrader        websocket.Upgrader
	logger          logger.Interface
}

func NewDroneWebSocketHandler(
	connUC *usecase.DroneConnectionUseCase,
	telemetryUC *usecase.DroneTelemetryUseCase,
	deliveryUC *usecase.DroneDeliveryUseCase,
	commandUC *usecase.DroneCommandUseCase,
	log logger.Interface,
) *DroneWebSocketHandler {
	return &DroneWebSocketHandler{
		connUC:          connUC,
		telemetryUC:     telemetryUC,
		deliveryUC:      deliveryUC,
		commandUC:       commandUC,
		connectedDrones: make(map[string]*SafeConn),
		ipToID:          make(map[string]string),
		logger:          log,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// @Summary      Drone WebSocket connection
// @Description  Establishes WebSocket connection for drone. First message must be registration with ip_address
// @Description  Supported message types: heartbeat, status_update, delivery_update, video_frame, arrived_at_destination, cargo_dropped
// @Tags         websocket
// @Accept       json
// @Produce      json
// @Param        drone_id path string false "Drone ID (optional)"
// @Success      101 {string} string "Switching Protocols"
// @Router       /ws/drone [get]
// @Router       /ws/drone/{drone_id} [get]
func (h *DroneWebSocketHandler) HandleDroneConnection(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", err, nil)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	safeConn := &SafeConn{Conn: conn}
	ctx := c.Request.Context()

	_, message, err := conn.ReadMessage()
	if err != nil {
		h.logger.Error("Failed to read registration message", err, nil)
		return
	}

	var regData map[string]any
	if err := json.Unmarshal(message, &regData); err != nil {
		h.logger.Error("Failed to unmarshal registration message", err, nil)
		return
	}

	if regData["type"] != "register" {
		_ = safeConn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "First message must be registration with ip_address",
		})
		return
	}

	ipAddress, ok := regData["ip_address"].(string)
	if !ok || ipAddress == "" {
		_ = safeConn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "ip_address is required for registration",
		})
		return
	}

	droneID, err := h.connUC.RegisterDrone(ctx, ipAddress)
	if err != nil {
		_ = safeConn.WriteJSON(map[string]string{
			"type":    "error",
			"message": err.Error(),
		})
		return
	}

	h.mu.Lock()
	h.connectedDrones[droneID] = safeConn
	h.ipToID[ipAddress] = droneID
	h.mu.Unlock()

	_ = safeConn.WriteJSON(map[string]any{
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
		_ = h.connUC.UnregisterDrone(ctx, droneID)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			h.logger.Warn("Connection closed for drone", err, map[string]any{"drone_id": droneID})
			break
		}

		if err := h.processDroneMessage(ctx, droneID, message); err != nil {
			handleWebSocketError(err, safeConn, h.logger)
		}
	}
}

func (h *DroneWebSocketHandler) processDroneMessage(ctx context.Context, droneID string, message []byte) error {
	var data map[string]any
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

func (h *DroneWebSocketHandler) handleHeartbeat(ctx context.Context, droneID string, data map[string]any) error {
	payload, ok := data["payload"].(map[string]any)
	if !ok {
		payload = data
	}

	return h.telemetryUC.ProcessHeartbeat(ctx, droneID, payload)
}

func (h *DroneWebSocketHandler) handleStatusUpdate(ctx context.Context, droneID string, data map[string]any) error {
	payload, ok := data["payload"].(map[string]any)
	if !ok {
		return nil
	}

	return h.telemetryUC.ProcessStatusUpdate(ctx, droneID, payload)
}

func (h *DroneWebSocketHandler) handleDeliveryUpdate(ctx context.Context, droneID string, data map[string]any) error {
	payload, ok := data["payload"].(map[string]any)
	if !ok {
		return nil
	}

	return h.deliveryUC.ProcessDeliveryUpdate(ctx, droneID, payload)
}

func (h *DroneWebSocketHandler) handleArrivedAtDestination(ctx context.Context, droneID string, data map[string]any) error {
	payload, ok := data["payload"].(map[string]any)
	if !ok {
		return nil
	}

	return h.deliveryUC.ProcessArrivedAtDestination(ctx, droneID, payload)
}

func (h *DroneWebSocketHandler) handleCargoDropped(ctx context.Context, data map[string]any) error {
	payload, ok := data["payload"].(map[string]any)
	if !ok {
		return nil
	}

	return h.deliveryUC.ProcessCargoDropped(ctx, payload)
}

func (h *DroneWebSocketHandler) handleVideoFrame(ctx context.Context, droneID string, data map[string]any) error {
	payload, ok := data["payload"].(map[string]any)
	if !ok {
		payload = data
	}

	frameData, _ := payload["frame"].(string)
	deliveryID, _ := payload["delivery_id"].(string)

	if frameData != "" {
		return h.deliveryUC.ProcessVideoFrame(ctx, droneID, []byte(frameData), deliveryID)
	}

	return nil
}

func (h *DroneWebSocketHandler) SendToDrone(ctx context.Context, droneID string, message map[string]any) error {
	h.mu.RLock()
	conn, exists := h.connectedDrones[droneID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("drone %s is not connected", droneID)
	}

	return conn.WriteJSON(message)
}
