package request

type SendCommand struct {
	Command string         `json:"command" binding:"required" example:"start_delivery"`
	Payload map[string]any `json:"payload,omitempty"`
}
