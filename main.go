package main

import (
	"api-devices/initialization"
	"api-devices/mqttclient"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger, server, listener, mongoClient := initialization.Start()
	defer logger.Sync()

	logger.Info("MQTT starting...")
	if err := mqttclient.InitMqtt(logger); err != nil {
		logger.Fatalf("MQTT initialization failed: %v", err)
	}
	if token := mqttclient.Connect(); token.Wait() && token.Error() != nil {
		logger.Fatalf("MQTT connection failed: %v", token.Error())
	}
	logger.Info("MQTT running")

	// Start gRPC server in a goroutine
	go func() {
		logger.Infof("gRPC - starting server at %v", listener.Addr())
		if err := server.Serve(listener); err != nil {
			logger.Fatalf("gRPC server failed to serve: %v", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Infof("Received signal %v, shutting down...", sig)

	// Gracefully stop gRPC server
	server.GracefulStop()
	logger.Info("gRPC server stopped")

	// Disconnect MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := mongoClient.Disconnect(ctx); err != nil {
		logger.Errorf("MongoDB disconnect error: %v", err)
	} else {
		logger.Info("MongoDB disconnected")
	}
}
