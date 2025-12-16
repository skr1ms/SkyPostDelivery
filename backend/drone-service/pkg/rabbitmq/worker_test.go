package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDeliveryHandler struct {
	mock.Mock
}

func (m *mockDeliveryHandler) ExecuteDelivery(
	ctx context.Context,
	droneID string,
	orderID string,
	goodID string,
	parcelAutomatID string,
	arucoID int,
	coordinates string,
	weight float64,
	height float64,
	length float64,
	width float64,
	internalLockerCellID *string,
) error {
	args := m.Called(ctx, droneID, orderID, goodID, parcelAutomatID, arucoID, coordinates, weight, height, length, width, internalLockerCellID)
	return args.Error(0)
}

func (m *mockDeliveryHandler) HandleReturnTask(ctx context.Context, droneID string, deliveryID string, baseMarkerID int) error {
	args := m.Called(ctx, droneID, deliveryID, baseMarkerID)
	return args.Error(0)
}

func TestDeliveryWorker_handleDeliveryTask_Success(t *testing.T) {
	mockHandler := new(mockDeliveryHandler)
	client := &Client{}
	logger := logger.New("test")
	worker := NewDeliveryWorker(client, mockHandler, logger)

	ctx := context.Background()
	message := map[string]any{
		"drone_id":                "drone-123",
		"order_id":                "order-456",
		"good_id":                 "good-789",
		"parcel_automat_id":       "automat-456",
		"aruco_id":                131,
		"coordinates":             "55.7558,37.6173",
		"weight":                  1.5,
		"height":                  10.0,
		"length":                  20.0,
		"width":                   15.0,
		"internal_locker_cell_id": "internal-cell-123",
	}

	body, _ := json.Marshal(message)
	delivery := amqp.Delivery{Body: body}

	mockHandler.On("ExecuteDelivery",
		ctx,
		"drone-123",
		"order-456",
		"good-789",
		"automat-456",
		131,
		"55.7558,37.6173",
		1.5,
		10.0,
		20.0,
		15.0,
		mock.MatchedBy(func(id *string) bool {
			return id != nil && *id == "internal-cell-123"
		}),
	).Return(nil)

	err := worker.handleDeliveryTask(ctx, delivery)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestDeliveryWorker_handleDeliveryTask_InvalidJSON(t *testing.T) {
	mockHandler := new(mockDeliveryHandler)
	client := &Client{}
	logger := logger.New("test")
	worker := NewDeliveryWorker(client, mockHandler, logger)

	ctx := context.Background()
	delivery := amqp.Delivery{Body: []byte("invalid json")}

	err := worker.handleDeliveryTask(ctx, delivery)

	assert.Error(t, err)
	mockHandler.AssertNotCalled(t, "ExecuteDelivery")
}

func TestDeliveryWorker_handleDeliveryTask_HandlerError(t *testing.T) {
	mockHandler := new(mockDeliveryHandler)
	client := &Client{}
	logger := logger.New("test")
	worker := NewDeliveryWorker(client, mockHandler, logger)

	ctx := context.Background()
	message := map[string]any{
		"drone_id":          "drone-123",
		"order_id":          "order-456",
		"good_id":           "good-789",
		"parcel_automat_id": "automat-456",
		"aruco_id":          131,
		"coordinates":       "55.7558,37.6173",
		"weight":            1.5,
		"height":            10.0,
		"length":            20.0,
		"width":             15.0,
	}

	body, _ := json.Marshal(message)
	delivery := amqp.Delivery{Body: body}

	mockHandler.On("ExecuteDelivery", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("handler error"))

	err := worker.handleDeliveryTask(ctx, delivery)

	assert.Error(t, err)
	mockHandler.AssertExpectations(t)
}

func TestDeliveryWorker_handleReturnTask_Success(t *testing.T) {
	mockHandler := new(mockDeliveryHandler)
	client := &Client{}
	logger := logger.New("test")
	worker := NewDeliveryWorker(client, mockHandler, logger)

	ctx := context.Background()
	message := map[string]any{
		"drone_id":    "drone-123",
		"delivery_id": "delivery-456",
		"aruco_id":    131,
	}

	body, _ := json.Marshal(message)
	delivery := amqp.Delivery{Body: body}

	mockHandler.On("HandleReturnTask", ctx, "drone-123", "delivery-456", 131).Return(nil)

	err := worker.handleReturnTask(ctx, delivery)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestDeliveryWorker_handleReturnTask_DefaultArucoID(t *testing.T) {
	mockHandler := new(mockDeliveryHandler)
	client := &Client{}
	logger := logger.New("test")
	worker := NewDeliveryWorker(client, mockHandler, logger)

	ctx := context.Background()
	message := map[string]any{
		"drone_id":    "drone-123",
		"delivery_id": "delivery-456",
	}

	body, _ := json.Marshal(message)
	delivery := amqp.Delivery{Body: body}

	mockHandler.On("HandleReturnTask", ctx, "drone-123", "delivery-456", 131).Return(nil)

	err := worker.handleReturnTask(ctx, delivery)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestDeliveryWorker_handleReturnTask_IntArucoID(t *testing.T) {
	mockHandler := new(mockDeliveryHandler)
	client := &Client{}
	logger := logger.New("test")
	worker := NewDeliveryWorker(client, mockHandler, logger)

	ctx := context.Background()
	message := map[string]any{
		"drone_id":    "drone-123",
		"delivery_id": "delivery-456",
		"aruco_id":    200,
	}

	body, _ := json.Marshal(message)
	delivery := amqp.Delivery{Body: body}

	mockHandler.On("HandleReturnTask", ctx, "drone-123", "delivery-456", 200).Return(nil)

	err := worker.handleReturnTask(ctx, delivery)

	assert.NoError(t, err)
	mockHandler.AssertExpectations(t)
}

func TestDeliveryWorker_handleReturnTask_InvalidJSON(t *testing.T) {
	mockHandler := new(mockDeliveryHandler)
	client := &Client{}
	logger := logger.New("test")
	worker := NewDeliveryWorker(client, mockHandler, logger)

	ctx := context.Background()
	delivery := amqp.Delivery{Body: []byte("invalid json")}

	err := worker.handleReturnTask(ctx, delivery)

	assert.Error(t, err)
	mockHandler.AssertNotCalled(t, "HandleReturnTask")
}

func TestDeliveryWorker_handleReturnTask_HandlerError(t *testing.T) {
	mockHandler := new(mockDeliveryHandler)
	client := &Client{}
	logger := logger.New("test")
	worker := NewDeliveryWorker(client, mockHandler, logger)

	ctx := context.Background()
	message := map[string]any{
		"drone_id":    "drone-123",
		"delivery_id": "delivery-456",
		"aruco_id":    131,
	}

	body, _ := json.Marshal(message)
	delivery := amqp.Delivery{Body: body}

	mockHandler.On("HandleReturnTask", ctx, "drone-123", "delivery-456", 131).Return(errors.New("handler error"))

	err := worker.handleReturnTask(ctx, delivery)

	assert.Error(t, err)
	mockHandler.AssertExpectations(t)
}
