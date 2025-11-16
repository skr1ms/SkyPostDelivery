package grpc

type CellOpenResponse struct {
	Success        bool
	Message        string
	CellID         string
	InternalCellID string
}
