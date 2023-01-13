package main

import (
	"api-devices/initialization"
	"api-devices/mqtt-client"
)

func main() {
	logger, server, listener, _, _ := initialization.Start()
	defer logger.Sync()

	logger.Info("MQTT starting...")
	mqtt_client.InitMqtt()
	if token := mqtt_client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	logger.Info("MQTT running")

	logger.Infof("gRPC - starting server at %v", listener.Addr())
	if errGrpc := server.Serve(listener); errGrpc != nil {
		logger.Fatalf("gRPC server failed to serve: %v", errGrpc)
	}
}
