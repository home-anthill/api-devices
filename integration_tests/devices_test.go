package integration_tests

import (
	"api-devices/api"
	device2 "api-devices/api/device"
	"api-devices/initialization"
	"api-devices/models"
	mqtt_client "api-devices/mqtt-client"
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
		ProfileOwnerId: primitive.NewObjectID(),
		ApiToken:       "473a4861-632b-4915-b01e-cf1d418966c6",
		Status:         models.Status{},
		CreatedAt:      time.Time{},
		ModifiedAt:     time.Time{},
	}

	BeforeEach(func() {
		logger, server, listener, ctx, collectionACs = initialization.Start()
		defer logger.Sync()

		// create and start a mocked MQTT client
		mqtt_client.SetMqttClient(test_utils.NewMockClient())
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

		test_utils.DropAllCollections(ctx, collectionACs)
	})

	Context("calling devices grpc api", func() {
		It("should setValues of an existing device and get those values via getValues", func() {
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

		When("getValues", func() {
			It("should return an error, because device doesn't exist on db", func() {
				missingMacDevice := "99:99:99:99:99:99"
				client := api.NewDevicesGrpc(ctx, logger, collectionACs)
				_, err := client.GetStatus(ctx, &device2.StatusRequest{
					Id:       device.ID.Hex(),
					Uuid:     device.UUID,
					Mac:      missingMacDevice,
					ApiToken: device.ApiToken,
				})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal("cannot find device with mac " + missingMacDevice))
			})
		})

		When("setValues", func() {
			It("should return an error, because device doesn't exist on db", func() {
				status := models.Status{
					On:          true,
					Temperature: 28,
					Mode:        1,
					FanSpeed:    2,
				}
				missingMacDevice := "99:99:99:99:99:99"
				client := api.NewDevicesGrpc(ctx, logger, collectionACs)
				_, err := client.SetValues(ctx, &device2.ValuesRequest{
					Id:          device.ID.Hex(),
					Uuid:        device.UUID,
					Mac:         missingMacDevice,
					ApiToken:    device.ApiToken,
					On:          status.On,
					Temperature: int32(status.Temperature),
					Mode:        int32(status.Mode),
					FanSpeed:    int32(status.FanSpeed),
				})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("cannot find a unique device with mac " + missingMacDevice))
			})
		})
	})
})
