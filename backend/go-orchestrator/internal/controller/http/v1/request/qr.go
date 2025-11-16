package request

type ValidateRequest struct {
	QRData string `json:"qr_data" binding:"required"`
}
