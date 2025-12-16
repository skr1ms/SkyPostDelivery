package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
)

type (
	UserRepo interface {
		Create(ctx context.Context, user *entity.User) (*entity.User, error)
		CreateWithCustomDate(ctx context.Context, user *entity.User, createdAt time.Time) (*entity.User, error)
		GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
		GetByEmail(ctx context.Context, email string) (*entity.User, error)
		GetByPhone(ctx context.Context, phone string) (*entity.User, error)
		List(ctx context.Context) ([]*entity.User, error)
		UpdatePassword(ctx context.Context, user *entity.User) (*entity.User, error)
		UpdateVerificationCode(ctx context.Context, user *entity.User) (*entity.User, error)
		VerifyPhone(ctx context.Context, id uuid.UUID) (*entity.User, error)
		UpdateQR(ctx context.Context, user *entity.User) (*entity.User, error)
	}

	DeviceRepo interface {
		Upsert(ctx context.Context, device *entity.Device) error
		ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Device, error)
		DeleteByToken(ctx context.Context, token string) error
	}

	GoodRepo interface {
		Create(ctx context.Context, good *entity.Good) (*entity.Good, error)
		GetByID(ctx context.Context, id uuid.UUID) (*entity.Good, error)
		List(ctx context.Context) ([]*entity.Good, error)
		ListAvailable(ctx context.Context) ([]*entity.Good, error)
		Update(ctx context.Context, good *entity.Good) (*entity.Good, error)
		UpdateQuantity(ctx context.Context, id uuid.UUID, delta int) (*entity.Good, error)
		Delete(ctx context.Context, id uuid.UUID) error
	}

	OrderRepo interface {
		Create(ctx context.Context, order *entity.Order) (*entity.Order, error)
		CreateWithCell(ctx context.Context, order *entity.Order) (*entity.Order, error)
		GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error)
		GetByLockerCellID(ctx context.Context, lockerCellID uuid.UUID) (*entity.Order, error)
		ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Order, error)
		ListByUserIDWithGoods(ctx context.Context, userID uuid.UUID) ([]struct {
			Order *entity.Order
			Good  *entity.Good
		}, error)
		UpdateStatus(ctx context.Context, order *entity.Order) (*entity.Order, error)
	}

	DroneRepo interface {
		Create(ctx context.Context, drone *entity.Drone) (*entity.Drone, error)
		GetByID(ctx context.Context, id uuid.UUID) (*entity.Drone, error)
		GetAvailable(ctx context.Context) (*entity.Drone, error)
		List(ctx context.Context) ([]*entity.Drone, error)
		Update(ctx context.Context, drone *entity.Drone) (*entity.Drone, error)
		UpdateStatus(ctx context.Context, drone *entity.Drone) error
		Delete(ctx context.Context, id uuid.UUID) error
	}

	ParcelAutomatRepo interface {
		Create(ctx context.Context, automat *entity.ParcelAutomat) (*entity.ParcelAutomat, error)
		GetByID(ctx context.Context, id uuid.UUID) (*entity.ParcelAutomat, error)
		List(ctx context.Context) ([]*entity.ParcelAutomat, error)
		ListWorking(ctx context.Context) ([]*entity.ParcelAutomat, error)
		Update(ctx context.Context, automat *entity.ParcelAutomat) (*entity.ParcelAutomat, error)
		UpdateStatus(ctx context.Context, automat *entity.ParcelAutomat) (*entity.ParcelAutomat, error)
		Delete(ctx context.Context, id uuid.UUID) error
	}

	DeliveryRepo interface {
		Create(ctx context.Context, delivery *entity.Delivery) (*entity.Delivery, error)
		GetByID(ctx context.Context, id uuid.UUID) (*entity.Delivery, error)
		GetByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Delivery, error)
		UpdateStatus(ctx context.Context, delivery *entity.Delivery) (*entity.Delivery, error)
		UpdateDrone(ctx context.Context, delivery *entity.Delivery) error
		ListByStatus(ctx context.Context, status string) ([]*entity.Delivery, error)
	}

	LockerRepo interface {
		Create(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error)
		CreateWithNumber(ctx context.Context, cell *entity.LockerCell, cellNumber int) (*entity.LockerCell, error)
		CreateCell(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error)
		GetCellByID(ctx context.Context, id uuid.UUID) (*entity.LockerCell, error)
		FindAvailableCell(ctx context.Context, height, length, width float64) (*entity.LockerCell, error)
		UpdateCellStatus(ctx context.Context, cell *entity.LockerCell) error
		UpdateDimensions(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error)
		ListCellsByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.LockerCell, error)
	}

	InternalLockerRepo interface {
		Create(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error)
		CreateWithNumber(ctx context.Context, cell *entity.LockerCell, cellNumber int) (*entity.LockerCell, error)
		CreateCell(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error)
		GetCellByID(ctx context.Context, id uuid.UUID) (*entity.LockerCell, error)
		FindAvailableCell(ctx context.Context, height, length, width float64) (*entity.LockerCell, error)
		UpdateCellStatus(ctx context.Context, cell *entity.LockerCell) error
		UpdateDimensions(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error)
		ListCellsByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.LockerCell, error)
	}

	SMSAeroWebAPI interface {
		SendSMS(ctx context.Context, phone, message string) error
		CheckBalance(ctx context.Context) (float64, error)
	}

	QRWebAPI interface {
		GenerateQR(ctx context.Context, userID uuid.UUID, email, name string) (string, error)
	}

	OrangePIWebAPI interface {
		SendCellUUIDs(ctx context.Context, ipAddress string, parcelAutomatID uuid.UUID, outCellUUIDs []uuid.UUID, internalCellUUIDs []uuid.UUID) error
		OpenCell(ctx context.Context, ipAddress string, cellID uuid.UUID) error
	}

	Sender interface {
		SendDeliveryNotification(ctx context.Context, tokens []string, orderID string, lockerCellID *string) ([]string, error)
	}
)
