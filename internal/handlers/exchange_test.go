package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Foreground-Eclipse/grpcexchanger/proto"
)

type MockExchanger struct {
	balance map[string]float64
	err     error
}

func (m *MockExchanger) GetUserBalance(username string) (map[string]float64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.balance, nil
}

func (m *MockExchanger) UpdateUsersBalance(username, currency string, amount float64) error {
	if m.err != nil {
		return m.err
	}
	if m.balance != nil {
		if _, ok := m.balance[currency]; ok {
			m.balance[currency] += amount
		} else {
			m.balance[currency] = amount
		}
	}
	return nil
}

func (m *MockExchangeServiceClient) GetExchangeRates(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ExchangeRatesResponse, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &pb.ExchangeRatesResponse{
		Rates: m.rates,
	}, nil
}

type MockExchangeServiceClient struct {
	rates map[string]float32
	err   error
}

func (m *MockExchangeServiceClient) GetExchangeRateForCurrency(ctx context.Context, in *pb.CurrencyRequest, opts ...grpc.CallOption) (*pb.ExchangeRateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetExchangeRateForCurrency not implemented")
}

func (s *MockExchangeServiceClient) mustEmbedUnimplementedExchangeServiceServer() {}

func TestHandleExchange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validToken, err := GenerateJWT("testuser")
	if err != nil {
		t.Fatalf("Failed to generate valid JWT: %v", err)
	}

	testCases := []struct {
		name               string
		token              string
		requestBody        string
		mockBalance        map[string]float64
		mockExchangerError error
		mockGrpcRates      map[string]float32
		mockGrpcError      error
		expectedStatus     int
		expectedResponse   string
	}{
		{
			name:               "No Token",
			token:              "",
			requestBody:        `{"from_currency":"USD","to_currency":"EUR","amount":100}`,
			mockBalance:        nil,
			mockExchangerError: nil,
			mockGrpcRates: map[string]float32{
				"RUB_USD": 0.012,
				"RUB_EUR": 0.011,
			},
			mockGrpcError:    nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"not authorized", "status":"error"}`,
		},
		{
			name:               "Invalid Token",
			token:              "invalid_token",
			requestBody:        `{"from_currency":"USD","to_currency":"EUR","amount":100}`,
			mockBalance:        nil,
			mockExchangerError: nil,
			mockGrpcRates: map[string]float32{
				"RUB_USD": 0.012,
				"RUB_EUR": 0.011,
			},
			mockGrpcError:    nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"token contains an invalid number of segments", "status":"error"}`,
		},
		{
			name:               "Invalid JSON - Wrong Type",
			token:              validToken,
			requestBody:        `{"from_currency":"USD","to_currency":"EUR","amount":"string"}`,
			mockBalance:        nil,
			mockExchangerError: nil,
			mockGrpcRates: map[string]float32{
				"RUB_USD": 0.012,
				"RUB_EUR": 0.011,
			},
			mockGrpcError:    nil,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"request contains wrong data", "status":"error"}`,
		},
		{
			name:               "Not Enough Money",
			token:              validToken,
			requestBody:        `{"from_currency":"USD","to_currency":"EUR","amount":100}`,
			mockBalance:        map[string]float64{"USD": 50, "EUR": 50},
			mockExchangerError: nil,
			mockGrpcRates: map[string]float32{
				"RUB_USD": 0.012,
				"RUB_EUR": 0.011,
			},
			mockGrpcError:    nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"not enough money", "status":"error"}`,
		},
		{
			name:               "gRPC Error",
			token:              validToken,
			requestBody:        `{"from_currency":"USD","to_currency":"EUR","amount":100}`,
			mockBalance:        map[string]float64{"USD": 100, "EUR": 50},
			mockExchangerError: nil,
			mockGrpcRates:      nil,
			mockGrpcError:      status.Error(codes.Unavailable, "gRPC service unavailable"),
			expectedStatus:     http.StatusInternalServerError,
			expectedResponse:   `{"error":"failed to retrieve exchange rates", "status":"error"}`,
		},
		{
			name:               "GetUserBalance Error",
			token:              validToken,
			requestBody:        `{"from_currency":"USD","to_currency":"EUR","amount":100}`,
			mockBalance:        nil,
			mockExchangerError: errors.New("database error"),
			mockGrpcRates: map[string]float32{
				"RUB_USD": 0.012,
				"RUB_EUR": 0.011,
			},
			mockGrpcError:    nil,
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"error":"database error", "status":"error"}`,
		},
		{
			name:               "Valid Request - Success",
			token:              validToken,
			requestBody:        `{"from_currency":"USD","to_currency":"EUR","amount":100}`,
			mockBalance:        map[string]float64{"USD": 100, "EUR": 50},
			mockExchangerError: nil,
			mockGrpcRates: map[string]float32{
				"RUB_USD": 0.012,
				"RUB_EUR": 0.011,
			},
			mockGrpcError:    nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"message":"exchanged successfully","exchanged_amount":91.66666666666667,"new_balance":{"EUR":141.66666666666669,"USD":0}}`,
		},
		{
			name:               "Update Users Balance error",
			token:              validToken,
			requestBody:        `{"from_currency":"USD","to_currency":"EUR","amount":100}`,
			mockBalance:        map[string]float64{"USD": 100, "EUR": 50},
			mockExchangerError: errors.New("update balance error"),
			mockGrpcRates: map[string]float32{
				"RUB_USD": 0.012,
				"RUB_EUR": 0.011,
			},
			mockGrpcError:    nil,
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"error":"update balance error", "status":"error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body := bytes.NewBufferString(tc.requestBody)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/exchange", body)
			c.Request.Header.Set("Content-Type", "application/json")

			c.Request.Header.Set("Authorization", tc.token)

			logger := newTestLogger()

			mockExchanger := &MockExchanger{
				balance: tc.mockBalance,
				err:     tc.mockExchangerError,
			}

			mockGrpcClient := &MockExchangeServiceClient{
				rates: tc.mockGrpcRates,
				err:   tc.mockGrpcError,
			}

			HandleExchange(logger, mockGrpcClient, mockExchanger)(c)

			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")
			assert.JSONEq(t, tc.expectedResponse, w.Body.String(), "Response body mismatch")
		})
	}
}
