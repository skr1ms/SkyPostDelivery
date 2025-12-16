package error

import "errors"

var (
	ErrGoodNotFound          = errors.New("good not found")
	ErrGoodOutOfStock        = errors.New("good is out of stock")
	ErrGoodInvalidName       = errors.New("good name cannot be empty")
	ErrGoodInvalidDimensions = errors.New("invalid good dimensions")
	ErrGoodInvalidQuantity   = errors.New("invalid quantity")
	ErrGoodCreateFailed      = errors.New("failed to create good")
	ErrGoodUpdateFailed      = errors.New("failed to update good")
	ErrGoodDeleteFailed      = errors.New("failed to delete good")
)
