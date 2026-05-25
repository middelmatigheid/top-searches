package main

import (
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/middelmatigheid/top-searches/producer/internal/config"
	"github.com/middelmatigheid/top-searches/producer/internal/handler"
	"github.com/middelmatigheid/top-searches/producer/internal/models"
	"github.com/middelmatigheid/top-searches/producer/internal/server"
	"github.com/middelmatigheid/top-searches/producer/internal/service"
	proto "github.com/middelmatigheid/top-searches/producer/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func getLogger(path string) (*slog.Logger, error) {
	var writer io.Writer

	if path != "" {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		writer = io.MultiWriter(os.Stdout, file)
	}

	opts := slog.HandlerOptions{Level: slog.LevelInfo}
	return slog.New(slog.NewJSONHandler(writer, &opts)), nil
}

func createProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	return sarama.NewSyncProducer(brokers, config)
}

func startAutoSender(service *service.Service, config *config.Config, logger *slog.Logger) {
	go func() {
		logger.Info("Auto sender is on")
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			search := config.Searches[rand.IntN(len(config.Searches))]
			user := config.Users[rand.IntN(len(config.Users))]
			request := models.Search{
				Search: search,
				User:   user,
			}
			service.SendSearch(request)
		}
	}()
}

func runGRPC(service *service.Service, config *config.Config, logger *slog.Logger) {
	grpcServer := grpc.NewServer()

	proto.RegisterProducerServer(grpcServer, server.NewSearchServer(service))

	reflection.Register(grpcServer)

	if config.AutoSender {
		startAutoSender(service, config, logger)
	}

	lis, err := net.Listen("tcp", "0.0.0.0:"+config.PortGRPC)
	if err != nil {
		logger.Error("Failed to listen", "error", err)
		return
	}

	logger.Info("gRPC server listening", "port", config.PortGRPC)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("Failed to serve", "error", err)
	}
}

func runRest(service *service.Service, config *config.Config, logger *slog.Logger) {
	handler := handler.NewHandler(service, logger)

	if config.AutoSender {
		startAutoSender(service, config, logger)
	}

	server := gin.Default()
	server.Use(corsMiddleware())
	server.POST("/search", handler.Search)
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	logger.Info("Running the producer", "port", config.Port)
	if err := server.Run(":" + config.Port); err != nil {
		logger.Error("Error while running the server", "error", err, "port", config.Port)
		return
	}
}

func main() {
	config, err := config.GetConfig("config.yaml")
	if err != nil {
		slog.Error("Error while getting producer config", "error", err)
		return
	}
	logger, err := getLogger(config.LogFile)
	if err != nil {
		slog.Error("Error while getting producer logger", "error", err)
		return
	}

	logger.Info("Starting the producer")

	// Waiting for kafka to start
	time.Sleep(10 * time.Second)

	producer, err := createProducer(config.Brokers)
	if err != nil {
		logger.Error("Error while getting producer", "error", err)
		return
	}

	service := service.NewService(producer, config.Topic)

	if config.GRPC {
		runGRPC(service, config, logger)
	} else {
		runRest(service, config, logger)
	}
}
