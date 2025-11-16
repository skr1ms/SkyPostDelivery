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
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase"
)

type DroneWebSocketHandler struct {
	stateRepo        usecase.DroneStateRepo
	deliveryTaskRepo usecase.DeliveryTaskRepo
	droneManager     usecase.DroneManager
	deliveryUseCase  *usecase.DeliveryUseCase
	videoHandler     *VideoHandler
	connectedDrones  map[string]*websocket.Conn
	ipToID           map[string]string
	mu               sync.RWMutex
	upgrader         websocket.Upgrader
}

func NewDroneWebSocketHandler(
	stateRepo usecase.DroneStateRepo,
	deliveryTaskRepo usecase.DeliveryTaskRepo,
	droneManager usecase.DroneManager,
	videoHandler *VideoHandler,
) *DroneWebSocketHandler {
	return &DroneWebSocketHandler{
		stateRepo:        stateRepo,
		deliveryTaskRepo: deliveryTaskRepo,
		droneManager:     droneManager,
		videoHandler:     videoHandler,
		connectedDrones:  make(map[string]*websocket.Conn),
		ipToID:           make(map[string]string),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *DroneWebSocketHandler) SetDeliveryUseCase(deliveryUseCase *usecase.DeliveryUseCase) {
	h.deliveryUseCase = deliveryUseCase
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
	defer conn.Close()

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
		conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "First message must be registration with ip_address",
		})
		return
	}

	ipAddress, ok := regData["ip_address"].(string)
	if !ok || ipAddress == "" {
		conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "ip_address is required for registration",
		})
		return
	}

	droneID, err := h.stateRepo.GetDroneIDByIP(ctx, ipAddress)
	if err != nil || droneID == "" {
		conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": fmt.Sprintf("No drone found with IP %s", ipAddress),
		})
		return
	}

	h.mu.Lock()
	h.connectedDrones[droneID] = conn
	h.ipToID[ipAddress] = droneID
	h.mu.Unlock()

	if err := h.droneManager.RegisterDrone(ctx, droneID); err != nil {
		log.Printf("Failed to register drone %s: %v", droneID, err)
	}

	conn.WriteJSON(map[string]interface{}{
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
		h.droneManager.UnregisterDrone(ctx, droneID)
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

	status, _ := payload["status"].(string)
	if status == "" {
		status = "idle"
	}

	batteryLevel, _ := payload["battery_level"].(float64)
	if batteryLevel == 0 {
		batteryLevel = 100.0
	}

	currentDeliveryID, _ := payload["current_delivery_id"].(string)
	errorMessage, _ := payload["error_message"].(string)

	position := entity.Position{}
	if pos, ok := payload["position"].(map[string]interface{}); ok {
		position.Latitude, _ = pos["latitude"].(float64)
		position.Longitude, _ = pos["longitude"].(float64)
		position.Altitude, _ = pos["altitude"].(float64)
	}

	speed, _ := payload["speed"].(float64)

	if batteryLevel > 0 {
		if err := h.stateRepo.UpdateDroneBattery(ctx, droneID, batteryLevel); err != nil {
			return err
		}
	}

	state := &entity.DroneState{
		DroneID:         droneID,
		Status:          entity.DroneStatus(status),
		BatteryLevel:    batteryLevel,
		CurrentPosition: position,
		Speed:           speed,
		LastUpdated:     time.Now(),
	}

	if currentDeliveryID != "" {
		state.CurrentDeliveryID = &currentDeliveryID
	}

	if errorMessage != "" {
		state.ErrorMessage = &errorMessage
	}

	return h.stateRepo.SaveDroneState(ctx, state)
}

func (h *DroneWebSocketHandler) handleStatusUpdate(ctx context.Context, droneID string, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil
	}

	status, _ := payload["status"].(string)
	batteryLevel, _ := payload["battery_level"].(float64)
	currentDeliveryID, _ := payload["current_delivery_id"].(string)
	errorMessage, _ := payload["error_message"].(string)

	position := entity.Position{}
	if pos, ok := payload["position"].(map[string]interface{}); ok {
		position.Latitude, _ = pos["latitude"].(float64)
		position.Longitude, _ = pos["longitude"].(float64)
		position.Altitude, _ = pos["altitude"].(float64)
	}

	speed, _ := payload["speed"].(float64)

	state := &entity.DroneState{
		DroneID:         droneID,
		Status:          entity.DroneStatus(status),
		BatteryLevel:    batteryLevel,
		CurrentPosition: position,
		Speed:           speed,
		LastUpdated:     time.Now(),
	}

	if currentDeliveryID != "" {
		state.CurrentDeliveryID = &currentDeliveryID
	}

	if errorMessage != "" {
		state.ErrorMessage = &errorMessage
	}

	return h.stateRepo.SaveDroneState(ctx, state)
}

