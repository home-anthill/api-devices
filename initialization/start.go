package initialization

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

// Start function
func Start() (*zap.SugaredLogger, *grpc.Server, net.Listener, context.Context, *mongo.Client) {
	// 1. Init logger
	logger := InitLogger()

	// 2. Init env
	InitEnv(logger)

	// 3. Init and start gRPC server
	server, listener, ctx, client := StartServer(logger)
	//logger.Infof("gRPC server listening at %v", listener.Addr())
	//if errGrpc := server.Serve(listener); errGrpc != nil {
	//  logger.Fatalf("gRPC server failed to serve: %v", errGrpc)
	//}

	return logger, server, listener, ctx, client
}
