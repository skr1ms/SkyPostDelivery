package response

type ValidateResponse struct {
	Valid            bool   `json:"valid"`
	UserID           string `json:"user_id"`
	Email            string `json:"email"`
	FullName         string `json:"name"`
	AccessExpiresAt  int64  `json:"access_expires_at"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
}

type RefreshResponse struct {
	QRCode           string `json:"qr_code"`
	IssuedAt         int64  `json:"issued_at"`
	AccessExpiresAt  int64  `json:"access_expires_at"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
}

type QRResponse struct {
	QRCode    string `json:"qr_code"`
	IssuedAt  int64  `json:"issued_at"`
	ExpiresAt int64  `json:"expires_at"`
}
