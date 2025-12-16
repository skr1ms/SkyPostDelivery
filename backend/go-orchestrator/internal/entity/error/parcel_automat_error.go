package error

import "errors"

var (
	ErrParcelAutomatNotFound             = errors.New("parcel automat not found")
	ErrParcelAutomatCreateFailed         = errors.New("failed to create parcel automat")
	ErrParcelAutomatUpdateFailed         = errors.New("failed to update parcel automat")
	ErrParcelAutomatDeleteFailed         = errors.New("failed to delete parcel automat")
	ErrQRNoOrdersForPickup               = errors.New("no orders available for pickup")
	ErrParcelAutomatPartialPickupFailure = errors.New("some cells failed to process during pickup")
)
