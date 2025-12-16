package server

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase"
	pb "github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/pb"
)

type OrchestratorServer struct {
	pb.UnimplementedOrchestratorServiceServer
	deliveryUseCase      *usecase.DeliveryUseCase
	parcelAutomatUseCase *usecase.ParcelAutomatUseCase
}

func NewOrchestratorServer(deliveryUseCase *usecase.DeliveryUseCase, parcelAutomatUseCase *usecase.ParcelAutomatUseCase) *OrchestratorServer {
	return &OrchestratorServer{
		deliveryUseCase:      deliveryUseCase,
		parcelAutomatUseCase: parcelAutomatUseCase,
	}
}

func (s *OrchestratorServer) RequestCellOpen(ctx context.Context, req *pb.CellOpenRequest) (*pb.CellOpenResponse, error) {
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return &pb.CellOpenResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid order ID: %v", err),
		}, nil
	}

	parcelAutomatID, err := uuid.Parse(req.ParcelAutomatId)
	if err != nil {
		return &pb.CellOpenResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid parcel automat ID: %v", err),
		}, nil
	}

	cellID, internalDoorID, err := s.parcelAutomatUseCase.PrepareCell(ctx, orderID, parcelAutomatID)
	if err != nil {
		return &pb.CellOpenResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to prepare cell: %v", err),
		}, nil
	}

	return &pb.CellOpenResponse{
		Success: true,
		Message: "Cell opened successfully",
		CellId:  cellID.String(),
		InternalCellId: func() string {
			if internalDoorID == nil {
				return ""
			}
			return internalDoorID.String()
		}(),
	}, nil
}
