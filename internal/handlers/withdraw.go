package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Foreground-Eclipse/transferer/internal/api/requests"
	jwt "github.com/Foreground-Eclipse/transferer/pkg/auth"
	"github.com/Foreground-Eclipse/transferer/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BalanceWithdrawer interface {
	UpdateUsersBalance(username, currency string, amount float64) error
	GetUserBalance(username string) (map[string]float64, error)
}

// HandleWithdraw godoc
// @Summary Снятие средств с баланса пользователя
// @Description  Снимает указанную сумму в указанной валюте с баланса пользователя.
// @Tags withdraw
// @Accept  json
// @Produce  json
// @Param   request body requests.WithdrawRequest true "Данные для снятия"
// @Success 200 {object} requests.DepositResponse "OK"
// @Failure 400 {object} requests.BadRequestError "Некорректный запрос"
// @Failure 401 {object} requests.NotAuthorizedError "Не авторизован"
// @Failure 403 {object} requests.NotEnoughFundsError "Недостаточно средств"
// @Security ApiKeyAuth
// @Router /api/v1/withdraw [post]
func HandleWithdraw(logger *zap.Logger, balanceWithdrawer BalanceWithdrawer) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "api/v1/HandleWithdraw"
		var req requests.WithdrawRequest

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

		balance, err := balanceWithdrawer.GetUserBalance(username)
		if err != nil {
			logError(c, logger, err, http.StatusOK, "")
			return
		}

		err = utils.VerifyWithdrawalAmount(req.Amount, balance[req.Currency])
		if err != nil {
			logError(c, logger, err, http.StatusOK, "")
			return
		}

		err = balanceWithdrawer.UpdateUsersBalance(username, req.Currency, float64(req.Amount*-1.0))
		if err != nil {
			logError(c, logger, err, http.StatusOK, "")
			return
		}

		balance, err = balanceWithdrawer.GetUserBalance(username)
		if err != nil {
			logError(c, logger, err, http.StatusOK, "")
			return
		}

		response := requests.DepositResponse{
			Message: "Withdrawal successfull",
			Balance: balance,
		}

		c.JSON(http.StatusOK, response)

	}

}
