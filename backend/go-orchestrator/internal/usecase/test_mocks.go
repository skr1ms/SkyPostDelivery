package usecase

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/skr1ms/hitech-ekb/internal/entity"
	"github.com/skr1ms/hitech-ekb/pkg/qr"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, name, email, phoneNumber, passHash, role string) (*entity.User, error) {
	args := m.Called(ctx, name, email, phoneNumber, passHash, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) List(ctx context.Context) ([]*entity.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passHash string) (*entity.User, error) {
	args := m.Called(ctx, id, passHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) UpdateVerificationCode(ctx context.Context, id uuid.UUID, code string, expiresAt time.Time) (*entity.User, error) {
	args := m.Called(ctx, id, code, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) VerifyPhone(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) UpdateQR(ctx context.Context, id uuid.UUID, issuedAt, expiresAt time.Time) (*entity.User, error) {
	args := m.Called(ctx, id, issuedAt, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) Create(ctx context.Context, userID, goodID, parcelAutomatID uuid.UUID, status string) (*entity.Order, error) {
	args := m.Called(ctx, userID, goodID, parcelAutomatID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Order), args.Error(1)
}

func (m *MockOrderRepo) ListByUserIDWithGoods(ctx context.Context, userID uuid.UUID) ([]struct {
	Order *entity.Order
	Good  *entity.Good
}, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]struct {
		Order *entity.Order
		Good  *entity.Good
	}), args.Error(1)
}

func (m *MockOrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*entity.Order, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

type MockGoodRepo struct {
	mock.Mock
}

func (m *MockGoodRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Good, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Good), args.Error(1)
}

func (m *MockGoodRepo) Create(ctx context.Context, name string, weight, height, length, width float64, quantity int) (*entity.Good, error) {
	args := m.Called(ctx, name, weight, height, length, width, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Good), args.Error(1)
}

func (m *MockGoodRepo) List(ctx context.Context) ([]*entity.Good, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Good), args.Error(1)
}

func (m *MockGoodRepo) ListAvailable(ctx context.Context) ([]*entity.Good, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Good), args.Error(1)
}

func (m *MockGoodRepo) Update(ctx context.Context, id uuid.UUID, name string, weight, height, length, width float64) (*entity.Good, error) {
	args := m.Called(ctx, id, name, weight, height, length, width)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Good), args.Error(1)
}

func (m *MockGoodRepo) UpdateQuantity(ctx context.Context, id uuid.UUID, delta int) (*entity.Good, error) {
	args := m.Called(ctx, id, delta)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Good), args.Error(1)
}

func (m *MockGoodRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockLockerRepo struct {
	mock.Mock
}

func (m *MockLockerRepo) Create(ctx context.Context, postID uuid.UUID, height, length, width float64) (*entity.LockerCell, error) {
	args := m.Called(ctx, postID, height, length, width)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LockerCell), args.Error(1)
}

func (m *MockLockerRepo) FindAvailableCell(ctx context.Context, height, length, width float64) (*entity.LockerCell, error) {
	args := m.Called(ctx, height, length, width)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LockerCell), args.Error(1)
}

func (m *MockLockerRepo) UpdateCellStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockLockerRepo) UpdateDimensions(ctx context.Context, id uuid.UUID, height, length, width float64) (*entity.LockerCell, error) {
	args := m.Called(ctx, id, height, length, width)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LockerCell), args.Error(1)
}

func (m *MockLockerRepo) CreateCell(ctx context.Context, postID uuid.UUID, height, length, width float64, status string) (*entity.LockerCell, error) {
	args := m.Called(ctx, postID, height, length, width, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LockerCell), args.Error(1)
}

func (m *MockLockerRepo) GetCellByID(ctx context.Context, id uuid.UUID) (*entity.LockerCell, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LockerCell), args.Error(1)
}

func (m *MockLockerRepo) ListCellsByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.LockerCell, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.LockerCell), args.Error(1)
}

