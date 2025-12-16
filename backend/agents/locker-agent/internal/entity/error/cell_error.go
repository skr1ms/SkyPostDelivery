package error

import "errors"

var (
	ErrCellNotInitialized     = errors.New("cell mapping not initialized")
	ErrCellNotFound           = errors.New("cell not found in mapping")
	ErrCellInvalidNumber      = errors.New("invalid cell number")
	ErrCellInvalidUUID        = errors.New("invalid cell UUID")
	ErrCellSyncFailed         = errors.New("failed to sync cells")
	ErrCellOpenFailed         = errors.New("failed to open cell")
	ErrInternalDoorOpenFailed = errors.New("failed to open internal door")
	ErrCellPrepareFailed      = errors.New("failed to prepare cell")
)
