package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type MockBalanceGetter struct {
	balance map[string]float64
	err     error
}

func (m *MockBalanceGetter) GetUserBalance(username string) (map[string]float64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.balance, nil
}

func newTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

var jwtKey = []byte("supersecretkey")

type JWTClaim struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.StandardClaims
}

func GenerateJWT(username string) (tokenString string, err error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &JWTClaim{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)
	return
}

func ValidateToken(signedToken string) (string, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
	)
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return "", err
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return "", err
	}
	return claims.Username, nil
}

func TestHandleBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validToken, err := GenerateJWT("testuser")
	if err != nil {
		t.Fatalf("Failed to generate valid JWT: %v", err)
	}

	testCases := []struct {
		name             string
		token            string
		mockBalance      map[string]float64
		mockError        error
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "No Token",
			token:            "",
			mockBalance:      nil,
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"not authorized", "status":"error"}`,
		},
		{
			name:             "Invalid Token",
			token:            "invalid_token",
			mockBalance:      nil,
			mockError:        errors.New("invalid token"),
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"token contains an invalid number of segments", "status":"error"}`,
		},
		{
			name:             "Valid Token - Success",
			token:            validToken,
			mockBalance:      map[string]float64{"USD": 100, "EUR": 50},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"balance":{"EUR":50,"USD":100}}`,
		},
		{
			name:             "Valid Token - GetBalance Error",
			token:            validToken,
			mockBalance:      nil,
			mockError:        errors.New("database error"),
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"database error",  "status":"error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/balance", nil)

			c.Request.Header.Set("Authorization", tc.token)

			logger := newTestLogger()

			mockBalanceGetter := &MockBalanceGetter{
				balance: tc.mockBalance,
				err:     tc.mockError,
			}

			HandleBalance(logger, mockBalanceGetter)(c)

			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")
			assert.JSONEq(t, tc.expectedResponse, w.Body.String(), "Response body mismatch") // Use JSONEq for comparing JSON
		})
	}
}
