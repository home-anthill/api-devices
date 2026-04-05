package initialization

import (
	"api-devices/api"
	pbd "api-devices/api/device"
	pbr "api-devices/api/register"
	"api-devices/db"
	"net"
	"os"
	"path/filepath"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// StartServer initializes the database, gRPC server, and network listener.
func StartServer(logger *zap.SugaredLogger) (*grpc.Server, net.Listener, *mongo.Client) {
	// Connect to DB
	client := db.InitDb(logger)

	// Instantiate gRPC and apply some middlewares
	logger.Info("StartServer - gRPC - Initializing...")

	// Create gRPC API instances
	registerGrpc := api.NewRegisterGrpc(logger, client)
	devicesGrpc := api.NewDevicesGrpc(logger, client)

	// Create new gRPC server with (blank) options
	var server *grpc.Server
	if os.Getenv("GRPC_TLS") == "true" {
		certFolder := os.Getenv("CERT_FOLDER_PATH")
		creds, credErr := credentials.NewServerTLSFromFile(
			filepath.Join(certFolder, "server-cert.pem"),
			filepath.Join(certFolder, "server-key.pem"),
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

	// Register standard health check server using grpc_health_v1 package
	hs := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, hs)

	// Register the service with the server
	pbr.RegisterRegistrationServer(server, registerGrpc)
	pbd.RegisterDeviceServer(server, devicesGrpc)

	// Start gRPC listener
	grpcURL := os.Getenv("GRPC_URL")
	listener, errGrpc := net.Listen("tcp", grpcURL)
	if errGrpc != nil {
		logger.Fatalf("StartServer - failed to listen: %v", errGrpc)
	}
	logger.Infof("StartServer - gRPC client listening at %s", listener.Addr().String())

	return server, listener, client
}
