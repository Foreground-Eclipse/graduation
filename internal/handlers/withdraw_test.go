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

type MockBalanceWithdrawer struct {
	balance map[string]float64
	err     error
}

func (m *MockBalanceWithdrawer) UpdateUsersBalance(username, currency string, amount float64) error {
	if m.err != nil {
		return m.err
	}
	if m.balance != nil {
		if _, ok := m.balance[currency]; ok {
			m.balance[currency] += amount
		}
	}
	return nil
}

func (m *MockBalanceWithdrawer) GetUserBalance(username string) (map[string]float64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.balance, nil
}

func TestHandleWithdraw(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Generate valid JWT
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
			requestBody:      `{"currency":"USD","amount":50}`,
			mockBalance:      map[string]float64{"USD": 100},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"not authorized", "status":"error"}`,
		},
		{
			name:             "Invalid Token",
			token:            "invalid_token",
			requestBody:      `{"currency":"USD","amount":50}`,
			mockBalance:      map[string]float64{"USD": 100},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"token contains an invalid number of segments", "status":"error"}`,
		},
		{
			name:             "Invalid JSON - Wrong Type",
			token:            validToken,
			requestBody:      `{"currency":"USD","amount":"string"}`,
			mockBalance:      map[string]float64{"USD": 100},
			mockError:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"request contains wrong data", "status":"error"}`,
		},
		{
			name:             "Insufficient Funds",
			token:            validToken,
			requestBody:      `{"currency":"USD","amount":150}`,
			mockBalance:      map[string]float64{"USD": 100},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"not enough money to withdraw", "status":"error"}`,
		},
		{
			name:             "Withdrawal Error",
			token:            validToken,
			requestBody:      `{"currency":"USD","amount":50}`,
			mockBalance:      map[string]float64{"USD": 100},
			mockError:        errors.New("database error"),
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"database error", "status":"error"}`,
		},
		{
			name:             "Successful Withdrawal",
			token:            validToken,
			requestBody:      `{"currency":"USD","amount":50}`,
			mockBalance:      map[string]float64{"USD": 100.0, "EUR": 50.0},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"message":"Withdrawal successfull","balance":{"EUR":50,"USD":50}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body := bytes.NewBufferString(tc.requestBody)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/withdraw", body)
			c.Request.Header.Set("Content-Type", "application/json")

			c.Request.Header.Set("Authorization", tc.token)

			logger := newTestLogger()

			mockBalanceWithdrawer := &MockBalanceWithdrawer{
				balance: tc.mockBalance,
				err:     tc.mockError,
			}

			HandleWithdraw(logger, mockBalanceWithdrawer)(c)

			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")
			assert.JSONEq(t, tc.expectedResponse, w.Body.String(), "Response body mismatch")
		})
	}
}