type MockDroneRepo struct {
	mock.Mock
}

func (m *MockDroneRepo) GetAvailable(ctx context.Context) (*entity.Drone, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Drone), args.Error(1)
}

func (m *MockDroneRepo) Create(ctx context.Context, model, status, ipAddress string) (*entity.Drone, error) {
	args := m.Called(ctx, model, status, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Drone), args.Error(1)
}

func (m *MockDroneRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Drone, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Drone), args.Error(1)
}

func (m *MockDroneRepo) List(ctx context.Context) ([]*entity.Drone, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Drone), args.Error(1)
}

func (m *MockDroneRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockDroneRepo) Update(ctx context.Context, id uuid.UUID, model, ipAddress string) (*entity.Drone, error) {
	args := m.Called(ctx, id, model, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Drone), args.Error(1)
}

func (m *MockDroneRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockDeliveryRepo struct {
	mock.Mock
}

func (m *MockDeliveryRepo) Create(ctx context.Context, orderID uuid.UUID, droneID *uuid.UUID, parcelAutomatID uuid.UUID, status string) (*entity.Delivery, error) {
	args := m.Called(ctx, orderID, droneID, parcelAutomatID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Delivery), args.Error(1)
}

func (m *MockDeliveryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Delivery, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Delivery), args.Error(1)
}

func (m *MockDeliveryRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Delivery, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Delivery), args.Error(1)
}

func (m *MockDeliveryRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*entity.Delivery, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Delivery), args.Error(1)
}

func (m *MockDeliveryRepo) ListByStatus(ctx context.Context, status string) ([]*entity.Delivery, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Delivery), args.Error(1)
}

func (m *MockDeliveryRepo) UpdateDrone(ctx context.Context, id uuid.UUID, droneID uuid.UUID) error {
	args := m.Called(ctx, id, droneID)
	return args.Error(0)
}

type MockParcelAutomatRepo struct {
	mock.Mock
}

func (m *MockParcelAutomatRepo) Create(ctx context.Context, city, address string, numberOfCells int, ipAddress, coordinates string, arucoID int, isWorking bool) (*entity.ParcelAutomat, error) {
	args := m.Called(ctx, city, address, numberOfCells, ipAddress, coordinates, arucoID, isWorking)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ParcelAutomat), args.Error(1)
}

func (m *MockParcelAutomatRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ParcelAutomat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ParcelAutomat), args.Error(1)
}

func (m *MockParcelAutomatRepo) List(ctx context.Context) ([]*entity.ParcelAutomat, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ParcelAutomat), args.Error(1)
}

func (m *MockParcelAutomatRepo) Update(ctx context.Context, id uuid.UUID, city, address, ipAddress, coordinates string) (*entity.ParcelAutomat, error) {
	args := m.Called(ctx, id, city, address, ipAddress, coordinates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ParcelAutomat), args.Error(1)
}

func (m *MockParcelAutomatRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockParcelAutomatRepo) ListWorking(ctx context.Context) ([]*entity.ParcelAutomat, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ParcelAutomat), args.Error(1)
}

func (m *MockParcelAutomatRepo) UpdateStatus(ctx context.Context, id uuid.UUID, isWorking bool) (*entity.ParcelAutomat, error) {
	args := m.Called(ctx, id, isWorking)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ParcelAutomat), args.Error(1)
}

type MockNotificationPkg struct {
	mock.Mock
}

func (m *MockNotificationPkg) Broadcast(message map[string]any) {
	m.Called(message)
}

type MockNotification = MockNotificationPkg

type MockQRWebAPI struct {
	mock.Mock
}

func (m *MockQRWebAPI) GenerateQR(ctx context.Context, userID uuid.UUID, email, name string) (string, error) {
	args := m.Called(ctx, userID, email, name)
	return args.String(0), args.Error(1)
}

