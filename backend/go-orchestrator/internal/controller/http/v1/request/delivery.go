package request

type UpdateDeliveryStatus struct {
	Status string `json:"status" binding:"required"`
}

type ConfirmGoodsLoadedRequest struct {
	OrderID      string `json:"order_id" binding:"required"`
	LockerCellID string `json:"locker_cell_id" binding:"required"`
	Timestamp    string `json:"timestamp"`
}
