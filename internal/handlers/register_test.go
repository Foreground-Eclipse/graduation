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

type MockUserRegisterer struct {
	userExists         bool
	emailExists        bool
	registerErr        error
	doesUserExistsErr  error
	doesEmailExistsErr error
}

func (m *MockUserRegisterer) RegisterUser(username, passwordHash, email string) error {
	return m.registerErr
}

func (m *MockUserRegisterer) DoesUserExists(username string) (bool, error) {
	return m.userExists, m.doesUserExistsErr
}

func (m *MockUserRegisterer) DoesEmailExists(email string) (bool, error) {
	return m.emailExists, m.doesEmailExistsErr
}

func TestHandleRegisterUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name                   string
		requestBody            string
		mockUserExists         bool
		mockEmailExists        bool
		mockRegisterErr        error
		mockDoesUserExistsErr  error
		mockDoesEmailExistsErr error
		expectedStatus         int
		expectedResponse       string
	}{
		{
			name:                   "Invalid JSON - Wrong Type",
			requestBody:            `{"username":123,"password":"password","email":"test@example.com"}`,
			mockUserExists:         false,
			mockEmailExists:        false,
			mockRegisterErr:        nil,
			mockDoesUserExistsErr:  nil,
			mockDoesEmailExistsErr: nil,
			expectedStatus:         http.StatusBadRequest,
			expectedResponse:       `{"error":"request contains wrong data", "status":"error"}`,
		},
		{
			name:                   "Username Already Exists",
			requestBody:            `{"username":"testuser","password":"password","email":"test@example.com"}`,
			mockUserExists:         true,
			mockEmailExists:        false,
			mockRegisterErr:        nil,
			mockDoesUserExistsErr:  nil,
			mockDoesEmailExistsErr: nil,
			expectedStatus:         http.StatusOK,
			expectedResponse:       `{"error":"username already exists", "status":"error"}`,
		},
		{
			name:                   "Email Already Exists",
			requestBody:            `{"username":"testuser","password":"password","email":"test@example.com"}`,
			mockUserExists:         false,
			mockEmailExists:        true,
			mockRegisterErr:        nil,
			mockDoesUserExistsErr:  nil,
			mockDoesEmailExistsErr: nil,
			expectedStatus:         http.StatusOK,
			expectedResponse:       `{"error":"email already exists", "status":"error"}`,
		},
		{
			name:                   "Registration Success",
			requestBody:            `{"username":"testuser","password":"password","email":"test@example.com"}`,
			mockUserExists:         false,
			mockEmailExists:        false,
			mockRegisterErr:        nil,
			mockDoesUserExistsErr:  nil,
			mockDoesEmailExistsErr: nil,
			expectedStatus:         http.StatusOK,
			expectedResponse:       `{"message":"user registered"}`,
		},
		{
			name:                   "Registration Error",
			requestBody:            `{"username":"testuser","password":"password","email":"test@example.com"}`,
			mockUserExists:         false,
			mockEmailExists:        false,
			mockRegisterErr:        errors.New("database error"),
			mockDoesUserExistsErr:  nil,
			mockDoesEmailExistsErr: nil,
			expectedStatus:         http.StatusInternalServerError,
			expectedResponse:       `{"error":"database error", "status":"error"}`,
		},
		{
			name:                   "DoesUserExists Error",
			requestBody:            `{"username":"testuser","password":"password","email":"test@example.com"}`,
			mockUserExists:         false,
			mockEmailExists:        false,
			mockRegisterErr:        nil,
			mockDoesUserExistsErr:  errors.New("user check error"),
			mockDoesEmailExistsErr: nil,
			expectedStatus:         http.StatusOK,
			expectedResponse:       `{"error":"user check error", "status":"error"}`,
		},
		{
			name:                   "DoesEmailExists Error",
			requestBody:            `{"username":"testuser","password":"password","email":"test@example.com"}`,
			mockUserExists:         false,
			mockEmailExists:        false,
			mockRegisterErr:        nil,
			mockDoesUserExistsErr:  nil,
			mockDoesEmailExistsErr: errors.New("email check error"),
			expectedStatus:         http.StatusOK,
			expectedResponse:       `{"error":"email check error", "status":"error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body := bytes.NewBufferString(tc.requestBody)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/register", body)
			c.Request.Header.Set("Content-Type", "application/json")

			logger := newTestLogger()

			mockUserRegisterer := &MockUserRegisterer{
				userExists:         tc.mockUserExists,
				emailExists:        tc.mockEmailExists,
				registerErr:        tc.mockRegisterErr,
				doesUserExistsErr:  tc.mockDoesUserExistsErr,
				doesEmailExistsErr: tc.mockDoesEmailExistsErr,
			}

			HandleRegisterUser(logger, mockUserRegisterer)(c)

			assert.Equal(t, tc.expectedStatus, w.Code, "Status code mismatch")
			assert.JSONEq(t, tc.expectedResponse, w.Body.String(), "Response body mismatch")
		})
	}
}
