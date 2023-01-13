package integration_tests

import (
	"api-devices/api"
	"api-devices/api/keepalive"
	"api-devices/init_config"
	mqttClient "api-devices/mqtt-client"
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

var _ = Describe("KeepAlive", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var server *grpc.Server
	var listener net.Listener

	BeforeEach(func() {
		// 1. Init config
		logger = init_config.BuildConfig()
		defer logger.Sync()
		// 2. Init and start
		mqttClient.InitMqtt()
		// 3. Init and start gRPC server
		server, listener, ctx, _ = init_config.BuildServer(logger)
		go func() {
			server.Serve(listener)
		}()
	})

	AfterEach(func() {
		errLis := listener.Close()
		Expect(errLis).ShouldNot(HaveOccurred())

		server.Stop()
	})

	Context("calling keepalive grpc api", func() {
		It("should return success", func() {
			client := api.NewKeepAliveGrpc(ctx, logger)
			response, err := client.GetKeepAlive(ctx, &keepalive.StatusRequest{})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response).ShouldNot(BeNil())
			//fmt.Printf("response %#v\n", response)
		})
	})
})
