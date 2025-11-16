package grpc

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrchestratorClient struct {
	conn   *grpc.ClientConn
	client interface{}
}

func NewOrchestratorClient(address string) (*OrchestratorClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	return &OrchestratorClient{
		conn: conn,
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
