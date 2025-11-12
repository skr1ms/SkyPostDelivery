package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/hitech-ekb/internal/controller/http/v1/request"
	"github.com/skr1ms/hitech-ekb/internal/entity"
)

type ParcelAutomatUseCase struct {
	parcelAutomatRepo  ParcelAutomatRepo
	lockerRepo         LockerRepo
	internalLockerRepo InternalLockerRepo
	orderRepo          OrderRepo
	deliveryRepo       DeliveryRepo
	qrUseCase          *QRUseCase
	orangePIWebAPI     OrangePIWebAPI
}

func NewParcelAutomatUseCase(
	parcelAutomatRepo ParcelAutomatRepo,
	lockerRepo LockerRepo,
	internalLockerRepo InternalLockerRepo,
	orderRepo OrderRepo,
	deliveryRepo DeliveryRepo,
	qrUseCase *QRUseCase,
	orangePIWebAPI OrangePIWebAPI,
) *ParcelAutomatUseCase {
	return &ParcelAutomatUseCase{
		parcelAutomatRepo:  parcelAutomatRepo,
		lockerRepo:         lockerRepo,
		internalLockerRepo: internalLockerRepo,
		orderRepo:          orderRepo,
		deliveryRepo:       deliveryRepo,
		qrUseCase:          qrUseCase,
		orangePIWebAPI:     orangePIWebAPI,
	}
}

const defaultInternalDoorCount = 3

func (uc *ParcelAutomatUseCase) Create(ctx context.Context, city, address, ipAddress, coordinates string, numberOfCells int, arucoID int, cells []request.CellDimensions) (*entity.ParcelAutomat, error) {
	automat, err := uc.parcelAutomatRepo.Create(ctx, city, address, numberOfCells, ipAddress, coordinates, arucoID, true)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - Create - parcelAutomatRepo.Create: %w", err)
	}

	cellUUIDs := make([]uuid.UUID, 0, len(cells))
	for i, cell := range cells {
		createdCell, err := uc.lockerRepo.CreateWithNumber(ctx, automat.ID, cell.Height, cell.Length, cell.Width, i+1)
		if err != nil {
			return nil, fmt.Errorf("parcel automat usecase - Create - lockerRepo.CreateWithNumber: %w", err)
		}
		cellUUIDs = append(cellUUIDs, createdCell.ID)
	}

	internalCellUUIDs := make([]uuid.UUID, 0, defaultInternalDoorCount)
	for i := 0; i < defaultInternalDoorCount; i++ {
		createdCell, err := uc.internalLockerRepo.CreateWithNumber(ctx, automat.ID, 0, 0, 0, i+1)
		if err != nil {
			return nil, fmt.Errorf("parcel automat usecase - Create - internalLockerRepo.CreateWithNumber: %w", err)
		}
		internalCellUUIDs = append(internalCellUUIDs, createdCell.ID)
	}

	if ipAddress != "" {
		if err := uc.orangePIWebAPI.SendCellUUIDs(ctx, ipAddress, automat.ID, cellUUIDs, internalCellUUIDs); err != nil {
			fmt.Printf("Warning: Failed to send cell UUIDs to OrangePI at %s: %v\n", ipAddress, err)
		}
	}

	return automat, nil
}

func (uc *ParcelAutomatUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.ParcelAutomat, error) {
	automat, err := uc.parcelAutomatRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - GetByID - parcelAutomatRepo.GetByID: %w", err)
	}
	return automat, nil
}

func (uc *ParcelAutomatUseCase) List(ctx context.Context) ([]*entity.ParcelAutomat, error) {
	automats, err := uc.parcelAutomatRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - List - parcelAutomatRepo.List: %w", err)
	}
	return automats, nil
}

func (uc *ParcelAutomatUseCase) ListWorking(ctx context.Context) ([]*entity.ParcelAutomat, error) {
	automats, err := uc.parcelAutomatRepo.ListWorking(ctx)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - ListWorking - parcelAutomatRepo.ListWorking: %w", err)
	}
	return automats, nil
}

