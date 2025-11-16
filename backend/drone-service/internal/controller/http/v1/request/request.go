package request

type SendCommand struct {
	Command string                 `json:"command" binding:"required" example:"start_delivery"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}
