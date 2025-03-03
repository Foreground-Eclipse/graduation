package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Foreground-Eclipse/transferer/config"
	"github.com/Foreground-Eclipse/transferer/internal/api/requests"
	"github.com/Foreground-Eclipse/transferer/internal/middleware"
	jwt "github.com/Foreground-Eclipse/transferer/pkg/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserLogger interface {
	GetUsersPassHash(username string) (string, error)
}

// HandleLoginUser godoc
// @Summary Аутентификация пользователя
// @Description Аутентифицирует пользователя и возвращает JWT токен.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.LoginRequest true "Логин пользователя"
// @Success 200 {object} map[string]interface{} "Успешная аутентификация"
// @Failure 400 {object} requests.BadRequestError "Некорректный запрос"
// @Failure 500 {object} requests.CantCreateJWTError "Внутренняя ошибка сервера"
// @Router /api/v1/login [post]
func HandleLoginUser(logger *zap.Logger, userLogger UserLogger, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.LoginRequest
		const op = "api/v1/HandleLoginUser"

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

		passhash, err := userLogger.GetUsersPassHash(req.Username)
		if err != nil {
			logError(c, logger, errors.New("wrong password"), http.StatusOK, "wrong password")
			return
		}

		isRight, err := middleware.VerifyPassword(req.Password, passhash)
		if err != nil {
			logError(c, logger, errors.New("wrong password"), http.StatusOK, "wrong password")
			return
		}

		if !isRight {
			logError(c, logger, errors.New("wrong password"), http.StatusOK, "wrong password")
			return
		}

		token, err := jwt.GenerateJWT(req.Username, cfg)
		if err != nil {
			logError(c, logger, errors.New("could not create JWT token"), http.StatusInternalServerError, "")
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
		})

	}

}
