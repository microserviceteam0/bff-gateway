package clients

import (
	"fmt"
	"net/http"

	"github.com/microserviceteam0/bff-gateway/bff/internal/apperr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapStatusToError converts an HTTP status code and response body into a domain error.
func MapStatusToError(statusCode int, body string) error {
	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", apperr.ErrInvalidInput, body)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", apperr.ErrUnauthorized, body)
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", apperr.ErrForbidden, body)
	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", apperr.ErrNotFound, body)
	case http.StatusConflict:
		return fmt.Errorf("%w: %s", apperr.ErrAlreadyExists, body)
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		return fmt.Errorf("%w: %s", apperr.ErrServiceUnavailable, body)
	case http.StatusRequestTimeout:
		return fmt.Errorf("%w: %s", apperr.ErrTimeout, body)
	default:
		return fmt.Errorf("downstream service error: status %d, body: %s", statusCode, body)
	}
}

// MapGRPCError converts a gRPC error into an application-level error.
func MapGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.OK:
		return nil
	case codes.NotFound:
		return fmt.Errorf("%w: %s", apperr.ErrNotFound, st.Message())
	case codes.InvalidArgument, codes.FailedPrecondition:
		return fmt.Errorf("%w: %s", apperr.ErrInvalidInput, st.Message())
	case codes.Unauthenticated:
		return fmt.Errorf("%w: %s", apperr.ErrUnauthorized, st.Message())
	case codes.PermissionDenied:
		return fmt.Errorf("%w: %s", apperr.ErrForbidden, st.Message())
	case codes.AlreadyExists:
		return fmt.Errorf("%w: %s", apperr.ErrAlreadyExists, st.Message())
	case codes.Unavailable:
		return fmt.Errorf("%w: %s", apperr.ErrServiceUnavailable, st.Message())
	case codes.DeadlineExceeded:
		return fmt.Errorf("%w: %s", apperr.ErrTimeout, st.Message())
	case codes.Internal:
		return fmt.Errorf("%w: %s", apperr.ErrInternal, st.Message())
	default:
		return fmt.Errorf("%w: %s", apperr.ErrInternal, st.Message())
	}
}
