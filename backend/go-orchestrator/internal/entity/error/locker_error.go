package error

import "errors"

var (
	ErrLockerCellNotFound      = errors.New("locker cell not found")
	ErrLockerCellCreateFailed  = errors.New("failed to create locker cell")
	ErrLockerCellAlreadyExists = errors.New("locker cell already exists")
	ErrLockerCellUpdateFailed  = errors.New("failed to update locker cell")
	ErrLockerCellCannotUpdate  = errors.New("cannot update cell dimensions while cell is not available")
	ErrLockerCellInvalidStatus = errors.New("locker cell has invalid status for this operation")
	ErrLockerInvalidStatus     = errors.New("invalid cell status")
)
