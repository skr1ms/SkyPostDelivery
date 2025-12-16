package webapi

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type fcmSender struct {
	client *messaging.Client
}

type noopSender struct{}

type PushSender interface {
	SendDeliveryNotification(ctx context.Context, tokens []string, orderID string, lockerCellID *string) ([]string, error)
}

func NewFCMSender(ctx context.Context, credentialsFile, projectID string) (*fcmSender, error) {
	opts := []option.ClientOption{}
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}

	var config *firebase.Config
	if projectID != "" {
		config = &firebase.Config{
			ProjectID: projectID,
		}
	}

	app, err := firebase.NewApp(ctx, config, opts...)
	if err != nil {
		return nil, fmt.Errorf("fcmSender - NewFCMSender - NewApp: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("fcmSender - NewFCMSender - Messaging: %w", err)
	}
	return &fcmSender{client: client}, nil
}

func NewNoopSender() *noopSender {
	return &noopSender{}
}

func (s *fcmSender) SendDeliveryNotification(ctx context.Context, tokens []string, orderID string, lockerCellID *string) ([]string, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	data := map[string]string{
		"order_id": orderID,
	}
	if lockerCellID != nil && *lockerCellID != "" {
		data["locker_cell_id"] = *lockerCellID
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: "Посылка доставлена",
			Body:  "Заказ готов к выдаче",
		},
		Data: data,
	}

	sendResponses, err := s.client.SendEachForMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("fcmSender - SendDeliveryNotification - SendEachForMulticast: %w", err)
	}

	invalid := make([]string, 0)
	for index, resp := range sendResponses.Responses {
		if resp.Success || resp.Error == nil {
			continue
		}

		if messaging.IsUnregistered(resp.Error) ||
			messaging.IsInvalidArgument(resp.Error) {
			invalid = append(invalid, tokens[index])
		}
	}
	return invalid, nil
}

func (s *noopSender) SendDeliveryNotification(ctx context.Context, tokens []string, orderID string, lockerCellID *string) ([]string, error) {
	return nil, nil
}
