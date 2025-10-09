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

	"github.com/google/uuid"
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
	var controllersCollection *mongo.Collection
	var server *grpc.Server
	var listener net.Listener

	// profile info
	apiToken := "473a4861-632b-4915-b01e-cf1d41896601"
	profileOwnerId := primitive.NewObjectID()
	// device info
	airConditionerId := primitive.NewObjectID()
	airConditionerUUID := uuid.NewString()
	airConditionerMac := "11:22:33:44:55:66"
	airConditionerModel := "ac-beko"
	airConditionerManufacturer := "ks89"
	thermostatId := primitive.NewObjectID()
	thermostatUUID := uuid.NewString()
	thermostatMac := "AA:BB:CC:DD:EE:FF"
	thermostatModel := "thermostat"
	thermostatManufacturer := "ks89"
	// features info
	onFeatureUUID := uuid.NewString()
	setpointFeatureUUID := uuid.NewString()
	modeFeatureUUID := uuid.NewString()
	fanSpeedFeatureUUID := uuid.NewString()
	toleranceFeatureUUID := uuid.NewString()

	var onAirConditioner = models.Controller{
		// profile info
		ProfileOwnerID: profileOwnerId,
		APIToken:       apiToken,
		// device info
		ID:           primitive.NewObjectID(), // controller _id
		DeviceUUID:   airConditionerUUID,
		Mac:          airConditionerMac,
		Model:        airConditionerModel,
		Manufacturer: airConditionerManufacturer,
		// feature info
		FeatureUUID: onFeatureUUID,
		FeatureName: "on",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}
	var setpointAirConditioner = models.Controller{
		// profile info
		ProfileOwnerID: profileOwnerId,
		APIToken:       apiToken,
		// device info
		ID:           primitive.NewObjectID(), // controller _id
		DeviceUUID:   airConditionerUUID,
		Mac:          airConditionerMac,
		Model:        airConditionerModel,
		Manufacturer: airConditionerManufacturer,
		// feature info
		FeatureUUID: setpointFeatureUUID,
		FeatureName: "setpoint",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}
	var modeAirConditioner = models.Controller{
		// profile info
		ProfileOwnerID: profileOwnerId,
		APIToken:       apiToken,
		// device info
		ID:           primitive.NewObjectID(), // controller _id
		DeviceUUID:   airConditionerUUID,
		Mac:          airConditionerMac,
		Model:        airConditionerModel,
		Manufacturer: airConditionerManufacturer,
		// feature info
		FeatureUUID: modeFeatureUUID,
		FeatureName: "mode",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}
	var fanSpeedAirConditioner = models.Controller{
		// profile info
		ProfileOwnerID: profileOwnerId,
		APIToken:       apiToken,
		// device info
		ID:           primitive.NewObjectID(), // controller _id
		DeviceUUID:   airConditionerUUID,
		Mac:          airConditionerMac,
		Model:        airConditionerModel,
		Manufacturer: airConditionerManufacturer,
		// feature info
		FeatureUUID: fanSpeedFeatureUUID,
		FeatureName: "fanSpeed",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}

	var setpointThermostat = models.Controller{
		// profile info
		ProfileOwnerID: profileOwnerId,
		APIToken:       apiToken,
		// device info
		ID:           primitive.NewObjectID(), // controller _id
		DeviceUUID:   thermostatUUID,
		Mac:          thermostatMac,
		Model:        thermostatModel,
		Manufacturer: thermostatManufacturer,
		// feature info
		FeatureUUID: setpointFeatureUUID,
		FeatureName: "setpoint",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}
	var toleranceThermostat = models.Controller{
		// profile info
		ProfileOwnerID: profileOwnerId,
		APIToken:       apiToken,
		// device info
		ID:           primitive.NewObjectID(), // controller _id
		DeviceUUID:   thermostatUUID,
		Mac:          thermostatMac,
		Model:        thermostatModel,
		Manufacturer: thermostatManufacturer,
		// feature info
		FeatureUUID: toleranceFeatureUUID,
		FeatureName: "tolerance",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}

	checkFeature := func(client *api.DevicesGrpc, status models.Status, deviceId primitive.ObjectID, deviceUUID, mac, model, manufacturer string, featureId primitive.ObjectID) {
		controller, err := testutils.FindOneById[models.Controller](ctx, controllersCollection, featureId)
		Expect(err).ShouldNot(HaveOccurred())
		// check profile info
		Expect(controller.APIToken).To(Equal(apiToken))
		Expect(controller.ProfileOwnerID).To(Equal(profileOwnerId))
		// check device info
		Expect(controller.ID).To(Equal(featureId))
		Expect(controller.DeviceUUID).To(Equal(deviceUUID))
		Expect(controller.Mac).To(Equal(mac))
		Expect(controller.Model).To(Equal(model))
		Expect(controller.Manufacturer).To(Equal(manufacturer))
		// check feature info
		Expect(controller.FeatureUUID).To(Equal(controller.FeatureUUID))
		Expect(controller.FeatureName).To(Equal(controller.FeatureName))
		Expect(controller.Status.Value).To(Equal(status.Value))

		responseGet, err := client.GetValue(ctx, &device2.GetValueRequest{
			// profile info
			ApiToken: apiToken,
			// device info
			Id:         deviceId.Hex(),
			DeviceUuid: deviceUUID,
			Mac:        mac,
			// feature info
			FeatureUuid: controller.FeatureUUID,
			FeatureName: controller.FeatureName,
		})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(responseGet.GetValue()).To(Equal(status.Value))

		controller, err = testutils.FindOneById[models.Controller](ctx, controllersCollection, featureId)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(controller.Status.Value).To(Equal(status.Value))
	}

	BeforeEach(func() {
		logger, server, listener, ctx, client = initialization.Start()
		defer logger.Sync()

		controllersCollection = db.GetCollections(client).Controllers

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

		testutils.DropAllCollections(ctx, controllersCollection)
	})

	Context("calling devices grpc api", func() {
		It("should setValues of an existing air-conditioner and get those values via getValues", func() {
			err := testutils.InsertOne(ctx, controllersCollection, onAirConditioner)
			Expect(err).ShouldNot(HaveOccurred())
			err = testutils.InsertOne(ctx, controllersCollection, setpointAirConditioner)
			Expect(err).ShouldNot(HaveOccurred())
			err = testutils.InsertOne(ctx, controllersCollection, modeAirConditioner)
			Expect(err).ShouldNot(HaveOccurred())
			err = testutils.InsertOne(ctx, controllersCollection, fanSpeedAirConditioner)
			Expect(err).ShouldNot(HaveOccurred())

			onStatus := models.Status{
				Value: 1,
			}
			setpointStatus := models.Status{
				Value: 10.23,
			}
			modeStatus := models.Status{
				Value: 1,
			}
			fanSpeedStatus := models.Status{
				Value: 2,
			}

			client := api.NewDevicesGrpc(ctx, logger, client)
			responseSet, err := client.SetValues(ctx, &device2.SetValuesRequest{
				// profile info
				ApiToken: apiToken,
				// device info
				Id:         airConditionerId.Hex(),
				DeviceUuid: airConditionerUUID,
				Mac:        airConditionerMac,
				// feature info
				FeatureValues: []*device2.SetValueRequest{
					{
						FeatureUuid: onAirConditioner.FeatureUUID,
						FeatureName: onAirConditioner.FeatureName,
						Value:       onStatus.Value,
					}, {
						FeatureUuid: setpointAirConditioner.FeatureUUID,
						FeatureName: setpointAirConditioner.FeatureName,
						Value:       setpointStatus.Value,
					}, {
						FeatureUuid: modeAirConditioner.FeatureUUID,
						FeatureName: modeAirConditioner.FeatureName,
						Value:       modeStatus.Value,
					}, {
						FeatureUuid: fanSpeedAirConditioner.FeatureUUID,
						FeatureName: fanSpeedAirConditioner.FeatureName,
						Value:       fanSpeedStatus.Value,
					},
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseSet.GetStatus()).To(Equal("200"))
			Expect(responseSet.GetMessage()).To(Equal("Updated"))

			// check air conditioner features
			checkFeature(client, onStatus, airConditionerId, airConditionerUUID, airConditionerMac, airConditionerModel, airConditionerManufacturer, onAirConditioner.ID)
			checkFeature(client, setpointStatus, airConditionerId, airConditionerUUID, airConditionerMac, airConditionerModel, airConditionerManufacturer, setpointAirConditioner.ID)
			checkFeature(client, modeStatus, airConditionerId, airConditionerUUID, airConditionerMac, airConditionerModel, airConditionerManufacturer, modeAirConditioner.ID)
			checkFeature(client, fanSpeedStatus, airConditionerId, airConditionerUUID, airConditionerMac, airConditionerModel, airConditionerManufacturer, fanSpeedAirConditioner.ID)
		})

		It("should setValues of an existing thermostat and get those values via getValues", func() {
			err := testutils.InsertOne(ctx, controllersCollection, setpointThermostat)
			Expect(err).ShouldNot(HaveOccurred())
			err = testutils.InsertOne(ctx, controllersCollection, toleranceThermostat)
			Expect(err).ShouldNot(HaveOccurred())

			setpointStatus := models.Status{
				Value: 10.23,
			}
			toleranceStatus := models.Status{
				Value: 2.1,
			}

			client := api.NewDevicesGrpc(ctx, logger, client)
			responseSet, err := client.SetValues(ctx, &device2.SetValuesRequest{
				// profile info
				ApiToken: apiToken,
				// device info
				Id:         thermostatId.Hex(),
				DeviceUuid: thermostatUUID,
				Mac:        thermostatMac,
				// feature info
				FeatureValues: []*device2.SetValueRequest{
					{
						FeatureUuid: setpointThermostat.FeatureUUID,
						FeatureName: setpointThermostat.FeatureName,
						Value:       setpointStatus.Value,
					}, {
						FeatureUuid: toleranceThermostat.FeatureUUID,
						FeatureName: toleranceThermostat.FeatureName,
						Value:       toleranceStatus.Value,
					},
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(responseSet.GetStatus()).To(Equal("200"))
			Expect(responseSet.GetMessage()).To(Equal("Updated"))

			// check air conditioner features
			checkFeature(client, setpointStatus, thermostatId, thermostatUUID, thermostatMac, thermostatModel, thermostatManufacturer, setpointThermostat.ID)
			checkFeature(client, toleranceStatus, thermostatId, thermostatUUID, thermostatMac, thermostatModel, thermostatManufacturer, toleranceThermostat.ID)
		})

		When("getValues", func() {
			It("should return an error, because controller doesn't exist on db", func() {
				missingMacDevice := "99:99:99:99:99:99"
				client := api.NewDevicesGrpc(ctx, logger, client)
				_, err := client.GetValue(ctx, &device2.GetValueRequest{
					// profile info
					ApiToken: apiToken,
					// device info
					Id:         airConditionerId.Hex(),
					DeviceUuid: airConditionerUUID,
					Mac:        missingMacDevice,
					// feature info
					FeatureUuid: onFeatureUUID,
					FeatureName: "on",
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
				_, err := client.SetValues(ctx, &device2.SetValuesRequest{
					// profile info
					ApiToken: apiToken,
					// device info
					Id:         airConditionerId.Hex(),
					DeviceUuid: airConditionerUUID,
					Mac:        missingMacDevice,
					// feature info
					FeatureValues: []*device2.SetValueRequest{
						{
							FeatureUuid: onFeatureUUID,
							FeatureName: "on",
							Value:       status.Value,
						},
					},
				})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("cannot find controller with mac " + missingMacDevice))
			})
		})
	})
})
