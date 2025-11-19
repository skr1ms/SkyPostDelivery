package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/persistent/sqlc"
)

type UserRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db, q: sqlc.New(db)}
}

func toEntityUser(u sqlc.User) *entity.User {
	var codeExpiresAt, qrIssuedAt, qrExpiresAt *time.Time

	if u.CodeExpiresAt.Valid {
		t := u.CodeExpiresAt.Time
		codeExpiresAt = &t
	}
	if u.QrIssuedAt.Valid {
		t := u.QrIssuedAt.Time
		qrIssuedAt = &t
	}
	if u.QrExpiresAt.Valid {
		t := u.QrExpiresAt.Time
		qrExpiresAt = &t
	}

	return &entity.User{
		ID:               u.ID,
		FullName:         u.FullName,
		Email:            u.Email,
		PhoneNumber:      u.PhoneNumber,
		PassHash:         u.PassHash,
		PhoneVerified:    u.PhoneVerified,
		VerificationCode: u.VerificationCode,
		CodeExpiresAt:    codeExpiresAt,
		CreatedAt:        u.CreatedAt.Time,
		Role:             u.Role,
		QRIssuedAt:       qrIssuedAt,
		QRExpiresAt:      qrExpiresAt,
	}
}

func (r *UserRepo) Create(ctx context.Context, fullName, email, phoneNumber, passHash, role string) (*entity.User, error) {
	var emailPtr, phonePtr, passHashPtr *string

	if email != "" {
		emailPtr = &email
	}
	if phoneNumber != "" {
		phonePtr = &phoneNumber
	}
	if passHash != "" {
		passHashPtr = &passHash
	}

	u, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    fullName,
		Email:       emailPtr,
		PhoneNumber: phonePtr,
		PassHash:    passHashPtr,
		Role:        role,
	})
	if err != nil {
		return nil, fmt.Errorf("UserRepo - Create - q.CreateUser: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) CreateWithCustomDate(ctx context.Context, fullName, email, phoneNumber, passHash, role string, createdAt time.Time) (*entity.User, error) {
	var emailPtr, phonePtr, passHashPtr *string

	if email != "" {
		emailPtr = &email
	}
	if phoneNumber != "" {
		phonePtr = &phoneNumber
	}
	if passHash != "" {
		passHashPtr = &passHash
	}

	u, err := r.q.CreateUserWithCustomDate(ctx, sqlc.CreateUserWithCustomDateParams{
		FullName:    fullName,
		Email:       emailPtr,
		PhoneNumber: phonePtr,
		PassHash:    passHashPtr,
		Role:        role,
		CreatedAt:   pgtype.Timestamp{Time: createdAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("UserRepo - CreateWithCustomDate - q.CreateUserWithCustomDate: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetByID - q.GetUserByID: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	u, err := r.q.GetUserByEmail(ctx, &email)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetByEmail - q.GetUserByEmail: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	u, err := r.q.GetUserByPhone(ctx, &phone)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetByPhone - q.GetUserByPhone: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) List(ctx context.Context) ([]*entity.User, error) {
	rows, err := r.q.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - List - q.ListUsers: %w", err)
	}
	users := make([]*entity.User, 0, len(rows))
	for _, u := range rows {
		users = append(users, toEntityUser(u))
	}
	return users, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passHash string) (*entity.User, error) {
	u, err := r.q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:       id,
		PassHash: &passHash,
	})
	if err != nil {
		return nil, fmt.Errorf("UserRepo - UpdatePassword - q.UpdateUserPassword: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) UpdateVerificationCode(ctx context.Context, id uuid.UUID, code string, expiresAt time.Time) (*entity.User, error) {
	var codePtr *string
	if code != "" {
		codePtr = &code
	}

	u, err := r.q.UpdateUserVerificationCode(ctx, sqlc.UpdateUserVerificationCodeParams{
		ID:               id,
		VerificationCode: codePtr,
		CodeExpiresAt:    pgtype.Timestamp{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("UserRepo - UpdateVerificationCode - q.UpdateVerificationCode: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) VerifyPhone(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	u, err := r.q.VerifyUserPhone(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - VerifyPhone - q.VerifyUserPhone: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) UpdateQR(ctx context.Context, id uuid.UUID, issuedAt, expiresAt time.Time) (*entity.User, error) {
	u, err := r.q.UpdateUserQR(ctx, sqlc.UpdateUserQRParams{
		ID:          id,
		QrIssuedAt:  pgtype.Timestamp{Time: issuedAt, Valid: true},
		QrExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("UserRepo - UpdateQR - q.UpdateUserQR: %w", err)
	}
	return toEntityUser(u), nil
}
