package response

type Error struct {
	Error string `json:"error" example:"internal server error"`
}

type Health struct {
	Status  string `json:"status" example:"healthy"`
	Service string `json:"service" example:"drone-service"`
}

type DroneStatus struct {
	DroneID           string   `json:"drone_id" example:"drone-001"`
	Status            string   `json:"status" example:"idle"`
	BatteryLevel      float64  `json:"battery_level" example:"85.5"`
	Position          Position `json:"position"`
	Speed             float64  `json:"speed" example:"15.3"`
	CurrentDeliveryID *string  `json:"current_delivery_id" example:"delivery-123"`
	ErrorMessage      string   `json:"error_message,omitempty" example:""`
}

type Position struct {
	Latitude  float64 `json:"latitude" example:"55.751244"`
	Longitude float64 `json:"longitude" example:"37.618423"`
	Altitude  float64 `json:"altitude" example:"150.0"`
}

type CommandResponse struct {
	Status  string `json:"status" example:"command_sent"`
	Message string `json:"message" example:"Command successfully queued for delivery"`
}
