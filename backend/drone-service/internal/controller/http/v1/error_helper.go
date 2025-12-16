package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/controller/http/v1/response"
	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	grpcError "github.com/skr1ms/SkyPostDelivery/drone-service/pkg/grpc"
	rabbitmqError "github.com/skr1ms/SkyPostDelivery/drone-service/pkg/rabbitmq"
)

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entityError.ErrDroneInvalidModel),
		errors.Is(err, entityError.ErrDroneInvalidIP),
		errors.Is(err, entityError.ErrDroneInvalidStatus),
		errors.Is(err, entityError.ErrDroneNothingToUpdate),
		errors.Is(err, entityError.ErrDeliveryInvalidStatus):
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})

	case errors.Is(err, entityError.ErrDroneNotFound),
		errors.Is(err, entityError.ErrDroneNotAvailable),
		errors.Is(err, entityError.ErrDroneStateNotFound),
		errors.Is(err, entityError.ErrDeliveryNotFound),
		errors.Is(err, entityError.ErrDeliveryTaskNotFound):
		c.JSON(http.StatusNotFound, response.Error{Error: err.Error()})

	case errors.Is(err, entityError.ErrDroneCannotDelete):
		c.JSON(http.StatusConflict, response.Error{Error: err.Error()})

	case errors.Is(err, grpcError.ErrGRPCClientNotReady),
		errors.Is(err, grpcError.ErrGRPCConnectionFailed),
		errors.Is(err, grpcError.ErrGRPCRequestFailed),
		errors.Is(err, rabbitmqError.ErrClientNotReady),
		errors.Is(err, rabbitmqError.ErrConnectionFailed),
		errors.Is(err, rabbitmqError.ErrChannelOpenFailed),
		errors.Is(err, rabbitmqError.ErrChannelConfirmFailed),
		errors.Is(err, rabbitmqError.ErrQueueDeclareFailed),
		errors.Is(err, rabbitmqError.ErrMessageMarshalFailed),
		errors.Is(err, rabbitmqError.ErrMessagePublishFailed),
		errors.Is(err, rabbitmqError.ErrMessageNacked),
		errors.Is(err, rabbitmqError.ErrPublishTimeout),
		errors.Is(err, rabbitmqError.ErrPublishContextCancelled),
		errors.Is(err, rabbitmqError.ErrConsumerRegisterFailed):
		c.JSON(http.StatusServiceUnavailable, response.Error{Error: err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, response.Error{Error: "Internal server error"})
	}
}
