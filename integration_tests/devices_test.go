package integration_tests

import (
	"api-devices/api"
	devicepb "api-devices/api/device"
	"api-devices/db"
	"api-devices/initialization"
	"api-devices/models"
	mqttclient "api-devices/mqttclient"
	"api-devices/testutils"
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ = Describe("Devices", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var client *mongo.Client
	var controllersCollection *mongo.Collection
	var server *grpc.Server

	// profile info
	apiToken := "473a4861-632b-4915-b01e-cf1d41896601"
	profileOwnerId := bson.NewObjectID()
	// device info
	airConditionerUUID := uuid.NewString()
	airConditionerMac := "11:22:33:44:55:66"
	airConditionerModel := "ac-beko"
	airConditionerManufacturer := "ks89"
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
		ID:           bson.NewObjectID(), // controller _id
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
		ID:           bson.NewObjectID(), // controller _id
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
		ID:           bson.NewObjectID(), // controller _id
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
		ID:           bson.NewObjectID(), // controller _id
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
		ID:           bson.NewObjectID(), // controller _id
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
		ID:           bson.NewObjectID(), // controller _id
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

	checkFeature := func(client *api.DevicesGrpc, status models.Status, deviceUUID, mac, model, manufacturer, expectedFeatureUUID, expectedFeatureName string, featureId bson.ObjectID) {
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
		Expect(controller.FeatureUUID).To(Equal(expectedFeatureUUID))
		Expect(controller.FeatureName).To(Equal(expectedFeatureName))
		Expect(controller.Status.Value).To(Equal(status.Value))

		responseGet, err := client.GetValue(ctx, &devicepb.GetValueRequest{
			// profile info
			ApiToken: apiToken,
			// device info
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
		var listener net.Listener
		logger, server, listener, client = initialization.Start()
		go server.Serve(listener) //nolint:errcheck
		ctx = context.Background()

		controllersCollection = db.GetCollections(client).Controllers

		// create and start a mocked MQTT client
		mqttclient.SetMqttClient(testutils.NewMockClient())
		if token := mqttclient.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	})

	AfterEach(func() {
		server.Stop()

		testutils.DropCollection(ctx, controllersCollection)

		logger.Sync()
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

			client := api.NewDevicesGrpc(logger, client)
			responseSet, err := client.SetValues(ctx, &devicepb.SetValuesRequest{
				// profile info
				ApiToken: apiToken,
				// device info
				DeviceUuid: airConditionerUUID,
				Mac:        airConditionerMac,
				// feature info
				FeatureValues: []*devicepb.SetValueRequest{
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
			checkFeature(client, onStatus, airConditionerUUID, airConditionerMac, airConditionerModel, airConditionerManufacturer, onAirConditioner.FeatureUUID, onAirConditioner.FeatureName, onAirConditioner.ID)
			checkFeature(client, setpointStatus, airConditionerUUID, airConditionerMac, airConditionerModel, airConditionerManufacturer, setpointAirConditioner.FeatureUUID, setpointAirConditioner.FeatureName, setpointAirConditioner.ID)
			checkFeature(client, modeStatus, airConditionerUUID, airConditionerMac, airConditionerModel, airConditionerManufacturer, modeAirConditioner.FeatureUUID, modeAirConditioner.FeatureName, modeAirConditioner.ID)
			checkFeature(client, fanSpeedStatus, airConditionerUUID, airConditionerMac, airConditionerModel, airConditionerManufacturer, fanSpeedAirConditioner.FeatureUUID, fanSpeedAirConditioner.FeatureName, fanSpeedAirConditioner.ID)
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

			client := api.NewDevicesGrpc(logger, client)
			responseSet, err := client.SetValues(ctx, &devicepb.SetValuesRequest{
				// profile info
				ApiToken: apiToken,
				// device info
				DeviceUuid: thermostatUUID,
				Mac:        thermostatMac,
				// feature info
				FeatureValues: []*devicepb.SetValueRequest{
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

			// check thermostat features
			checkFeature(client, setpointStatus, thermostatUUID, thermostatMac, thermostatModel, thermostatManufacturer, setpointThermostat.FeatureUUID, setpointThermostat.FeatureName, setpointThermostat.ID)
			checkFeature(client, toleranceStatus, thermostatUUID, thermostatMac, thermostatModel, thermostatManufacturer, toleranceThermostat.FeatureUUID, toleranceThermostat.FeatureName, toleranceThermostat.ID)
		})

		When("getValues", func() {
			It("should return an error, because controller doesn't exist on db", func() {
				missingMacDevice := "99:99:99:99:99:99"
				client := api.NewDevicesGrpc(logger, client)
				_, err := client.GetValue(ctx, &devicepb.GetValueRequest{
					// profile info
					ApiToken: apiToken,
					// device info
					DeviceUuid: airConditionerUUID,
					Mac:        missingMacDevice,
					// feature info
					FeatureUuid: onFeatureUUID,
					FeatureName: "on",
				})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("cannot find controller"))
			})
		})

		When("setValues", func() {
			It("should return an error, because controller doesn't exist on db", func() {
				status := models.Status{
					Value: 20,
				}
				missingMacDevice := "99:99:99:99:99:99"
				client := api.NewDevicesGrpc(logger, client)
				_, err := client.SetValues(ctx, &devicepb.SetValuesRequest{
					// profile info
					ApiToken: apiToken,
					// device info
					DeviceUuid: airConditionerUUID,
					Mac:        missingMacDevice,
					// feature info
					FeatureValues: []*devicepb.SetValueRequest{
						{
							FeatureUuid: onFeatureUUID,
							FeatureName: "on",
							Value:       status.Value,
						},
					},
				})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("cannot find controller"))
			})
		})
	})
})
