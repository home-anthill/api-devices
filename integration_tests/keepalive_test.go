package integration_tests

import (
	"api-devices/api"
	"api-devices/api/keepalive"
	"api-devices/initialization"
	mqtt_client "api-devices/mqttclient"
	"api-devices/testutils"
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
		logger, server, listener, _, _ = initialization.Start()
		defer logger.Sync()

		// create and start a mocked MQTT client
		mqtt_client.SetMqttClient(testutils.NewMockClient())
		if token := mqtt_client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		logger.Infof("gRPC server listening at %v", listener.Addr())
		go func() {
			_ = server.Serve(listener)
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
