package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockBalanceUpdater struct {
	balance map[string]float64
	err     error
}

func (m *MockBalanceUpdater) UpdateUsersBalance(username, currency string, amount float64) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *MockBalanceUpdater) GetUserBalance(username string) (map[string]float64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.balance, nil
}

func TestHandleDeposit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validToken, err := GenerateJWT("testuser")
	if err != nil {
		t.Fatalf("Failed to generate valid JWT: %v", err)
	}

	testCases := []struct {
		name             string
		token            string
		requestBody      string
		mockBalance      map[string]float64
		mockError        error
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "No Token",
			token:            "",
			requestBody:      `{"currency":"USD","amount":100}`,
			mockBalance:      nil,
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"not authorized", "status":"error"}`,
		},
		{
			name:             "Invalid Token",
			token:            "invalid_token",
			requestBody:      `{"currency":"USD","amount":100}`,
			mockBalance:      nil,
			mockError:        errors.New("invalid token"),
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"token contains an invalid number of segments", "status":"error"}`,
		},
		{
			name:             "Invalid JSON - Wrong Type",
			token:            validToken,
			requestBody:      `{"currency":"USD","amount":"string"}`,
			mockBalance:      nil,
			mockError:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"request contains wrong data", "status":"error"}`,
		},
		{
			name:             "Valid Request - Success",
			token:            validToken,
			requestBody:      `{"currency":"USD","amount":100}`,
			mockBalance:      map[string]float64{"USD": 200, "EUR": 50},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"message":"Account topped up successfully","balance":{"EUR":50,"USD":200}}`,
		},
		{
			name:             "Valid Request - Update Balance Error",
			token:            validToken,
			requestBody:      `{"currency":"USD","amount":100}`,
			mockBalance:      nil,
			mockError:        errors.New("database error"),
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"database error", "status":"error"}`,
		},
		{
			name:             "Valid Request - Get Balance Error",
			token:            validToken,
			requestBody:      `{"currency":"USD","amount":100}`,
			mockBalance:      nil,
			mockError:        errors.New("get balance error"),
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"get balance error", "status":"error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body := bytes.NewBufferString(tc.requestBody)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/deposit", body)
			c.Request.Header.Set("Content-Type", "application/json")

			c.Request.Header.Set("Authorization", tc.token)

			logger := newTestLogger()

			mockBalanceUpdater := &MockBalanceUpdater{
				balance: tc.mockBalance,
				err:     tc.mockError,
			}

			HandleDeposit(logger, mockBalanceUpdater)(c)

			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")
			assert.JSONEq(t, tc.expectedResponse, w.Body.String(), "Response body mismatch")
		})
	}
}
