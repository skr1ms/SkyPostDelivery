package error

import "errors"

var (
	ErrDroneNotFound        = errors.New("drone not found")
	ErrDroneNotAvailable    = errors.New("no available drones")
	ErrDroneInvalidModel    = errors.New("drone model cannot be empty")
	ErrDroneInvalidIP       = errors.New("invalid drone IP address")
	ErrDroneInvalidStatus   = errors.New("invalid drone status")
	ErrDroneNothingToUpdate = errors.New("nothing to update")
	ErrDroneCannotDelete    = errors.New("cannot delete drone")
	ErrDroneCreateFailed    = errors.New("failed to create drone")
)
