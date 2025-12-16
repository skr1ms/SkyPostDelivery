package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/response"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	jwtError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/jwt"
)

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entityError.ErrDroneInvalidModel),
		errors.Is(err, entityError.ErrDroneInvalidIP),
		errors.Is(err, entityError.ErrDroneInvalidStatus),
		errors.Is(err, entityError.ErrDroneNothingToUpdate),
		errors.Is(err, entityError.ErrGoodInvalidName),
		errors.Is(err, entityError.ErrGoodInvalidDimensions),
		errors.Is(err, entityError.ErrGoodInvalidQuantity),
		errors.Is(err, entityError.ErrOrderCannotBeReturned),
		errors.Is(err, entityError.ErrOrderHasNoCellAssigned),
		errors.Is(err, entityError.ErrLockerInvalidStatus),
		errors.Is(err, entityError.ErrLockerCellInvalidStatus),
		errors.Is(err, entityError.ErrLockerCellCannotUpdate),
		errors.Is(err, entityError.ErrDeliveryInvalidStatus),
		errors.Is(err, entityError.ErrInvalidVerificationCode),
		errors.Is(err, entityError.ErrVerificationCodeExpired),
		errors.Is(err, entityError.ErrPasswordNotSet),
		errors.Is(err, entityError.ErrUserNotFoundByEmail),
		errors.Is(err, entityError.ErrUserEmailMismatch):
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})

	case errors.Is(err, entityError.ErrDroneNotFound),
		errors.Is(err, entityError.ErrDroneNotAvailable),
		errors.Is(err, entityError.ErrGoodNotFound),
		errors.Is(err, entityError.ErrOrderNotFound),
		errors.Is(err, entityError.ErrUserNotFound),
		errors.Is(err, entityError.ErrUserNotFoundByPhone),
		errors.Is(err, entityError.ErrLockerCellNotFound),
		errors.Is(err, entityError.ErrParcelAutomatNotFound),
		errors.Is(err, entityError.ErrDeliveryNotFound),
		errors.Is(err, entityError.ErrDeviceNotFound),
		errors.Is(err, entityError.ErrQRNoOrdersForPickup):
		c.JSON(http.StatusNotFound, response.Error{Error: err.Error()})

	case errors.Is(err, entityError.ErrDroneCannotDelete),
		errors.Is(err, entityError.ErrGoodOutOfStock),
		errors.Is(err, entityError.ErrOrderNoAvailableCell),
		errors.Is(err, entityError.ErrOrderNoWorkingAutomats),
		errors.Is(err, entityError.ErrUserAlreadyExists),
		errors.Is(err, entityError.ErrUserEmailAlreadyExists),
		errors.Is(err, entityError.ErrUserPhoneAlreadyExists),
		errors.Is(err, entityError.ErrLockerCellAlreadyExists):
		c.JSON(http.StatusConflict, response.Error{Error: err.Error()})

	case errors.Is(err, entityError.ErrInvalidCredentials),
		errors.Is(err, entityError.ErrPhoneNotVerified),
		errors.Is(err, entityError.ErrQRValidationFailed),
		errors.Is(err, entityError.ErrQRUserMismatch),
		errors.Is(err, jwtError.ErrTokenInvalid),
		errors.Is(err, jwtError.ErrTokenExpired),
		errors.Is(err, jwtError.ErrTokenInvalidType),
		errors.Is(err, jwtError.ErrRefreshTokenInvalid),
		errors.Is(err, jwtError.ErrAccessTokenInvalid),
		errors.Is(err, jwtError.ErrTokenValidationFailed),
		errors.Is(err, jwtError.ErrTokenUnexpectedSigning):
		c.JSON(http.StatusUnauthorized, response.Error{Error: err.Error()})

	case errors.Is(err, jwtError.ErrTokenGenerationFailed):
		c.JSON(http.StatusInternalServerError, response.Error{Error: "Failed to generate tokens"})

	case errors.Is(err, entityError.ErrOrderNotBelongsToUser):
		c.JSON(http.StatusForbidden, response.Error{Error: err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, response.Error{Error: "Internal server error"})
	}
}
