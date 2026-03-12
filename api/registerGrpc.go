package api

import (
	"api-devices/api/register"
	"api-devices/db"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

// RegisterGrpc struct
type RegisterGrpc struct {
	register.UnimplementedRegistrationServer
	client                *mongo.Client
	controllersCollection *mongo.Collection
	ctx                   context.Context
	logger                *zap.SugaredLogger
}

// NewRegisterGrpc function
func NewRegisterGrpc(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client) *RegisterGrpc {
	return &RegisterGrpc{
		client:                client,
		controllersCollection: db.GetCollections(client).Controllers,
		ctx:                   ctx,
		logger:                logger,
	}
}

// Register function
func (handler *RegisterGrpc) Register(ctx context.Context, in *register.RegisterRequest) (*register.RegisterReply, error) {
	handler.logger.Infof("gRPC - Register - Called with in: %#v", in)

	profileOwnerID, err := bson.ObjectIDFromHex(in.ProfileOwnerId)
	if err != nil {
		handler.logger.Error("gRPC - Register - Cannot update db because profileOwnerID = " + in.ProfileOwnerId + " is not a valid ObjectID")
		return nil, err
	}

	// query to upsert the registered controller
	var setQuery = bson.M{
		"$set": bson.M{
			// profile info
			"profileOwnerId": profileOwnerID,
			"apiToken":       in.ApiToken,
			// device info
			"deviceUuid":   in.DeviceUuid,
			"mac":          in.Mac,
			"model":        in.Model,
			"manufacturer": in.Manufacturer,
			// feature info
			"featureUuid":       in.Feature.FeatureUuid,
			"featureName":       in.Feature.FeatureName,
			"status.value":      -999,
			"status.createdAt":  time.Now(),
			"status.modifiedAt": time.Now(),
			// dates
			"createdAt":  time.Now(),
			"modifiedAt": time.Now(),
		},
	}

	_, err = handler.controllersCollection.UpdateOne(handler.ctx, bson.M{
		// profile info
		"profileOwnerId": profileOwnerID,
		"apiToken":       in.ApiToken,
		// device info
		"deviceUuid":   in.DeviceUuid,
		"mac":          in.Mac,
		"model":        in.Model,
		"manufacturer": in.Manufacturer,
		// feature info
		"featureUuid": in.Feature.FeatureUuid,
		"featureName": in.Feature.FeatureName,
	}, setQuery, options.UpdateOne().SetUpsert(true))

	if err != nil {
		handler.logger.Error("gRPC - Register - Cannot update db with the registered device with mac " + in.Mac)
		return nil, err
	}

	return &register.RegisterReply{Status: "200", Message: "Inserted"}, err
}
