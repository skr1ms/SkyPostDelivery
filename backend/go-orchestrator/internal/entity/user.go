package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	FullName         string     `json:"full_name"`
	Email            *string    `json:"-"`
	PhoneNumber      *string    `json:"-"`
	PassHash         *string    `json:"-"`
	PhoneVerified    bool       `json:"phone_verified"`
	VerificationCode *string    `json:"-"`
	CodeExpiresAt    *time.Time `json:"-"`
	CreatedAt        time.Time  `json:"-"`
	Role             string     `json:"role"`
	QRIssuedAt       *time.Time `json:"-"`
	QRExpiresAt      *time.Time `json:"-"`
}

func (u *User) GetEmail() string {
	if u.Email == nil {
		return ""
	}
	return *u.Email
}

func (u *User) GetPhoneNumber() string {
	if u.PhoneNumber == nil {
		return ""
	}
	return *u.PhoneNumber
}

func (u *User) IsCodeValid(code string) bool {
	return u.VerificationCode != nil && *u.VerificationCode == code
}

func (u *User) IsCodeExpired() bool {
	if u.CodeExpiresAt == nil {
		return true
	}
	return time.Now().After(*u.CodeExpiresAt)
}

func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		Email       *string    `json:"email,omitempty"`
		PhoneNumber *string    `json:"phone_number,omitempty"`
		CreatedAt   time.Time  `json:"created_at"`
		QRIssuedAt  *time.Time `json:"qr_issued_at,omitempty"`
		QRExpiresAt *time.Time `json:"qr_expires_at,omitempty"`
		*Alias
	}{
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		CreatedAt:   u.CreatedAt,
		QRIssuedAt:  u.QRIssuedAt,
		QRExpiresAt: u.QRExpiresAt,
		Alias:       (*Alias)(&u),
	})
}
