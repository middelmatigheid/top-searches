package main

import (
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/middelmatigheid/top-searches/internal/config"
	"github.com/middelmatigheid/top-searches/internal/consumer"
	"github.com/middelmatigheid/top-searches/internal/handler"
	"github.com/middelmatigheid/top-searches/internal/limiter"
	"github.com/middelmatigheid/top-searches/internal/prometheus"
	"github.com/middelmatigheid/top-searches/internal/server"
	"github.com/middelmatigheid/top-searches/internal/service"
	"github.com/middelmatigheid/top-searches/internal/stoplist"
	"github.com/middelmatigheid/top-searches/internal/storage"
	proto "github.com/middelmatigheid/top-searches/proto"

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

func runGRPC(service *service.Service, config *config.Config, logger *slog.Logger) {
	grpcServer := grpc.NewServer()

	proto.RegisterSearchesServer(grpcServer, server.NewSearchServer(service))
	proto.RegisterStoplistServer(grpcServer, server.NewStoplistServer(service))

	reflection.Register(grpcServer)

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

func runRest(service *service.Service, config *config.Config, metrics *prometheus.PrometheusMetrics, logger *slog.Logger) {
	handler := handler.NewHandler(service, logger)

	server := gin.Default()

	server.Use(corsMiddleware())
	server.Use(metrics.Middleware())

	server.GET("/metrics", gin.WrapH(promhttp.Handler()))

	server.GET("/top-searches/get/:n", handler.GetTopN)
	server.POST("/stoplist/add/:word", handler.AddStoplistWord)
	server.DELETE("/stoplist/remove/:word", handler.RemoveStoplistWord)
	server.GET("/stoplist/get", handler.GetStoplist)

	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	logger.Info("Running the server", "port", config.Port)
	if err := server.Run(":" + config.Port); err != nil {
		logger.Error("Error while running the server", "error", err, "port", config.Port)
		return
	}
}

func main() {
	config, err := config.GetConfig("config.yaml")
	if err != nil {
		slog.Error("Error while getting config", "error", err)
		return
	}

	logger, err := getLogger(config.LogFile)
	if err != nil {
		slog.Error("Error while getting logger", "error", err)
		return
	}

	logger.Info("Starting the server")

	storage, err := storage.NewStorage(config.Timespan, config.BucketTimespan, logger)
	if err != nil {
		logger.Error("Error while getting storage", "error", err)
		return
	}

	stoplist := stoplist.NewStoplist(config.Stoplist, logger)

	limiter, err := limiter.NewLimiter(config.Cooldown)
	if err != nil {
		logger.Error("Error while getting limiter", "error", err)
		return
	}

	metrics := prometheus.NewPrometheusMetrics()

	service := service.NewService(config.Timespan, storage, stoplist, limiter, metrics)

	// Waiting for kafka to start
	time.Sleep(10 * time.Second)

	consumer, err := consumer.NewConsumer(config.Brokers, config.Topic, service, logger)
	if err != nil {
		logger.Error("Error while getting consumer", "error", err)
		return
	}

	if err := consumer.Start(); err != nil {
		logger.Error("Error while starting consumer", "error", err)
		return
	}

	if config.GRPC {
		runGRPC(service, config, logger)
	} else {
		runRest(service, config, metrics, logger)
	}
}
