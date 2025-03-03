package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockExchangeServiceClient - Mock gRPC client

func TestHandleRates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validToken, err := GenerateJWT("testuser")
	if err != nil {
		t.Fatalf("Failed to generate valid JWT: %v", err)
	}

	testCases := []struct {
		name             string
		token            string
		mockGrpcRates    map[string]float32
		mockGrpcError    error
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "No Token",
			token:            "",
			mockGrpcRates:    nil,
			mockGrpcError:    nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"not authorized", "status":"error"}`,
		},
		{
			name:             "Invalid Token",
			token:            "invalid_token",
			mockGrpcRates:    nil,
			mockGrpcError:    nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"token contains an invalid number of segments", "status":"error"}`,
		},
		{
			name:             "gRPC Error",
			token:            validToken,
			mockGrpcRates:    nil,
			mockGrpcError:    status.Error(codes.Unavailable, "gRPC service unavailable"),
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"error":"Failed to retrieve exchange rates", "status":"error"}`,
		},
		{
			name:             "Valid Request - Success",
			token:            validToken,
			mockGrpcRates:    map[string]float32{"RUB_USD": 0.012, "RUB_EUR": 0.011},
			mockGrpcError:    nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"rates":{"RUB_EUR":0.011,"RUB_USD":0.012}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/rates", nil)
			c.Request.Header.Set("Authorization", tc.token)

			logger := newTestLogger()

			mockGrpcClient := &MockExchangeServiceClient{
				rates: tc.mockGrpcRates,
				err:   tc.mockGrpcError,
			}

			HandleRates(logger, mockGrpcClient)(c)

			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")
			assert.JSONEq(t, tc.expectedResponse, w.Body.String(), "Response body mismatch")
		})
	}
}
