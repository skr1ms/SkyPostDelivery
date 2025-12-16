package response

import (
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
)

type UserRegister struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    *string   `json:"email,omitempty"`
	Phone    *string   `json:"phone,omitempty"`
	Role     string    `json:"role"`
}

type Login struct {
	User             *entity.User `json:"user"`
	AccessToken      string       `json:"access_token"`
	RefreshToken     string       `json:"refresh_token"`
	AccessExpiresAt  int64        `json:"access_expires_at"`
	RefreshExpiresAt int64        `json:"refresh_expires_at"`
	QRCode           string       `json:"qr_code"`
}

type RefreshToken struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	AccessExpiresAt  int64  `json:"access_expires_at"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
}

type PhoneCodeSent struct {
	Message string `json:"message"`
}

type SuccessMessage struct {
	Message string `json:"message"`
}

type UserWithQR struct {
	User   *entity.User `json:"user"`
	QRCode string       `json:"qr_code"`
}
