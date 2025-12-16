package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DroneClient struct {
	conn *grpc.ClientConn
}

func NewDroneClient(address string) (*DroneClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("DroneClient - NewDroneClient - NewClient: %w", err)
	}

	return &DroneClient{conn: conn}, nil
}

func (c *DroneClient) Close() error {
	return c.conn.Close()
}

func (c *DroneClient) StartDelivery(ctx context.Context, orderID, goodID, lockerCellID, parcelAutomatID uuid.UUID, weight, height, length, width float64) (map[string]any, error) {
	return map[string]any{
		"success":     true,
		"delivery_id": orderID,
		"message":     "Delivery started",
	}, nil
}

func (c *DroneClient) GetStatus(ctx context.Context, droneID uuid.UUID) (map[string]any, error) {
	return map[string]any{
		"drone_id":      droneID,
		"status":        "idle",
		"battery_level": 85.5,
		"position": map[string]float64{
			"latitude":  55.7558,
			"longitude": 37.6173,
			"altitude":  0.0,
		},
	}, nil
}

func (c *DroneClient) CancelDelivery(ctx context.Context, deliveryID uuid.UUID) (map[string]any, error) {
	return map[string]any{
		"success": true,
		"message": "Delivery cancelled",
	}, nil
}
