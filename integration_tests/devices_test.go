package integration_tests

import (
	"api-devices/api"
	device2 "api-devices/api/device"
	"api-devices/init_config"
	"api-devices/models"
	mqttClient "api-devices/mqtt-client"
	"api-devices/test_utils"
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"time"
)

var _ = Describe("Devices", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var collectionACs *mongo.Collection
	var server *grpc.Server
	var listener net.Listener

	var device = models.AirConditioner{
		ID:             primitive.NewObjectID(),
		UUID:           "65a24635-abb8-418c-ba35-0c0ed30aeefe",
		Mac:            "11:22:33:44:55:66",
		Name:           "ac-beko",
		Manufacturer:   "ks89",
		Model:          "ac-beko",
		ProfileOwnerId: primitive.NewObjectID().Hex(),
		ApiToken:       "473a4861-632b-4915-b01e-cf1d418966c6",
		Status:         models.Status{},
		CreatedAt:      time.Time{},
		ModifiedAt:     time.Time{},
	}

	BeforeEach(func() {
		// 1. Init config
		logger = init_config.BuildConfig()
		defer logger.Sync()
		// 2. Init and start
		mqttClient.InitMqtt()
		// 3. Init and start gRPC server
		server, listener, ctx, collectionACs = init_config.BuildServer(logger)
		go func() {
			server.Serve(listener)
		}()
	})

	AfterEach(func() {
		errLis := listener.Close()
		Expect(errLis).ShouldNot(HaveOccurred())

		server.Stop()

		test_utils.DropAllCollections(ctx, collectionACs)
	})

	Context("calling devices grpc api", func() {
		It("should return success", func() {
			err := test_utils.InsertOne(ctx, collectionACs, device)
			Expect(err).ShouldNot(HaveOccurred())

			status := models.Status{
				On:          true,
				Temperature: 28,
				Mode:        1,
				FanSpeed:    2,
			}

			client := api.NewDevicesGrpc(ctx, logger, collectionACs)
			responseSet, err := client.SetValues(ctx, &device2.ValuesRequest{
				Id:          device.ID.Hex(),
				Uuid:        device.UUID,
				Mac:         device.Mac,
				ApiToken:    device.ApiToken,
				On:          status.On,
				Temperature: int32(status.Temperature),
				Mode:        int32(status.Mode),
				FanSpeed:    int32(status.FanSpeed),
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseSet.GetStatus()).To(Equal("200"))
			Expect(responseSet.GetMessage()).To(Equal("Updated"))

			ac, err := test_utils.FindOneById[models.AirConditioner](ctx, collectionACs, device.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ac.ID).To(Equal(device.ID))
			Expect(ac.UUID).To(Equal(device.UUID))
			Expect(ac.Mac).To(Equal(device.Mac))
			Expect(ac.Name).To(Equal(device.Name))
			Expect(ac.Manufacturer).To(Equal(device.Manufacturer))
			Expect(ac.Model).To(Equal(device.Model))
			Expect(ac.ProfileOwnerId).To(Equal(device.ProfileOwnerId))
			Expect(ac.ApiToken).To(Equal(device.ApiToken))
			Expect(ac.Status).To(Equal(status))

			responseGet, err := client.GetStatus(ctx, &device2.StatusRequest{
				Id:       device.ID.Hex(),
				Uuid:     device.UUID,
				Mac:      device.Mac,
				ApiToken: device.ApiToken,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseGet.GetOn()).To(Equal(status.On))
			Expect(responseGet.GetTemperature()).To(Equal(int32(status.Temperature)))
			Expect(responseGet.GetMode()).To(Equal(int32(status.Mode)))
			Expect(responseGet.GetFanSpeed()).To(Equal(int32(status.FanSpeed)))

			ac, err = test_utils.FindOneById[models.AirConditioner](ctx, collectionACs, device.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ac.Status).To(Equal(status))
		})
	})
})