func (uc *ParcelAutomatUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, isWorking bool) (*entity.ParcelAutomat, error) {
	automat, err := uc.parcelAutomatRepo.UpdateStatus(ctx, id, isWorking)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - UpdateStatus - parcelAutomatRepo.UpdateStatus: %w", err)
	}

	return automat, nil
}

func (uc *ParcelAutomatUseCase) Update(ctx context.Context, id uuid.UUID, city, address, ipAddress, coordinates string) (*entity.ParcelAutomat, error) {
	automat, err := uc.parcelAutomatRepo.Update(ctx, id, city, address, ipAddress, coordinates)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - Update - parcelAutomatRepo.Update: %w", err)
	}

	if ipAddress != "" {
		outCells, err := uc.lockerRepo.ListCellsByPostID(ctx, automat.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to get cells for automat %s: %v\n", automat.ID, err)
		} else {
			internalCells, err := uc.internalLockerRepo.ListCellsByPostID(ctx, automat.ID)
			if err != nil {
				fmt.Printf("Warning: Failed to get internal cells for automat %s: %v\n", automat.ID, err)
			} else {
				outCellUUIDs := make([]uuid.UUID, 0, len(outCells))
				for _, cell := range outCells {
					outCellUUIDs = append(outCellUUIDs, cell.ID)
				}

				internalCellUUIDs := make([]uuid.UUID, 0, len(internalCells))
				for _, cell := range internalCells {
					internalCellUUIDs = append(internalCellUUIDs, cell.ID)
				}

				if err := uc.orangePIWebAPI.SendCellUUIDs(ctx, ipAddress, automat.ID, outCellUUIDs, internalCellUUIDs); err != nil {
					fmt.Printf("Warning: Failed to send cell UUIDs to OrangePI at %s: %v\n", ipAddress, err)
				}
			}
		}
	}

	return automat, nil
}

func (uc *ParcelAutomatUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := uc.parcelAutomatRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("parcel automat usecase - Delete - parcelAutomatRepo.Delete: %w", err)
	}

	return nil
}

func (uc *ParcelAutomatUseCase) GetAutomatCells(ctx context.Context, automatID uuid.UUID) ([]*entity.LockerCell, error) {
	cells, err := uc.lockerRepo.ListCellsByPostID(ctx, automatID)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - GetAutomatCells - lockerRepo.ListCellsByPostID: %w", err)
	}
	return cells, nil
}

func (uc *ParcelAutomatUseCase) ProcessQRScan(ctx context.Context, qrDataJSON string, automatID uuid.UUID) ([]uuid.UUID, error) {
	user, err := uc.qrUseCase.ValidateQR(ctx, qrDataJSON)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - ProcessQRScan - qrUseCase.ValidateQR: %w", err)
	}

	orders, err := uc.orderRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - ProcessQRScan - orderRepo.ListByUserID: %w", err)
	}

	var cellIDs []uuid.UUID
	for _, order := range orders {
		if order.Status == "delivered" && order.LockerCellID != nil {
			delivery, err := uc.deliveryRepo.GetByOrderID(ctx, order.ID)
			if err != nil {
				continue
			}

			if delivery.ParcelAutomatID == automatID {
				cell, err := uc.lockerRepo.GetCellByID(ctx, *order.LockerCellID)
				if err != nil {
					continue
				}

				if cell.Status == "occupied" {
					cellIDs = append(cellIDs, cell.ID)
				}
			}
		}
	}

	for _, cellID := range cellIDs {
		if err := uc.lockerRepo.UpdateCellStatus(ctx, cellID, "opened"); err != nil {
			log.Printf("[PARCEL_AUTOMAT] Failed to open external cell %s: %v", cellID, err)
		}
	}

	return cellIDs, nil
}

