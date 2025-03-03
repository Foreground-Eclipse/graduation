package handlers

import (
	"context"
	"errors"
	"net/http"

	pb "github.com/Foreground-Eclipse/grpcexchanger/proto"
	"github.com/Foreground-Eclipse/transferer/internal/api/requests"
	jwt "github.com/Foreground-Eclipse/transferer/pkg/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HandleRates godoc
// @Summary Получение курсов валют
// @Description  Получает текущие курсы валют.
// @Tags rates
// @Accept  json
// @Produce  json
// @Success 200 {object} requests.RatesResponse "OK"
// @Failure 401 {object} requests.NotAuthorizedError "Не авторизован"
// @Failure 500 {object} requests.RetrieveRatesError "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/v1/rates [get]
func HandleRates(logger *zap.Logger, client pb.ExchangeServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "api/v1/HandleRates"

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

		_, err := jwt.ValidateToken(tokenString)
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

		var rates requests.RatesResponse
		rates.Rates = response.Rates

		c.JSON(http.StatusOK, rates)

	}

}
