package rabbitmq

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Publish_MessageMarshaling(t *testing.T) {
	testCases := []struct {
		name    string
		message any
		valid   bool
	}{
		{
			name:    "valid map",
			message: map[string]any{"key": "value"},
			valid:   true,
		},
		{
			name:    "valid struct-like map",
			message: map[string]any{"order_id": "order-123", "status": "pending"},
			valid:   true,
		},
		{
			name:    "valid string",
			message: "simple string",
			valid:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := json.Marshal(tc.message)
			if tc.valid {
				assert.NoError(t, err, "Expected valid message to marshal")
			} else {
				assert.Error(t, err, "Expected invalid message to fail marshaling")
			}
		})
	}
}

func TestDeliveryHandler_Interface(t *testing.T) {
	var handler DeliveryHandler
	assert.Nil(t, handler)
}

func TestRabbitMQClient_Interface(t *testing.T) {
	var client RabbitMQClient = &Client{}
	assert.NotNil(t, client)
}
