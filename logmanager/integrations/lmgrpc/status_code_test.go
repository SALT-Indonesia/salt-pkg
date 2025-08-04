package lmgrpc_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgrpc"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestConvertCodeToHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    codes.Code
		expected int
	}{
		{name: "OK", input: codes.OK, expected: http.StatusOK},
		{name: "Canceled", input: codes.Canceled, expected: 499},
		{name: "Unknown", input: codes.Unknown, expected: http.StatusInternalServerError},
		{name: "InvalidArgument", input: codes.InvalidArgument, expected: http.StatusBadRequest},
		{name: "DeadlineExceeded", input: codes.DeadlineExceeded, expected: http.StatusGatewayTimeout},
		{name: "NotFound", input: codes.NotFound, expected: http.StatusNotFound},
		{name: "AlreadyExists", input: codes.AlreadyExists, expected: http.StatusConflict},
		{name: "PermissionDenied", input: codes.PermissionDenied, expected: http.StatusForbidden},
		{name: "ResourceExhausted", input: codes.ResourceExhausted, expected: http.StatusTooManyRequests},
		{name: "FailedPrecondition", input: codes.FailedPrecondition, expected: http.StatusBadRequest},
		{name: "Aborted", input: codes.Aborted, expected: http.StatusConflict},
		{name: "OutOfRange", input: codes.OutOfRange, expected: http.StatusBadRequest},
		{name: "Unimplemented", input: codes.Unimplemented, expected: http.StatusNotImplemented},
		{name: "Internal", input: codes.Internal, expected: http.StatusInternalServerError},
		{name: "Unavailable", input: codes.Unavailable, expected: http.StatusServiceUnavailable},
		{name: "DataLoss", input: codes.DataLoss, expected: http.StatusInternalServerError},
		{name: "Unauthenticated", input: codes.Unauthenticated, expected: http.StatusUnauthorized},
		{name: "DefaultCase", input: codes.Code(999), expected: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := lmgrpc.ConvertCodeToHTTPStatus(tt.input)
			assert.Equal(t, tt.expected, output)
		})
	}
}
