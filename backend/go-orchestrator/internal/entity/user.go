package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	FullName         string     `json:"full_name"`
	Email            *string    `json:"email,omitempty"`
	PhoneNumber      *string    `json:"phone_number,omitempty"`
	PassHash         *string    `json:"-"`
	PhoneVerified    bool       `json:"phone_verified"`
	VerificationCode *string    `json:"-"`
	CodeExpiresAt    *time.Time `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	Role             string     `json:"role"`
	QRIssuedAt       *time.Time `json:"qr_issued_at,omitempty"`
	QRExpiresAt      *time.Time `json:"qr_expires_at,omitempty"`
}

func (u *User) GetEmail() string {
	if u.Email == nil {
		return ""
	}
	return *u.Email
}

func (u *User) IsCodeValid(code string) bool {
	if u.VerificationCode == nil {
		return false
	}
	return *u.VerificationCode == code
}

func (u *User) IsCodeExpired() bool {
	if u.CodeExpiresAt == nil {
		return true
	}
	return time.Now().After(*u.CodeExpiresAt)
}
