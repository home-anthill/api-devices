package main

import (
	"api-devices/init_config"
	mqttClient "api-devices/mqtt-client"
)

func main() {
	// 1. Init config
	logger := init_config.BuildConfig()
	defer logger.Sync()

	// 2. Init and start
	mqttClient.InitMqtt()
	logger.Info("MQTT initialized")

	// 3. Init and start gRPC server
	server, listener, _, _ := init_config.BuildServer(logger)
	logger.Infof("gRPC server listening at %v", listener.Addr())
	if errGrpc := server.Serve(listener); errGrpc != nil {
		logger.Fatalf("gRPC server failed to serve: %v", errGrpc)
	}
}
