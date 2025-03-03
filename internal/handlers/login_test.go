package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Foreground-Eclipse/transferer/config"
	"github.com/Foreground-Eclipse/transferer/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockUserLogger - Mock implementation of UserLogger
type MockUserLogger struct {
	passhash string
	err      error
}

func (m *MockUserLogger) GetUsersPassHash(username string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.passhash, nil
}

func TestHandleLoginUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testPassword := "password123"
	hashedPassword := middleware.HashPassword(testPassword)

	testCases := []struct {
		name             string
		requestBody      string
		mockPasshash     string
		mockError        error
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "Invalid JSON - Wrong Type",
			requestBody:      `{"username":"testuser","password":123}`,
			mockPasshash:     "",
			mockError:        nil,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"request contains wrong data", "status":"error"}`,
		},
		{
			name:             "Invalid Credentials - Wrong Username",
			requestBody:      `{"username":"nonexistentuser","password":"password123"}`,
			mockPasshash:     "",
			mockError:        errors.New("user not found"),
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"wrong password", "status":"error"}`,
		},
		{
			name:             "Invalid Credentials - Wrong Password",
			requestBody:      `{"username":"testuser","password":"wrongpassword"}`,
			mockPasshash:     hashedPassword,
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"error":"wrong password", "status":"error"}`,
		},
		{
			name:             "Valid Credentials - Success",
			requestBody:      `{"username":"testuser","password":"password123"}`,
			mockPasshash:     hashedPassword,
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"token":""}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			body := bytes.NewBufferString(tc.requestBody)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/login", body)
			c.Request.Header.Set("Content-Type", "application/json")

			logger := newTestLogger()

			mockUserLogger := &MockUserLogger{
				passhash: tc.mockPasshash,
				err:      tc.mockError,
			}

			cfg := &config.Config{
				JWT: config.JWTConfig{
					JWTExpirationTimeHours: 1,
				},
			}

			HandleLoginUser(logger, mockUserLogger, cfg)(c)
			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")

			if tc.name == "Valid Credentials - Success" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Failed to unmarshal response")
				assert.NotEmpty(t, response["token"], "Token should not be empty")
			} else {
				assert.JSONEq(t, tc.expectedResponse, w.Body.String(), "Response body mismatch")
			}
		})
	}
}
