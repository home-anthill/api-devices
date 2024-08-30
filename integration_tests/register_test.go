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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"time"
)

var _ = Describe("Register", func() {
	var ctx context.Context
	var logger *zap.SugaredLogger
	var client *mongo.Client
	var collectionACs *mongo.Collection
	var server *grpc.Server
	var listener net.Listener

	var device = models.AirConditioner{
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

	BeforeEach(func() {
		logger, server, listener, ctx, client = initialization.Start()
		defer logger.Sync()

		collectionACs = db.GetCollections(client).AirConditioners

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

		testutils.DropAllCollections(ctx, collectionACs)
	})

	Context("calling register grpc api", func() {
		It("should return success", func() {
			client := api.NewRegisterGrpc(ctx, logger, client)
			response, err := client.Register(ctx, &register.RegisterRequest{
				Id:             device.ID.Hex(),
				Uuid:           device.UUID,
				Mac:            device.Mac,
				Name:           device.Name,
				Manufacturer:   device.Manufacturer,
				Model:          device.Model,
				ProfileOwnerId: device.ProfileOwnerID.Hex(),
				ApiToken:       device.APIToken,
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.GetStatus()).To(Equal("200"))
			Expect(response.GetMessage()).To(Equal("Inserted"))

			ac, err := testutils.FindOneByKeyValue[models.AirConditioner](ctx, collectionACs, "mac", device.Mac)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ac.UUID).To(Equal(device.UUID))
			Expect(ac.Mac).To(Equal(device.Mac))
			Expect(ac.Name).To(Equal(device.Name))
			Expect(ac.Manufacturer).To(Equal(device.Manufacturer))
			Expect(ac.Model).To(Equal(device.Model))
			Expect(ac.ProfileOwnerID).To(Equal(device.ProfileOwnerID))
			Expect(ac.APIToken).To(Equal(device.APIToken))
		})
	})

	It("should return error, because profileOwnerId is not a valid ObjectId", func() {
		client := api.NewRegisterGrpc(ctx, logger, client)
		response, err := client.Register(ctx, &register.RegisterRequest{
			Id:             device.ID.Hex(),
			Uuid:           device.UUID,
			Mac:            device.Mac,
			Name:           device.Name,
			Manufacturer:   device.Manufacturer,
			Model:          device.Model,
			ProfileOwnerId: "bad_string_profile_owner_id",
			ApiToken:       device.APIToken,
		})
		Expect(response).Should(BeNil())
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("the provided hex string is not a valid ObjectID"))
	})
})
