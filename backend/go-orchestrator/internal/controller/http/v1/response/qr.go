package response

type ValidateResponse struct {
	Valid     bool   `json:"valid"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FullName  string `json:"name"`
	ExpiresAt int64  `json:"expires_at"`
}

type RefreshResponse struct {
	QRCode    string `json:"qr_code"`
	ExpiresAt int64  `json:"expires_at"`
}
