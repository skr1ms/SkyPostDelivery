package request

type CellDimensions struct {
	Height float64 `json:"height" binding:"required,gt=0"`
	Length float64 `json:"length" binding:"required,gt=0"`
	Width  float64 `json:"width" binding:"required,gt=0"`
}

type CreateParcelAutomatRequest struct {
	City          string           `json:"city" binding:"required"`
	Address       string           `json:"address" binding:"required"`
	IPAddress     string           `json:"ip_address"`
	Coordinates   string           `json:"coordinates"`
	ArucoID       int              `json:"aruco_id" binding:"required"`
	NumberOfCells int              `json:"number_of_cells" binding:"required"`
	Cells         []CellDimensions `json:"cells" binding:"required,dive"`
}

type UpdateParcelAutomatRequest struct {
	City        string `json:"city" binding:"required"`
	Address     string `json:"address" binding:"required"`
	IPAddress   string `json:"ip_address"`
	Coordinates string `json:"coordinates"`
}

type UpdateParcelAutomatStatusRequest struct {
	IsWorking bool `json:"is_working"`
}

type UpdateCellRequest struct {
	Height float64 `json:"height" binding:"required"`
	Length float64 `json:"length" binding:"required"`
	Width  float64 `json:"width" binding:"required"`
}

type QRScanRequest struct {
	QRData          string `json:"qr_data" binding:"required"`
	ParcelAutomatID string `json:"parcel_automat_id" binding:"required"`
}

type ConfirmPickupRequest struct {
	CellIDs []string `json:"cell_ids" binding:"required"`
}
