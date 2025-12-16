package request

type CreateDroneRequest struct {
	Model     string `json:"model" binding:"required"`
	IPAddress string `json:"ip_address"`
}

type UpdateDroneRequest struct {
	Model     string `json:"model" binding:"required"`
	IPAddress string `json:"ip_address"`
}

type UpdateDroneStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
