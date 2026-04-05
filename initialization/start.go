package initialization

import (
	"net"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Start initializes logger, environment, database, and gRPC server.
func Start() (*zap.SugaredLogger, *grpc.Server, net.Listener, *mongo.Client) {
	// 1. Init logger
	logger := InitLogger()

	// 2. Init env
	if err := InitEnv(logger); err != nil {
		logger.Fatalf("InitEnv failed: %v", err)
	}

	// 3. Init and start gRPC server
	server, listener, client := StartServer(logger)

	return logger, server, listener, client
}
