package grpc

import (
	"context"
	"fmt"

	"github.com/skr1ms/SkyPostDelivery/drone-service/config"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrchestratorGRPCClient interface {
	RequestCellOpen(ctx context.Context, orderID string, parcelAutomatID string) (*CellOpenResponse, error)
}

type OrchestratorClient struct {
	conn   *grpc.ClientConn
	client OrchestratorGRPCClient
	logger logger.Interface
}

func NewOrchestratorGRPCClient(cfg *config.OrchestratorGRPC, log logger.Interface) (*OrchestratorClient, error) {
	conn, err := grpc.NewClient(cfg.URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("OrchestratorGRPCClient - NewOrchestratorGRPCClient - NewClient: %w", err)
	}

	return &OrchestratorClient{
		conn: conn,
		client: &OrchestratorClient{
			conn: conn,
		},
		logger: log,
	}, nil
}

func (c *OrchestratorClient) UpdateDeliveryStatus(ctx context.Context, deliveryID string, status string) error {
	c.logger.Info("Updating delivery status", nil, map[string]any{
		"delivery_id": deliveryID,
		"status":      status,
	})
	return nil
}

func (c *OrchestratorClient) RequestCellOpen(ctx context.Context, orderID string, parcelAutomatID string) (*CellOpenResponse, error) {
	c.logger.Info("Requesting cell open", nil, map[string]any{
		"order_id":          orderID,
		"parcel_automat_id": parcelAutomatID,
	})
	return &CellOpenResponse{
		Success:        true,
		Message:        "Cell open request sent",
		CellID:         "",
		InternalCellID: "",
	}, nil
}

func (c *OrchestratorClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