func (uc *ParcelAutomatUseCase) UpdateCell(ctx context.Context, cellID uuid.UUID, height, length, width float64) (*entity.LockerCell, error) {
	cell, err := uc.lockerRepo.GetCellByID(ctx, cellID)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - UpdateCell - lockerRepo.GetCellByID: %w", err)
	}

	if cell.Status != "available" {
		return nil, fmt.Errorf("cannot update cell dimensions while cell is %s", cell.Status)
	}

	updatedCell, err := uc.lockerRepo.UpdateDimensions(ctx, cellID, height, length, width)
	if err != nil {
		return nil, fmt.Errorf("parcel automat usecase - UpdateCell - lockerRepo.UpdateDimensions: %w", err)
	}

	return updatedCell, nil
}

func (uc *ParcelAutomatUseCase) ConfirmPickup(ctx context.Context, cellIDs []uuid.UUID) error {
	for _, cellID := range cellIDs {
		cell, err := uc.lockerRepo.GetCellByID(ctx, cellID)
		if err != nil {
			continue
		}

		if cell.Status == "opened" {
			if err := uc.lockerRepo.UpdateCellStatus(ctx, cellID, "available"); err != nil {
				return fmt.Errorf("parcel automat usecase - ConfirmPickup - lockerRepo.UpdateCellStatus: %w", err)
			}

			// Find order by cell_id and release internal cell
			orders, err := uc.orderRepo.List(ctx)
			if err == nil {
				for _, order := range orders {
					if order.LockerCellID != nil && *order.LockerCellID == cellID {
						delivery, err := uc.deliveryRepo.GetByOrderID(ctx, order.ID)
						if err == nil && delivery.InternalLockerCellID != nil {
							if uc.internalLockerRepo != nil {
								if err := uc.internalLockerRepo.UpdateCellStatus(ctx, *delivery.InternalLockerCellID, "available"); err != nil {
									log.Printf("[PARCEL_AUTOMAT] Failed to release internal cell %s: %v", *delivery.InternalLockerCellID, err)
								} else {
									log.Printf("[PARCEL_AUTOMAT] Released internal cell %s for order %s", *delivery.InternalLockerCellID, order.ID)
								}
							}
						}
						// Update order status to completed
						uc.orderRepo.UpdateStatus(ctx, order.ID, "completed")
						break
					}
				}
			}
		}
	}

	return nil
}

func (uc *ParcelAutomatUseCase) PrepareCell(ctx context.Context, orderID, parcelAutomatID uuid.UUID) (uuid.UUID, *uuid.UUID, error) {
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("parcel automat usecase - PrepareCell - orderRepo.GetByID: %w", err)
	}

	if order.LockerCellID == nil {
		return uuid.Nil, nil, fmt.Errorf("parcel automat usecase - PrepareCell: order has no cell assigned")
	}

	cell, err := uc.lockerRepo.GetCellByID(ctx, *order.LockerCellID)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("parcel automat usecase - PrepareCell - lockerRepo.GetCellByID: %w", err)
	}

	automat, err := uc.parcelAutomatRepo.GetByID(ctx, parcelAutomatID)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("parcel automat usecase - PrepareCell - parcelAutomatRepo.GetByID: %w", err)
	}

	var internalDoorID *uuid.UUID
	delivery, err := uc.deliveryRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
	if err != nil {
		fmt.Printf("parcel automat usecase - PrepareCell - deliveryRepo.GetByOrderID warning: %v\n", err)
	} else if delivery.InternalLockerCellID != nil {
		internalDoorID = delivery.InternalLockerCellID
		if err := uc.orangePIWebAPI.OpenCell(ctx, automat.IPAddress, *internalDoorID); err != nil {
			return uuid.Nil, nil, fmt.Errorf("parcel automat usecase - PrepareCell - orangePIWebAPI.OpenCell (internal): %w", err)
		}
		if uc.internalLockerRepo != nil {
			if err := uc.internalLockerRepo.UpdateCellStatus(ctx, *internalDoorID, "opened"); err != nil {
				fmt.Printf("parcel automat usecase - PrepareCell - internalLockerRepo.UpdateCellStatus: %v\n", err)
			}
		}
	}

	return cell.ID, internalDoorID, nil
}
