package websocket

import (
	"errors"
	"time"

	"github.com/gorilla/websocket"
	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

func handleWebSocketError(err error, conn *SafeConn, logger logger.Interface) {
	if err == nil {
		return
	}

	logger.Error("WS Error", err, nil)

	response := map[string]string{
		"status": "error",
		"error":  err.Error(),
	}

	switch {
	case errors.Is(err, entityError.ErrInvalidPayload),
		errors.Is(err, entityError.ErrMissingRequiredField),
		errors.Is(err, entityError.ErrInvalidValue):
		response["code"] = "VALIDATION_ERROR"
		_ = conn.WriteJSON(response)

	case errors.Is(err, entityError.ErrDroneNotFound):
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "drone not registered"),
			time.Now().Add(time.Second),
		)

	default:
		response["code"] = "INTERNAL_ERROR"
		response["error"] = "internal server error"
		_ = conn.WriteJSON(response)
	}
}
