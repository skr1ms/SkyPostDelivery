package response

import (
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
)

type AutomatWithCells struct {
	*entity.ParcelAutomat
	Cells []*entity.LockerCell `json:"cells"`
}

type SystemStatus struct {
	Drones           []*entity.Drone    `json:"drones"`
	Automats         []AutomatWithCells `json:"automats"`
	ActiveDeliveries []*entity.Delivery `json:"active_deliveries"`
}

type DroneDetails struct {
	Drone           *entity.Drone    `json:"drone"`
	CurrentDelivery *DeliveryDetails `json:"current_delivery"`
}

type DeliveryDetails struct {
	DeliveryID           string `json:"delivery_id"`
	OrderID              string `json:"order_id"`
	Status               string `json:"status"`
	RecipientName        string `json:"recipient_name"`
	ParcelAutomatAddress string `json:"parcel_automat_address"`
	ArucoID              int    `json:"aruco_id"`
}
