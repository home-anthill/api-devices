package integration_tests

import (
	"api-devices/api"
	device2 "api-devices/api/device"
	"api-devices/db"
	"api-devices/initialization"
	"api-devices/models"
	mqtt_client "api-devices/mqttclient"
	"api-devices/testutils"
	"context"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ = Describe("Devices", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var client *mongo.Client
	var airConditionerCollection *mongo.Collection
	var setpointCollection *mongo.Collection
	var toleranceCollection *mongo.Collection
	var server *grpc.Server
	var listener net.Listener

	var airConditioner = models.Device{
		ID:             primitive.NewObjectID(),
		UUID:           "65a24635-abb8-418c-ba35-0c0ed30aeefe",
		Mac:            "11:22:33:44:55:66",
		Name:           "ac-beko",
		Manufacturer:   "ks89",
		Model:          "ac-beko",
		ProfileOwnerID: primitive.NewObjectID(),
		APIToken:       "473a4861-632b-4915-b01e-cf1d418966c6",
		Status:         models.Status{},
		CreatedAt:      time.Time{},
		ModifiedAt:     time.Time{},
	}
	var setpoint = models.Device{
		ID:             primitive.ObjectID{},
		UUID:           "65a24635-abb8-418c-ba35-0c0ed30aeecc",
		Mac:            "11:22:33:44:55:77",
		Name:           "setpoint",
		Manufacturer:   "ks89",
		Model:          "setpoint",
		ProfileOwnerID: primitive.NewObjectID(),
		APIToken:       "473a4861-632b-4915-b01e-cf1d418966c6",
		Status:         models.Status{},
		CreatedAt:      time.Time{},
		ModifiedAt:     time.Time{},
	}

	var tolerance = models.Device{
		ID:             primitive.ObjectID{},
		UUID:           "65a24635-abb8-418c-ba35-0c0ed30aeedd",
		Mac:            "11:22:33:44:55:88",
		Name:           "tolerance",
		Manufacturer:   "ks89",
		Model:          "tolerance",
		ProfileOwnerID: primitive.NewObjectID(),
		APIToken:       "473a4861-632b-4915-b01e-cf1d418966c6",
		Status:         models.Status{},
		CreatedAt:      time.Time{},
		ModifiedAt:     time.Time{},
	}

	BeforeEach(func() {
		logger, server, listener, ctx, client = initialization.Start()
		defer logger.Sync()

		airConditionerCollection = db.GetCollections(client).AirConditioners
		setpointCollection = db.GetCollections(client).Setpoints
		toleranceCollection = db.GetCollections(client).Tolerances

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

		testutils.DropAllCollections(ctx, airConditionerCollection)
		testutils.DropAllCollections(ctx, setpointCollection)
		testutils.DropAllCollections(ctx, toleranceCollection)
	})

	Context("calling devices grpc api", func() {
		It("should setValues of an existing airconditioner and get those values via getValues", func() {
			err := testutils.InsertOne(ctx, airConditionerCollection, airConditioner)
			Expect(err).ShouldNot(HaveOccurred())

			status := models.Status{
				Value: 20,
			}

			client := api.NewDevicesGrpc(ctx, logger, client)
			responseSet, err := client.SetValue(ctx, &device2.SetValueRequest{
				Id:          airConditioner.ID.Hex(),
				FeatureUuid: airConditioner.UUID,
				FeatureName: airConditioner.Name,
				Mac:         airConditioner.Mac,
				Value:       status.Value,
				ApiToken:    airConditioner.APIToken,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseSet.GetStatus()).To(Equal("200"))
			Expect(responseSet.GetMessage()).To(Equal("Updated"))

			controller, err := testutils.FindOneById[models.Device](ctx, airConditionerCollection, airConditioner.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(controller.ID).To(Equal(airConditioner.ID))
			Expect(controller.UUID).To(Equal(airConditioner.UUID))
			Expect(controller.Mac).To(Equal(airConditioner.Mac))
			Expect(controller.Name).To(Equal(airConditioner.Name))
			Expect(controller.Manufacturer).To(Equal(airConditioner.Manufacturer))
			Expect(controller.Model).To(Equal(airConditioner.Model))
			Expect(controller.ProfileOwnerID).To(Equal(airConditioner.ProfileOwnerID))
			Expect(controller.APIToken).To(Equal(airConditioner.APIToken))
			Expect(controller.Status.Value).To(Equal(status.Value))

			responseGet, err := client.GetValue(ctx, &device2.GetValueRequest{
				Id:          airConditioner.ID.Hex(),
				FeatureUuid: airConditioner.UUID,
				FeatureName: airConditioner.Name,
				Mac:         airConditioner.Mac,
				ApiToken:    airConditioner.APIToken,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseGet.GetValue()).To(Equal(status.Value))

			controller, err = testutils.FindOneById[models.Device](ctx, airConditionerCollection, airConditioner.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(controller.Status.Value).To(Equal(status.Value))
		})

		It("should setValues of an existing setpoint and get those values via getValues", func() {
			err := testutils.InsertOne(ctx, setpointCollection, setpoint)
			Expect(err).ShouldNot(HaveOccurred())

			status := models.Status{
				Value: 20,
			}

			client := api.NewDevicesGrpc(ctx, logger, client)
			responseSet, err := client.SetValue(ctx, &device2.SetValueRequest{
				Id:          setpoint.ID.Hex(),
				FeatureUuid: setpoint.UUID,
				FeatureName: setpoint.Name,
				Mac:         setpoint.Mac,
				Value:       status.Value,
				ApiToken:    setpoint.APIToken,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseSet.GetStatus()).To(Equal("200"))
			Expect(responseSet.GetMessage()).To(Equal("Updated"))

			controller, err := testutils.FindOneById[models.Device](ctx, setpointCollection, setpoint.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(controller.ID).To(Equal(setpoint.ID))
			Expect(controller.UUID).To(Equal(setpoint.UUID))
			Expect(controller.Mac).To(Equal(setpoint.Mac))
			Expect(controller.Name).To(Equal(setpoint.Name))
			Expect(controller.Manufacturer).To(Equal(setpoint.Manufacturer))
			Expect(controller.Model).To(Equal(setpoint.Model))
			Expect(controller.ProfileOwnerID).To(Equal(setpoint.ProfileOwnerID))
			Expect(controller.APIToken).To(Equal(setpoint.APIToken))
			Expect(controller.Status.Value).To(Equal(status.Value))

			responseGet, err := client.GetValue(ctx, &device2.GetValueRequest{
				Id:          setpoint.ID.Hex(),
				FeatureUuid: setpoint.UUID,
				FeatureName: setpoint.Name,
				Mac:         setpoint.Mac,
				ApiToken:    setpoint.APIToken,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseGet.GetValue()).To(Equal(status.Value))

			controller, err = testutils.FindOneById[models.Device](ctx, setpointCollection, setpoint.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(controller.Status.Value).To(Equal(status.Value))
		})

		It("should setValues of an existing airconditioner and get those values via getValues", func() {
			err := testutils.InsertOne(ctx, toleranceCollection, tolerance)
			Expect(err).ShouldNot(HaveOccurred())

			status := models.Status{
				Value: 20,
			}

			client := api.NewDevicesGrpc(ctx, logger, client)
			responseSet, err := client.SetValue(ctx, &device2.SetValueRequest{
				Id:          tolerance.ID.Hex(),
				FeatureUuid: tolerance.UUID,
				FeatureName: tolerance.Name,
				Mac:         tolerance.Mac,
				Value:       status.Value,
				ApiToken:    tolerance.APIToken,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseSet.GetStatus()).To(Equal("200"))
			Expect(responseSet.GetMessage()).To(Equal("Updated"))

			controller, err := testutils.FindOneById[models.Device](ctx, toleranceCollection, tolerance.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(controller.ID).To(Equal(tolerance.ID))
			Expect(controller.UUID).To(Equal(tolerance.UUID))
			Expect(controller.Mac).To(Equal(tolerance.Mac))
			Expect(controller.Name).To(Equal(tolerance.Name))
			Expect(controller.Manufacturer).To(Equal(tolerance.Manufacturer))
			Expect(controller.Model).To(Equal(tolerance.Model))
			Expect(controller.ProfileOwnerID).To(Equal(tolerance.ProfileOwnerID))
			Expect(controller.APIToken).To(Equal(tolerance.APIToken))
			Expect(controller.Status.Value).To(Equal(status.Value))

			responseGet, err := client.GetValue(ctx, &device2.GetValueRequest{
				Id:          tolerance.ID.Hex(),
				FeatureUuid: tolerance.UUID,
				FeatureName: tolerance.Name,
				Mac:         tolerance.Mac,
				ApiToken:    tolerance.APIToken,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseGet.GetValue()).To(Equal(status.Value))

			controller, err = testutils.FindOneById[models.Device](ctx, toleranceCollection, tolerance.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(controller.Status.Value).To(Equal(status.Value))
		})

		When("getValues", func() {
			It("should return an error, because controller doesn't exist on db", func() {
				missingMacDevice := "99:99:99:99:99:99"
				client := api.NewDevicesGrpc(ctx, logger, client)
				_, err := client.GetValue(ctx, &device2.GetValueRequest{
					Id:          airConditioner.ID.Hex(),
					FeatureUuid: airConditioner.UUID,
					FeatureName: airConditioner.Name,
					Mac:         missingMacDevice,
					ApiToken:    airConditioner.APIToken,
				})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal("cannot find controller with mac " + missingMacDevice))
			})
		})

		When("setValues", func() {
			It("should return an error, because controller doesn't exist on db", func() {
				status := models.Status{
					Value: 20,
				}
				missingMacDevice := "99:99:99:99:99:99"
				client := api.NewDevicesGrpc(ctx, logger, client)
				_, err := client.SetValue(ctx, &device2.SetValueRequest{
					Id:          airConditioner.ID.Hex(),
					FeatureUuid: airConditioner.UUID,
					FeatureName: airConditioner.Name,
					Mac:         missingMacDevice,
					Value:       status.Value,
					ApiToken:    airConditioner.APIToken,
				})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("cannot find controller with mac " + missingMacDevice))
			})
		})
	})
})