func (h *DroneWebSocketHandler) handleDeliveryUpdate(ctx context.Context, droneID string, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil
	}

	droneStatus, _ := payload["drone_status"].(string)
	orderID, _ := payload["order_id"].(string)
	parcelAutomatID, _ := payload["parcel_automat_id"].(string)

	if droneStatus == "arrived_at_destination" {
		if orderID != "" && parcelAutomatID != "" && h.deliveryUseCase != nil {
			_, err := h.deliveryUseCase.HandleDroneArrived(ctx, droneID, orderID, parcelAutomatID)
			return err
		}
		return nil
	}

	deliveryID, ok := payload["delivery_id"].(string)
	if !ok || deliveryID == "" {
		return nil
	}

	switch droneStatus {
	case "arrived_at_locker":
		return h.deliveryTaskRepo.UpdateDeliveryStatus(ctx, deliveryID, entity.DeliveryStatusInProgress, nil)
	case "returning":
		return h.droneManager.ReleaseDrone(ctx, droneID)
	}

	return nil
}

func (h *DroneWebSocketHandler) handleArrivedAtDestination(ctx context.Context, droneID string, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil
	}

	orderID, _ := payload["order_id"].(string)
	parcelAutomatID, _ := payload["parcel_automat_id"].(string)

	if orderID == "" || parcelAutomatID == "" {
		return nil
	}

	if h.deliveryUseCase != nil {
		_, err := h.deliveryUseCase.HandleDroneArrived(ctx, droneID, orderID, parcelAutomatID)
		return err
	}

	return nil
}

func (h *DroneWebSocketHandler) handleCargoDropped(ctx context.Context, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil
	}

	orderID, _ := payload["order_id"].(string)
	lockerCellID, _ := payload["locker_cell_id"].(string)

	if orderID == "" {
		return nil
	}

	if h.deliveryUseCase != nil {
		_, err := h.deliveryUseCase.HandleCargoDropped(ctx, orderID, lockerCellID)
		return err
	}

	return nil
}

func (h *DroneWebSocketHandler) handleVideoFrame(_ context.Context, droneID string, data map[string]interface{}) error {
	if h.videoHandler == nil {
		return nil
	}

	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		payload = data
	}

	frameData, _ := payload["frame"].(string)
	deliveryID, _ := payload["delivery_id"].(string)

	if frameData != "" {
		h.videoHandler.BroadcastFrameToAdmins(droneID, frameData, deliveryID)
	}

	return nil
}

func (h *DroneWebSocketHandler) SendTaskToDrone(ctx context.Context, droneID string, task map[string]interface{}) error {
	h.mu.RLock()
	conn, exists := h.connectedDrones[droneID]
	h.mu.RUnlock()

	if !exists {
		return nil
	}

	message := map[string]interface{}{
		"type":      "delivery_task",
		"timestamp": time.Now().Format(time.RFC3339),
		"payload":   task,
	}

	return conn.WriteJSON(message)
}

func (h *DroneWebSocketHandler) SendCommandToDrone(ctx context.Context, droneID string, command map[string]interface{}) error {
	h.mu.RLock()
	conn, exists := h.connectedDrones[droneID]
	h.mu.RUnlock()

	if !exists {
		return nil
	}

	message := map[string]interface{}{
		"type":      "command",
		"timestamp": time.Now().Format(time.RFC3339),
		"payload":   command,
	}

	return conn.WriteJSON(message)
}
