package usecase

import "context"

type (
	DroneNotifier interface {
		SendToDrone(ctx context.Context, droneID string, message map[string]any) error
	}
)
