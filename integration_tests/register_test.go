package integration_tests

import (
	"api-devices/api"
	"api-devices/api/register"
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

var _ = Describe("Register", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var client *mongo.Client
	var controllersCollection *mongo.Collection
	var server *grpc.Server
	var listener net.Listener

	apiToken := "473a4861-632b-4915-b01e-cf1d41896601"
	deviceUUID := uuid.NewString()
	onFeatureUUID := uuid.NewString()
	setpointFeatureUUID := uuid.NewString()
	modeFeatureUUID := uuid.NewString()
	fanSpeedFeatureUUID := uuid.NewString()
	toleranceFeatureUUID := uuid.NewString()

	var onFeature = register.RegisterFeature{
		FeatureUuid: onFeatureUUID,
		FeatureName: "on",
	}
	var setpointFeature = register.RegisterFeature{
		FeatureUuid: setpointFeatureUUID,
		FeatureName: "setpoint",
	}
	var modeFeature = register.RegisterFeature{
		FeatureUuid: modeFeatureUUID,
		FeatureName: "mode",
	}
	var fanSpeedFeature = register.RegisterFeature{
		FeatureUuid: fanSpeedFeatureUUID,
		FeatureName: "fanSpeed",
	}
	var toleranceFeature = register.RegisterFeature{
		FeatureUuid: toleranceFeatureUUID,
		FeatureName: "tolerance",
	}

	var on = models.Controller{
		// profile info
		ProfileOwnerID: bson.NewObjectID(),
		APIToken:       apiToken,
		// device info
		ID:           bson.NewObjectID(),
		DeviceUUID:   deviceUUID,
		Mac:          "11:22:33:44:55:66",
		Model:        "ac-beko",
		Manufacturer: "ks89",
		// feature info
		FeatureUUID: onFeatureUUID,
		FeatureName: "on",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}
	var setpoint = models.Controller{
		// profile info
		ProfileOwnerID: bson.NewObjectID(),
		APIToken:       apiToken,
		// device info
		ID:           bson.NewObjectID(),
		DeviceUUID:   deviceUUID,
		Mac:          "11:22:33:44:55:66",
		Model:        "thermostat",
		Manufacturer: "ks89",
		// feature info
		FeatureUUID: setpointFeatureUUID,
		FeatureName: "setpoint",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}
	var mode = models.Controller{
		// profile info
		ProfileOwnerID: bson.NewObjectID(),
		APIToken:       apiToken,
		// device info
		ID:           bson.NewObjectID(),
		DeviceUUID:   deviceUUID,
		Mac:          "11:22:33:44:55:66",
		Model:        "ac-beko",
		Manufacturer: "ks89",
		// feature info
		FeatureUUID: modeFeatureUUID,
		FeatureName: "mode",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}
	var fanSpeed = models.Controller{
		// profile info
		ProfileOwnerID: bson.NewObjectID(),
		APIToken:       apiToken,
		// device info
		ID:           bson.NewObjectID(),
		DeviceUUID:   deviceUUID,
		Mac:          "11:22:33:44:55:66",
		Model:        "ac-beko",
		Manufacturer: "ks89",
		// feature info
		FeatureUUID: fanSpeedFeatureUUID,
		FeatureName: "fanSpeed",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}
	var tolerance = models.Controller{
		// profile info
		ProfileOwnerID: bson.NewObjectID(),
		APIToken:       apiToken,
		// device info
		ID:           bson.NewObjectID(),
		DeviceUUID:   deviceUUID,
		Mac:          "11:22:33:44:55:66",
		Model:        "thermostat",
		Manufacturer: "ks89",
		// feature info
		FeatureUUID: toleranceFeatureUUID,
		FeatureName: "tolerance",
		Status:      models.Status{},
		// dates
		CreatedAt:  time.Time{},
		ModifiedAt: time.Time{},
	}

	BeforeEach(func() {
		logger, server, listener, ctx, client = initialization.Start()
		defer logger.Sync()

		controllersCollection = db.GetCollections(client).Controllers

		// create and start a mocked MQTT client
		mqttclient.SetMqttClient(testutils.NewMockClient())
		if token := mqttclient.Connect(); token.Wait() && token.Error() != nil {
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

	Context("calling register grpc api", func() {
		It("should register on controller and return success", func() {
			client := api.NewRegisterGrpc(ctx, logger, client)
			response, err := client.Register(ctx, &register.RegisterRequest{
				// profile info
				ApiToken:       on.APIToken,
				ProfileOwnerId: on.ProfileOwnerID.Hex(),
				// device info
				DeviceUuid:   on.DeviceUUID,
				Mac:          on.Mac,
				Manufacturer: on.Manufacturer,
				Model:        on.Model,
				// feature info
				Feature: &onFeature,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.GetStatus()).To(Equal("200"))
			Expect(response.GetMessage()).To(Equal("Inserted"))

			controllerDb, err := testutils.FindOneByKeyValue[models.Controller](ctx, controllersCollection, "featureUuid", onFeatureUUID)
			Expect(err).ShouldNot(HaveOccurred())
			// check profile info
			Expect(controllerDb.APIToken).To(Equal(on.APIToken))
			Expect(controllerDb.ProfileOwnerID).To(Equal(on.ProfileOwnerID))
			// check device info
			Expect(controllerDb.DeviceUUID).To(Equal(on.DeviceUUID))
			Expect(controllerDb.Mac).To(Equal(on.Mac))
			Expect(controllerDb.Model).To(Equal(on.Model))
			Expect(controllerDb.Manufacturer).To(Equal(on.Manufacturer))
			// check feature info
			Expect(controllerDb.FeatureUUID).To(Equal(on.FeatureUUID))
			Expect(controllerDb.FeatureName).To(Equal(on.FeatureName))
		})

		It("should register mode controller and return success", func() {
			client := api.NewRegisterGrpc(ctx, logger, client)
			response, err := client.Register(ctx, &register.RegisterRequest{
				// profile info
				ApiToken:       mode.APIToken,
				ProfileOwnerId: mode.ProfileOwnerID.Hex(),
				// device info
				DeviceUuid:   mode.DeviceUUID,
				Mac:          mode.Mac,
				Manufacturer: mode.Manufacturer,
				Model:        mode.Model,
				// feature info
				Feature: &modeFeature,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.GetStatus()).To(Equal("200"))
			Expect(response.GetMessage()).To(Equal("Inserted"))

			controllerDb, err := testutils.FindOneByKeyValue[models.Controller](ctx, controllersCollection, "featureUuid", modeFeatureUUID)
			Expect(err).ShouldNot(HaveOccurred())
			// check profile info
			Expect(controllerDb.APIToken).To(Equal(mode.APIToken))
			Expect(controllerDb.ProfileOwnerID).To(Equal(mode.ProfileOwnerID))
			// check device info
			Expect(controllerDb.DeviceUUID).To(Equal(mode.DeviceUUID))
			Expect(controllerDb.Mac).To(Equal(mode.Mac))
			Expect(controllerDb.Model).To(Equal(mode.Model))
			Expect(controllerDb.Manufacturer).To(Equal(mode.Manufacturer))
			// check feature info
			Expect(controllerDb.FeatureUUID).To(Equal(mode.FeatureUUID))
			Expect(controllerDb.FeatureName).To(Equal(mode.FeatureName))
		})

		It("should register fanSpeed controller and return success", func() {
			client := api.NewRegisterGrpc(ctx, logger, client)
			response, err := client.Register(ctx, &register.RegisterRequest{
				// profile info
				ApiToken:       fanSpeed.APIToken,
				ProfileOwnerId: fanSpeed.ProfileOwnerID.Hex(),
				// device info
				DeviceUuid:   fanSpeed.DeviceUUID,
				Mac:          fanSpeed.Mac,
				Manufacturer: fanSpeed.Manufacturer,
				Model:        fanSpeed.Model,
				// feature info
				Feature: &fanSpeedFeature,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.GetStatus()).To(Equal("200"))
			Expect(response.GetMessage()).To(Equal("Inserted"))

			controllerDb, err := testutils.FindOneByKeyValue[models.Controller](ctx, controllersCollection, "featureUuid", fanSpeedFeatureUUID)
			Expect(err).ShouldNot(HaveOccurred())
			// check profile info
			Expect(controllerDb.APIToken).To(Equal(fanSpeed.APIToken))
			Expect(controllerDb.ProfileOwnerID).To(Equal(fanSpeed.ProfileOwnerID))
			// check device info
			Expect(controllerDb.DeviceUUID).To(Equal(fanSpeed.DeviceUUID))
			Expect(controllerDb.Mac).To(Equal(fanSpeed.Mac))
			Expect(controllerDb.Model).To(Equal(fanSpeed.Model))
			Expect(controllerDb.Manufacturer).To(Equal(fanSpeed.Manufacturer))
			// check feature info
			Expect(controllerDb.FeatureUUID).To(Equal(fanSpeed.FeatureUUID))
			Expect(controllerDb.FeatureName).To(Equal(fanSpeed.FeatureName))
		})

		It("should register setpoint controller and return success", func() {
			client := api.NewRegisterGrpc(ctx, logger, client)
			response, err := client.Register(ctx, &register.RegisterRequest{
				// profile info
				ApiToken:       setpoint.APIToken,
				ProfileOwnerId: setpoint.ProfileOwnerID.Hex(),
				// device info
				DeviceUuid:   setpoint.DeviceUUID,
				Mac:          setpoint.Mac,
				Manufacturer: setpoint.Manufacturer,
				Model:        setpoint.Model,
				// feature info
				Feature: &setpointFeature,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.GetStatus()).To(Equal("200"))
			Expect(response.GetMessage()).To(Equal("Inserted"))

			controllerDb, err := testutils.FindOneByKeyValue[models.Controller](ctx, controllersCollection, "featureUuid", setpointFeatureUUID)
			Expect(err).ShouldNot(HaveOccurred())
			// check profile info
			Expect(controllerDb.APIToken).To(Equal(setpoint.APIToken))
			Expect(controllerDb.ProfileOwnerID).To(Equal(setpoint.ProfileOwnerID))
			// check device info
			Expect(controllerDb.DeviceUUID).To(Equal(setpoint.DeviceUUID))
			Expect(controllerDb.Mac).To(Equal(setpoint.Mac))
			Expect(controllerDb.Model).To(Equal(setpoint.Model))
			Expect(controllerDb.Manufacturer).To(Equal(setpoint.Manufacturer))
			// check feature info
			Expect(controllerDb.FeatureUUID).To(Equal(setpoint.FeatureUUID))
			Expect(controllerDb.FeatureName).To(Equal(setpoint.FeatureName))
		})

		It("should register tolerance controller and return success", func() {
			client := api.NewRegisterGrpc(ctx, logger, client)
			response, err := client.Register(ctx, &register.RegisterRequest{
				// profile info
				ApiToken:       tolerance.APIToken,
				ProfileOwnerId: tolerance.ProfileOwnerID.Hex(),
				// device info
				DeviceUuid:   tolerance.DeviceUUID,
				Mac:          tolerance.Mac,
				Manufacturer: tolerance.Manufacturer,
				Model:        tolerance.Model,
				// feature info
				Feature: &toleranceFeature,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.GetStatus()).To(Equal("200"))
			Expect(response.GetMessage()).To(Equal("Inserted"))

			controllerDb, err := testutils.FindOneByKeyValue[models.Controller](ctx, controllersCollection, "featureUuid", toleranceFeatureUUID)
			Expect(err).ShouldNot(HaveOccurred())
			// check profile info
			Expect(controllerDb.APIToken).To(Equal(tolerance.APIToken))
			Expect(controllerDb.ProfileOwnerID).To(Equal(tolerance.ProfileOwnerID))
			// check device info
			Expect(controllerDb.DeviceUUID).To(Equal(tolerance.DeviceUUID))
			Expect(controllerDb.Mac).To(Equal(tolerance.Mac))
			Expect(controllerDb.Model).To(Equal(tolerance.Model))
			Expect(controllerDb.Manufacturer).To(Equal(tolerance.Manufacturer))
			// check feature info
			Expect(controllerDb.FeatureUUID).To(Equal(tolerance.FeatureUUID))
			Expect(controllerDb.FeatureName).To(Equal(tolerance.FeatureName))
		})
	})

	It("should return error, because profileOwnerId is not a valid ObjectId", func() {
		client := api.NewRegisterGrpc(ctx, logger, client)
		response, err := client.Register(ctx, &register.RegisterRequest{
			// profile info
			ApiToken:       tolerance.APIToken,
			ProfileOwnerId: "bad_string_profile_owner_id",
			// device info
			DeviceUuid: tolerance.DeviceUUID,
			Mac:        tolerance.Mac,
			// feature info
			Feature: &toleranceFeature,
		})
		Expect(response).Should(BeNil())
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("the provided hex string is not a valid ObjectID"))
	})
})