func (m *MockQRWebAPI) ValidateQR(ctx context.Context, qrData string) (uuid.UUID, error) {
	args := m.Called(ctx, qrData)
	if args.Get(0) == nil {
		return uuid.Nil, args.Error(1)
	}
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockQRWebAPI) RefreshQR(ctx context.Context, userID uuid.UUID, email, name string) (string, error) {
	args := m.Called(ctx, userID, email, name)
	return args.String(0), args.Error(1)
}

type MockSMSWebAPI struct {
	mock.Mock
}

func (m *MockSMSWebAPI) SendSMS(ctx context.Context, phone, code string) error {
	args := m.Called(ctx, phone, code)
	return args.Error(0)
}

func (m *MockSMSWebAPI) CheckBalance(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}

type MockOrangePIWebAPI struct {
	mock.Mock
}

func (m *MockOrangePIWebAPI) SendCellUUIDs(ctx context.Context, ipAddress string, parcelAutomatID uuid.UUID, cellUUIDs []uuid.UUID) error {
	args := m.Called(ctx, ipAddress, parcelAutomatID, cellUUIDs)
	return args.Error(0)
}

func (m *MockOrangePIWebAPI) OpenCell(ctx context.Context, ipAddress string, cellID uuid.UUID) error {
	args := m.Called(ctx, ipAddress, cellID)
	return args.Error(0)
}

type MockQRGenerator struct {
	mock.Mock
}

func (m *MockQRGenerator) GenerateQRCode(userID uuid.UUID, email, name string) (*qr.QRData, string, error) {
	args := m.Called(userID, email, name)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*qr.QRData), args.String(1), args.Error(2)
}

func (m *MockQRGenerator) ValidateQRCode(qrDataJSON string) (*qr.QRData, error) {
	args := m.Called(qrDataJSON)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*qr.QRData), args.Error(1)
}

func (m *MockQRGenerator) RefreshQRCode(userID uuid.UUID, email, name string) (*qr.QRData, string, error) {
	args := m.Called(userID, email, name)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*qr.QRData), args.String(1), args.Error(2)
}

type MockGoodInstanceRepo struct {
	mock.Mock
}

func (m *MockGoodInstanceRepo) Create(ctx context.Context, goodID uuid.UUID, status string, storageLocation *string) (*entity.GoodInstance, error) {
	args := m.Called(ctx, goodID, status, storageLocation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GoodInstance), args.Error(1)
}

func (m *MockGoodInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.GoodInstance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GoodInstance), args.Error(1)
}

func (m *MockGoodInstanceRepo) FindAvailable(ctx context.Context, goodID uuid.UUID) (*entity.GoodInstance, error) {
	args := m.Called(ctx, goodID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GoodInstance), args.Error(1)
}

func (m *MockGoodInstanceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*entity.GoodInstance, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GoodInstance), args.Error(1)
}

func (m *MockGoodInstanceRepo) CountAvailable(ctx context.Context, goodID uuid.UUID) (int, error) {
	args := m.Called(ctx, goodID)
	return args.Int(0), args.Error(1)
}

func (m *MockGoodInstanceRepo) ListByGoodID(ctx context.Context, goodID uuid.UUID) ([]*entity.GoodInstance, error) {
	args := m.Called(ctx, goodID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GoodInstance), args.Error(1)
}

func (m *MockGoodInstanceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockMinioClient struct {
	mock.Mock
}

func (m *MockMinioClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	args := m.Called(ctx, objectName, reader, objectSize, contentType)
	return args.Error(0)
}

func (m *MockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, bucketName, objectName, reader, objectSize, opts)
	return args.Get(0).(minio.UploadInfo), args.Error(1)
}

type MockRabbitMQClient struct {
	mock.Mock
}

func (m *MockRabbitMQClient) Publish(ctx context.Context, queueName string, message interface{}) error {
	args := m.Called(ctx, queueName, message)
	return args.Error(0)
}

func (m *MockRabbitMQClient) Consume(queueName string, handler func([]byte) error) error {
	args := m.Called(queueName, handler)
	return args.Error(0)
}
