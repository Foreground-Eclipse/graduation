package handlers

import (
	"errors"
	"net/http"

	"github.com/Foreground-Eclipse/transferer/internal/api/requests"
	jwt "github.com/Foreground-Eclipse/transferer/pkg/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BalanceGetter interface {
	GetUserBalance(username string) (map[string]float64, error)
}

// HandleBalance godoc
// @Summary Получение баланса пользователя
// @Description  Получает баланс пользователя на основе предоставленного токена.
// @Tags balance
// @Accept  json
// @Produce  json
// @Success 200 {object} requests.BalanceResponse "OK"
// @Failure 401 {object} requests.NotAuthorizedError "Не авторизован"
// @Router /api/v1/balance [get]
func HandleBalance(logger *zap.Logger, balanceGetter BalanceGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "api/v1/HandleBalance"

		logger.Info("proceeding new request", zap.String("op", op))

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			logError(c, logger, errors.New("not authorized"), http.StatusUnauthorized, "")
			return
		}

		logger.Info("request data: ",
			zap.String("discordid: ", c.Request.Method),
			zap.String("URL", c.Request.URL.String()),
			zap.String("Token", tokenString),
		)

		username, err := jwt.ValidateToken(tokenString)
		if err != nil {
			logError(c, logger, err, http.StatusUnauthorized, "")
			return
		}

		balance, err := balanceGetter.GetUserBalance(username)
		if err != nil {
			logError(c, logger, err, http.StatusOK, "")
			return
		}

		response := requests.BalanceResponse{
			Balance: balance,
		}

		c.JSON(http.StatusOK, response)

	}

}
