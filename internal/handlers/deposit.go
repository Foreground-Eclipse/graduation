package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Foreground-Eclipse/transferer/internal/api/requests"
	jwt "github.com/Foreground-Eclipse/transferer/pkg/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BalanceUpdater interface {
	UpdateUsersBalance(username, currency string, amount float64) error
	GetUserBalance(username string) (map[string]float64, error)
}

// HandleDeposit godoc
// @Summary Пополнение баланса пользователя
// @Description  Пополняет баланс пользователя на указанную сумму в указанной валюте.
// @Tags deposit
// @Accept  json
// @Produce  json
// @Param   request body requests.DepositRequest true "Данные для пополнения"
// @Success 200 {object} requests.DepositResponse "OK"
// @Failure 400 {object} requests.BadRequestError "Некорректный запрос"
// @Failure 401 {object} requests.NotAuthorizedError "Не авторизован"
// @Security ApiKeyAuth
// @Router /api/v1/deposit [post]
func HandleDeposit(logger *zap.Logger, balanceUpdater BalanceUpdater) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "api/v1/HandleDeposit"
		var req requests.DepositRequest

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
			logError(c, logger, errors.New("request contains wrong data"), http.StatusBadRequest, "failed to marshal request body")
			return
		}

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			logError(c, logger, errors.New("not authorized"), http.StatusUnauthorized, "")
			return
		}

		logger.Info("request data: ",
			zap.String("discordid: ", c.Request.Method),
			zap.String("URL", c.Request.URL.String()),
			zap.String("Token", tokenString),
			zap.String("body", string(reqBody)),
		)

		username, err := jwt.ValidateToken(tokenString)
		if err != nil {
			logError(c, logger, err, http.StatusUnauthorized, "")
			return
		}

		err = balanceUpdater.UpdateUsersBalance(username, req.Currency, float64(req.Amount))
		if err != nil {
			logError(c, logger, err, http.StatusOK, "")
			return
		}

		balance, err := balanceUpdater.GetUserBalance(username)
		if err != nil {
			logError(c, logger, err, http.StatusOK, "")
			return
		}

		response := requests.DepositResponse{
			Message: "Account topped up successfully",
			Balance: balance,
		}

		c.JSON(http.StatusOK, response)

	}

}
