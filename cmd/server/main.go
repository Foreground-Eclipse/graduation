package main

import (
	"fmt"
	"os"

	pb "github.com/Foreground-Eclipse/grpcexchanger/proto"
	"github.com/Foreground-Eclipse/transferer/config"
	_ "github.com/Foreground-Eclipse/transferer/docs"
	"github.com/Foreground-Eclipse/transferer/internal/handlers"
	"github.com/Foreground-Eclipse/transferer/internal/storage/postgres"
	"github.com/Foreground-Eclipse/transferer/pkg/logger"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title Transferer API
// @version 1.0
// @description This is a sample server for transferer service
// @host localhost:8088
// @BasePath /api/v1
// @schemes http https
func main() {
	cfg := config.MustLoad("config")

	log := logger.SetupLogger()
	if log == nil {
		fmt.Println("Logger wasnt initialized")
	}

	log.Info("Info message in main", zap.String("database host: ", cfg.Database.Host))

	storage, err := postgres.New(cfg)
	if err != nil {
		panic(err)
	}

	err = storage.InitExchangeRatesSchema()
	if err != nil {
		panic(err)
	}

	err = storage.InitUserSchema()
	if err != nil {
		panic(err)
	}

	err = storage.InitWalletSchema()
	if err != nil {
		panic(err)
	}
	exchangerHost := os.Getenv("EXCHANGER_HOST")
	conn, err := grpc.Dial(exchangerHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewExchangeServiceClient(conn)

	router := gin.Default()

	router.POST("/api/v1/register", handlers.HandleRegisterUser(log, storage))
	router.POST("/api/v1/login", handlers.HandleLoginUser(log, storage, cfg))
	router.GET("/api/v1/balance", handlers.HandleBalance(log, storage))
	router.POST("/api/v1/wallet/deposit", handlers.HandleDeposit(log, storage))
	router.GET("/api/v1/exchange/rates", handlers.HandleRates(log, client))
	router.POST("/api/v1/exchange", handlers.HandleExchange(log, client, storage))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.Run(cfg.Server.Address)
}
