package grpc

import "errors"

var (
	ErrGRPCConnectionFailed = errors.New("failed to connect to gRPC server")
	ErrGRPCRequestFailed    = errors.New("gRPC request failed")
	ErrGRPCClientNotReady   = errors.New("gRPC client is not ready")
)
