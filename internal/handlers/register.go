package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Foreground-Eclipse/transferer/internal/api/requests"
	"github.com/Foreground-Eclipse/transferer/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserRegisterer interface {
	RegisterUser(username, passwordHash, email string) error
	DoesUserExists(username string) (bool, error)
	DoesEmailExists(email string) (bool, error)
}

type EmailAlreadyExistsError struct {
	Message string
}

type UsernameAlreadyExists struct {
	Message string
}

func (e *EmailAlreadyExistsError) Error() string {
	return e.Message
}

func (e *UsernameAlreadyExists) Error() string {
	return e.Message
}

// HandleRegisterUser godoc
// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя в системе.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.RegisterRequest true "Данные для регистрации"
// @Success 200 {object} map[string]interface{} "Успешная регистрация"
// @Failure 400 {object} map[string]interface{} "Некорректный запрос"
// @Failure 409 {object} map[string]interface{} "Конфликт (имя пользователя или email уже существует)"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/v1/register [post]
func HandleRegisterUser(logger *zap.Logger, userRegisterer UserRegisterer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.RegisterRequest
		const op = "api/v1/HandleRegisterUser"

		logger.Info("proceeding new request", zap.String("op", op))

		if err := c.BindJSON(&req); err != nil {
			if errors.Is(err, io.EOF) {
				logError(c, logger, errors.New("empty json"), http.StatusBadRequest, "failed to process request")
			}
			logError(c, logger, errors.New("request contains wrong data"), http.StatusBadRequest, "failed to process request")
			return
		}
		reqBody, err := json.Marshal(req)
		if err != nil {
			logError(c, logger, err, http.StatusBadRequest, "failed to marshal request body")
			return
		}

		logger.Info("request data: ",
			zap.String("discordid: ", c.Request.Method),
			zap.String("URL", c.Request.URL.String()),
			zap.String("body", string(reqBody)),
		)

		err = verifyCredentials(req.Email, req.Username, userRegisterer)

		var emailErr *EmailAlreadyExistsError
		var usernameErr *UsernameAlreadyExists

		if errors.As(err, &usernameErr) {
			c.JSON(http.StatusOK, map[string]interface{}{
				"status": "error",
				"error":  "username already exists",
			})
			return
		}

		if errors.As(err, &emailErr) {
			c.JSON(http.StatusOK, map[string]interface{}{
				"status": "error",
				"error":  "email already exists",
			})
			return
		}

		passHash := middleware.HashPassword(req.Password)

		err = userRegisterer.RegisterUser(req.Username, passHash, req.Email)
		if err != nil {
			logError(c, logger, err, http.StatusInternalServerError, "failed to add the user")
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{
			"message": "user registered",
		})

	}
}

func logError(c *gin.Context, logger *zap.Logger, err error, status int, message string) {
	logger.Warn(message, zap.Error(err))
	c.JSON(status, map[string]interface{}{
		"status": "error",
		"error":  err.Error(),
	})
}

func verifyCredentials(email, username string, verifier UserRegisterer) error {
	ok, err := verifier.DoesEmailExists(email)
	if err != nil {
		return err
	}
	if ok {
		return &EmailAlreadyExistsError{Message: "email is taken"}
	}
	ok, err = verifier.DoesUserExists(username)
	if err != nil {
		return err
	}
	if ok {
		return &UsernameAlreadyExists{Message: "username is taken"}
	}
	return nil
}
