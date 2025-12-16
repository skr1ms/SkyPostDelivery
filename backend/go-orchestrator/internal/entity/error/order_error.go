package error

import "errors"

var (
	ErrOrderNotFound             = errors.New("order not found")
	ErrOrderNotBelongsToUser     = errors.New("order does not belong to user")
	ErrOrderCannotBeReturned     = errors.New("order cannot be returned")
	ErrOrderCreateFailed         = errors.New("failed to create order")
	ErrOrderUpdateFailed         = errors.New("failed to update order")
	ErrOrderNoWorkingAutomats    = errors.New("no working parcel automats available")
	ErrOrderNoAvailableCell      = errors.New("no available cell for good dimensions")
	ErrOrderHasNoCellAssigned    = errors.New("order has no cell assigned")
	ErrOrderCreateMultipleFailed = errors.New("failed to create any orders")
)
