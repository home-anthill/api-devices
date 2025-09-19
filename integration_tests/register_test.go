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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ = Describe("Register", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var client *mongo.Client
	var airConditionerCollection *mongo.Collection
	var setpointCollection *mongo.Collection
	var toleranceCollection *mongo.Collection
	var server *grpc.Server
	var listener net.Listener

	var feature = register.RegisterFeature{
		FeatureUuid: "246e3256-f0dd-4fcb-82c5-ee20c2267eeb",
		FeatureName: "ac-beko",
		Enable:      true,
		Order:       1,
		Unit:        "-",
	}

	var airconditioner = models.Device{
		ID:             primitive.ObjectID{},
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

		testutils.DropAllCollections(ctx, airConditionerCollection)
		testutils.DropAllCollections(ctx, setpointCollection)
		testutils.DropAllCollections(ctx, toleranceCollection)

	})

	Context("calling register grpc api", func() {
		It("should register aitconditioner controller and return success", func() {
			client := api.NewRegisterGrpc(ctx, logger, client)
			response, err := client.Register(ctx, &register.RegisterRequest{
				Id:             airconditioner.ID.Hex(),
				Uuid:           airconditioner.UUID,
				Mac:            airconditioner.Mac,
				Manufacturer:   airconditioner.Manufacturer,
				Model:          airconditioner.Model,
				ProfileOwnerId: airconditioner.ProfileOwnerID.Hex(),
				ApiToken:       airconditioner.APIToken,
				Feature:        &feature,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.GetStatus()).To(Equal("200"))
			Expect(response.GetMessage()).To(Equal("Inserted"))

			ac, err := testutils.FindOneByKeyValue[models.Device](ctx, airConditionerCollection, "mac", airconditioner.Mac)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ac.UUID).To(Equal(feature.FeatureUuid))
			Expect(ac.Mac).To(Equal(airconditioner.Mac))
			Expect(ac.Name).To(Equal(feature.FeatureName))
			Expect(ac.Manufacturer).To(Equal(airconditioner.Manufacturer))
			Expect(ac.Model).To(Equal(airconditioner.Model))
			Expect(ac.ProfileOwnerID).To(Equal(airconditioner.ProfileOwnerID))
			Expect(ac.APIToken).To(Equal(airconditioner.APIToken))
		})

		It("should register setpoint controller and return success", func() {
			client := api.NewRegisterGrpc(ctx, logger, client)
			response, err := client.Register(ctx, &register.RegisterRequest{
				Id:             setpoint.ID.Hex(),
				Uuid:           setpoint.UUID,
				Mac:            setpoint.Mac,
				Manufacturer:   setpoint.Manufacturer,
				Model:          setpoint.Model,
				ProfileOwnerId: setpoint.ProfileOwnerID.Hex(),
				ApiToken:       setpoint.APIToken,
				Feature:        &feature,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.GetStatus()).To(Equal("200"))
			Expect(response.GetMessage()).To(Equal("Inserted"))

			ac, err := testutils.FindOneByKeyValue[models.Device](ctx, airConditionerCollection, "mac", setpoint.Mac)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ac.UUID).To(Equal(feature.FeatureUuid))
			Expect(ac.Mac).To(Equal(setpoint.Mac))
			Expect(ac.Name).To(Equal(feature.FeatureName))
			Expect(ac.Manufacturer).To(Equal(setpoint.Manufacturer))
			Expect(ac.Model).To(Equal(setpoint.Model))
			Expect(ac.ProfileOwnerID).To(Equal(setpoint.ProfileOwnerID))
			Expect(ac.APIToken).To(Equal(setpoint.APIToken))
		})

		It("should register tolerance controller and return success", func() {
			client := api.NewRegisterGrpc(ctx, logger, client)
			response, err := client.Register(ctx, &register.RegisterRequest{
				Id:             tolerance.ID.Hex(),
				Uuid:           tolerance.UUID,
				Mac:            tolerance.Mac,
				Manufacturer:   tolerance.Manufacturer,
				Model:          tolerance.Model,
				ProfileOwnerId: tolerance.ProfileOwnerID.Hex(),
				ApiToken:       tolerance.APIToken,
				Feature:        &feature,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.GetStatus()).To(Equal("200"))
			Expect(response.GetMessage()).To(Equal("Inserted"))

			ac, err := testutils.FindOneByKeyValue[models.Device](ctx, airConditionerCollection, "mac", tolerance.Mac)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ac.UUID).To(Equal(feature.FeatureUuid))
			Expect(ac.Mac).To(Equal(tolerance.Mac))
			Expect(ac.Name).To(Equal(feature.FeatureName))
			Expect(ac.Manufacturer).To(Equal(tolerance.Manufacturer))
			Expect(ac.Model).To(Equal(tolerance.Model))
			Expect(ac.ProfileOwnerID).To(Equal(tolerance.ProfileOwnerID))
			Expect(ac.APIToken).To(Equal(tolerance.APIToken))
		})
	})

	It("should return error, because profileOwnerId is not a valid ObjectId", func() {
		client := api.NewRegisterGrpc(ctx, logger, client)
		response, err := client.Register(ctx, &register.RegisterRequest{
			Id:             airconditioner.ID.Hex(),
			Uuid:           airconditioner.UUID,
			Mac:            airconditioner.Mac,
			Manufacturer:   airconditioner.Manufacturer,
			Model:          airconditioner.Model,
			ProfileOwnerId: "bad_string_profile_owner_id",
			ApiToken:       airconditioner.APIToken,
			Feature:        &feature,
		})
		Expect(response).Should(BeNil())
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("the provided hex string is not a valid ObjectID"))
	})
})
