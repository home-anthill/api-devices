package api

import (
	"api-devices/api/register"
	"api-devices/db"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// RegisterGrpc struct
type RegisterGrpc struct {
	register.UnimplementedRegistrationServer
	client                   *mongo.Client
	airConditionerCollection *mongo.Collection
	setpointCollection       *mongo.Collection
	toleranceCollection      *mongo.Collection
	ctx                      context.Context
	logger                   *zap.SugaredLogger
}

// NewRegisterGrpc function
func NewRegisterGrpc(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client) *RegisterGrpc {
	return &RegisterGrpc{
		client:                   client,
		airConditionerCollection: db.GetCollections(client).AirConditioners,
		setpointCollection:       db.GetCollections(client).Setpoints,
		toleranceCollection:      db.GetCollections(client).Tolerances,
		ctx:                      ctx,
		logger:                   logger,
	}
}

// Register function
func (handler *RegisterGrpc) Register(ctx context.Context, in *register.RegisterRequest) (*register.RegisterReply, error) {
	handler.logger.Infof("gRPC - Register - Called with in: %#v", in)

	profileOwnerID, err := primitive.ObjectIDFromHex(in.ProfileOwnerId)
	if err != nil {
		handler.logger.Error("gRPC - Register - Cannot update db because profileOwnerID = " + in.ProfileOwnerId + " is not a valid ObjectID")
		return nil, err
	}
	// update controller
	upsert := true
	opts := options.UpdateOptions{
		Upsert: &upsert,
	}

	var setQuery = bson.M{
		"$set": bson.M{
			"mac":               in.Mac,
			"name":              in.Feature.FeatureName,
			"manufacturer":      in.Manufacturer,
			"model":             in.Model,
			"profileOwnerId":    profileOwnerID,
			"apiToken":          in.ApiToken,
			"createdAt":         time.Now(),
			"modifiedAt":        time.Now(),
			"status.createdAt":  time.Now(),
			"status.modifiedAt": time.Now(),
		},
	}

	var collection *mongo.Collection
	switch in.Feature.FeatureName {
	case "ac-lg", "ac-beko":
		collection = handler.airConditionerCollection
	case "setpoint":
		collection = handler.setpointCollection
	case "tolerance":
		collection = handler.toleranceCollection
	default:
		handler.logger.Error("gRPC - Register - Unknown controller feature '" + in.Feature.FeatureName + "' with mac " + in.Mac)
		return nil, fmt.Errorf("unknown controller feature %s with mac %s", in.Feature.FeatureName, in.Mac)
	}

	_, err = collection.UpdateOne(handler.ctx, bson.M{
		"uuid": in.Feature.FeatureUuid,
	}, setQuery, &opts)

	if err != nil {
		handler.logger.Error("gRPC - Register - Cannot update db with the registered device with mac " + in.Mac)
		return nil, err
	}

	return &register.RegisterReply{Status: "200", Message: "Inserted"}, err
}
