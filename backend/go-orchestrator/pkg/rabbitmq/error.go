package rabbitmq

import "errors"

var (
	ErrClientNotReady          = errors.New("rabbitmq client is not ready")
	ErrConnectionFailed        = errors.New("failed to connect to RabbitMQ")
	ErrChannelOpenFailed       = errors.New("failed to open channel")
	ErrChannelConfirmFailed    = errors.New("failed to put channel into confirm mode")
	ErrQueueDeclareFailed      = errors.New("failed to declare queue")
	ErrMessageMarshalFailed    = errors.New("failed to marshal message")
	ErrMessagePublishFailed    = errors.New("failed to publish message")
	ErrMessageNacked           = errors.New("message was nacked by RabbitMQ")
	ErrPublishTimeout          = errors.New("timeout waiting for publish confirmation")
	ErrPublishContextCancelled = errors.New("context cancelled while waiting for confirm")
	ErrConsumerRegisterFailed  = errors.New("failed to register consumer")
)
