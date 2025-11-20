package grpc

import (
	"context"
	"fmt"
	"log"

	"github.com/skr1ms/SkyPostDelivery/drone-service/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrchestratorGRPCClient interface {
	RequestCellOpen(ctx context.Context, orderID string, parcelAutomatID string) (*CellOpenResponse, error)
}

type OrchestratorClient struct {
	conn   *grpc.ClientConn
	client OrchestratorGRPCClient
}

func NewOrchestratorGRPCClient(cfg *config.OrchestratorGRPC) (*OrchestratorClient, error) {
	conn, err := grpc.NewClient(cfg.URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	return &OrchestratorClient{
		conn: conn,
		client: &OrchestratorClient{
			conn: conn,
		},
	}, nil
}

func (c *OrchestratorClient) UpdateDeliveryStatus(ctx context.Context, deliveryID string, status string) error {
	log.Printf("Updating delivery status: %s -> %s", deliveryID, status)
	return nil
}

func (c *OrchestratorClient) RequestCellOpen(ctx context.Context, orderID string, parcelAutomatID string) (*CellOpenResponse, error) {
	log.Printf("Requesting cell open for order %s at automat %s", orderID, parcelAutomatID)
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
