package initialization

import (
	"api-devices/api"
	pbd "api-devices/api/device"
	pbk "api-devices/api/keepalive"
	pbr "api-devices/api/register"
	"api-devices/db"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
)

// StartServer function
func StartServer(logger *zap.SugaredLogger) (*grpc.Server, net.Listener, context.Context, *mongo.Collection) {
	// Initialization
	ctx := context.Background()

	// Connect to DB
	collectionACs := db.InitDb(ctx, logger)

	// Instantiate gRPC and apply some middlewares
	logger.Info("StartServer - gRPC - Initializing...")

	// Create gRPC API instances
	registerGrpc := api.NewRegisterGrpc(ctx, logger, collectionACs)
	devicesGrpc := api.NewDevicesGrpc(ctx, logger, collectionACs)
	keepAliveGrpc := api.NewKeepAliveGrpc(ctx, logger)

	// Create new gRPC server with (blank) options
	var server *grpc.Server
	if os.Getenv("GRPC_TLS") == "true" {
		creds, credErr := credentials.NewServerTLSFromFile(
			os.Getenv("CERT_FOLDER_PATH")+"/server-cert.pem",
			os.Getenv("CERT_FOLDER_PATH")+"/server-key.pem",
		)
		if credErr != nil {
			logger.Fatalf("StartServer - NewServerTLSFromFile error %v", credErr)
		}
		logger.Info("StartServer - gRPC TLS security enabled")
		server = grpc.NewServer(grpc.Creds(creds))
	} else {
		logger.Info("StartServer - gRPC TLS security not enabled")
		server = grpc.NewServer()
	}

	// Register the service with the server
	pbr.RegisterRegistrationServer(server, registerGrpc)
	pbd.RegisterDeviceServer(server, devicesGrpc)
	pbk.RegisterKeepAliveServer(server, keepAliveGrpc)

	// Start gRPC listener
	grpcURL := os.Getenv("GRPC_URL")
	listener, errGrpc := net.Listen("tcp", grpcURL)
	if errGrpc != nil {
		logger.Fatalf("StartServer - failed to listen: %v", errGrpc)
	}
	logger.Info("StartServer - gRPC client listening at " + listener.Addr().String())

	return server, listener, ctx, collectionACs
}
