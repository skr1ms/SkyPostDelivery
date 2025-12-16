package entity

type HeartbeatPayload struct {
	Status            string   `json:"status"`
	BatteryLevel      float64  `json:"battery_level"`
	Position          Position `json:"position"`
	Speed             float64  `json:"speed"`
	CurrentDeliveryID string   `json:"current_delivery_id,omitempty"`
	ErrorMessage      string   `json:"error_message,omitempty"`
}

type StatusUpdatePayload struct {
	Status            string   `json:"status"`
	BatteryLevel      float64  `json:"battery_level"`
	Position          Position `json:"position"`
	Speed             float64  `json:"speed"`
	CurrentDeliveryID string   `json:"current_delivery_id,omitempty"`
	ErrorMessage      string   `json:"error_message,omitempty"`
}

type DeliveryUpdatePayload struct {
	DroneStatus     string `json:"drone_status"`
	DeliveryID      string `json:"delivery_id,omitempty"`
	OrderID         string `json:"order_id,omitempty"`
	ParcelAutomatID string `json:"parcel_automat_id,omitempty"`
}

type ArrivedAtDestinationPayload struct {
	OrderID         string `json:"order_id"`
	ParcelAutomatID string `json:"parcel_automat_id"`
}

type CargoDroppedPayload struct {
	OrderID      string `json:"order_id"`
	LockerCellID string `json:"locker_cell_id,omitempty"`
}
