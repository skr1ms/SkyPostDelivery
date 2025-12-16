package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
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

func (r *UserRepo) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	u, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		PassHash:    user.PassHash,
		Role:        user.Role,
	})
	if err != nil {
		if isPgUniqueViolation(err) {
			return nil, entityError.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("UserRepo - Create: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) CreateWithCustomDate(ctx context.Context, user *entity.User, createdAt time.Time) (*entity.User, error) {
	u, err := r.q.CreateUserWithCustomDate(ctx, sqlc.CreateUserWithCustomDateParams{
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		PassHash:    user.PassHash,
		Role:        user.Role,
		CreatedAt:   pgtype.Timestamp{Time: createdAt, Valid: true},
	})
	if err != nil {
		if isPgUniqueViolation(err) {
			return nil, entityError.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("UserRepo - CreateWithCustomDate: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepo - GetByID: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	u, err := r.q.GetUserByEmail(ctx, &email)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrUserNotFoundByEmail
		}
		return nil, fmt.Errorf("UserRepo - GetByEmail: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	u, err := r.q.GetUserByPhone(ctx, &phone)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrUserNotFoundByPhone
		}
		return nil, fmt.Errorf("UserRepo - GetByPhone: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) List(ctx context.Context) ([]*entity.User, error) {
	rows, err := r.q.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - List: %w", err)
	}
	users := make([]*entity.User, 0, len(rows))
	for _, u := range rows {
		users = append(users, toEntityUser(u))
	}
	return users, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, user *entity.User) (*entity.User, error) {
	u, err := r.q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:       user.ID,
		PassHash: user.PassHash,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepo - UpdatePassword: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) UpdateVerificationCode(ctx context.Context, user *entity.User) (*entity.User, error) {
	var codeExpiresAt pgtype.Timestamp
	if user.CodeExpiresAt != nil {
		codeExpiresAt = pgtype.Timestamp{Time: *user.CodeExpiresAt, Valid: true}
	} else {
		codeExpiresAt = pgtype.Timestamp{Valid: false}
	}

	u, err := r.q.UpdateUserVerificationCode(ctx, sqlc.UpdateUserVerificationCodeParams{
		ID:               user.ID,
		VerificationCode: user.VerificationCode,
		CodeExpiresAt:    codeExpiresAt,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepo - UpdateVerificationCode: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) VerifyPhone(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	u, err := r.q.VerifyUserPhone(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepo - VerifyPhone: %w", err)
	}
	return toEntityUser(u), nil
}

func (r *UserRepo) UpdateQR(ctx context.Context, user *entity.User) (*entity.User, error) {
	var qrIssuedAt, qrExpiresAt pgtype.Timestamp

	if user.QRIssuedAt != nil {
		qrIssuedAt = pgtype.Timestamp{Time: *user.QRIssuedAt, Valid: true}
	} else {
		qrIssuedAt = pgtype.Timestamp{Valid: false}
	}

	if user.QRExpiresAt != nil {
		qrExpiresAt = pgtype.Timestamp{Time: *user.QRExpiresAt, Valid: true}
	} else {
		qrExpiresAt = pgtype.Timestamp{Valid: false}
	}

	u, err := r.q.UpdateUserQR(ctx, sqlc.UpdateUserQRParams{
		ID:          user.ID,
		QrIssuedAt:  qrIssuedAt,
		QrExpiresAt: qrExpiresAt,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepo - UpdateQR: %w", err)
	}
	return toEntityUser(u), nil
}
