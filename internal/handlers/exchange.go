package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	pb "github.com/Foreground-Eclipse/grpcexchanger/proto"
	"github.com/Foreground-Eclipse/transferer/internal/api/requests"
	jwt "github.com/Foreground-Eclipse/transferer/pkg/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Exchanger interface {
	GetUserBalance(username string) (map[string]float64, error)
	UpdateUsersBalance(username, currency string, amount float64) error
}

// HandleExchange godoc
// @Summary Обмен валюты пользователя
// @Description  Выполняет обмен валюты пользователя из одной валюты в другую.
// @Tags exchange
// @Accept  json
// @Produce  json
// @Param   request body requests.ExchangeRequest true "Данные для обмена"
// @Success 200 {object} requests.ExchangeResponse "OK"
// @Failure 400 {object} requests.BadRequestError "Некорректный запрос"
// @Failure 401 {object} requests.NotAuthorizedError "Не авторизован"
// @Security ApiKeyAuth
// @Router /api/v1/exchange [post]
func HandleExchange(logger *zap.Logger, client pb.ExchangeServiceClient, exchanger Exchanger) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "api/v1/HandleExchange"

		var req requests.ExchangeRequest

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
			zap.String("Body", string(reqBody)),
		)

		username, err := jwt.ValidateToken(tokenString)
		if err != nil {
			logError(c, logger, err, http.StatusUnauthorized, "")
			return
		}

		ctx := context.Background()
		request := &pb.Empty{}
		response, err := client.GetExchangeRates(ctx, request)
		if err != nil {
			logError(c, logger, errors.New("failed to retrieve exchange rates"), http.StatusInternalServerError, "")
			return
		}

		balance, err := exchanger.GetUserBalance(username)
		if err != nil {
			logError(c, logger, err, http.StatusInternalServerError, "")
			return
		}
		if balance[req.FromCurrency] < req.Amount {
			logError(c, logger, errors.New("not enough money"), http.StatusOK, "")
			return
		}

		err = exchanger.UpdateUsersBalance(username, req.FromCurrency, req.Amount*-1.0)
		if err != nil {
			logError(c, logger, err, http.StatusInternalServerError, "")
			return
		}

		toAdd, err := convertCurrencies(req.FromCurrency, req.ToCurrency, req.Amount, response)
		if err != nil {
			logError(c, logger, err, http.StatusInternalServerError, "")
			return
		}

		err = exchanger.UpdateUsersBalance(username, req.ToCurrency, toAdd)
		if err != nil {
			logError(c, logger, err, http.StatusInternalServerError, "")
			return
		}

		var resp requests.ExchangeResponse
		resp.ExchangedAmount = toAdd
		resp.Message = "exchanged successfully"
		resp.NewBalance, err = exchanger.GetUserBalance(username)
		if err != nil {
			logError(c, logger, err, http.StatusInternalServerError, "")
			return
		}

		c.JSON(http.StatusOK, resp)
	}

}

func convertCurrencies(fromCurrency, toCurrency string, amount float64, response *pb.ExchangeRatesResponse) (float64, error) {
	fromRateKey := fmt.Sprintf("RUB_%s", fromCurrency)
	fromRate, fromOK := response.Rates[fromRateKey]
	if !fromOK {
		return 0, fmt.Errorf("курс для валюты %s не найден", fromCurrency)
	}

	toRateKey := fmt.Sprintf("RUB_%s", toCurrency)
	toRate, toOK := response.Rates[toRateKey]
	if !toOK {
		return 0, fmt.Errorf("курс для валюты %s не найден", toCurrency)
	}

	amountInRUB := amount / float64(fromRate)

	amountInToCurrency := amountInRUB * float64(toRate)

	return amountInToCurrency, nil
}
