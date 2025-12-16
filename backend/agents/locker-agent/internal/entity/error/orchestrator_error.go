package error

import "errors"

var (
	ErrOrchestratorConnectionFailed = errors.New("orchestrator connection failed")
	ErrOrchestratorRequestFailed    = errors.New("orchestrator request failed")
	ErrOrchestratorInvalidResponse  = errors.New("invalid orchestrator response")
	ErrOrchestratorTimeout          = errors.New("orchestrator request timeout")
)
