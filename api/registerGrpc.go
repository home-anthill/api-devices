package api

import (
	"api-devices/api/register"
	"api-devices/db"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

// RegisterGrpc struct
type RegisterGrpc struct {
	register.UnimplementedRegistrationServer
	client                   *mongo.Client
	airConditionerCollection *mongo.Collection
	ctx                      context.Context
	logger                   *zap.SugaredLogger
}

// NewRegisterGrpc function
func NewRegisterGrpc(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client) *RegisterGrpc {
	return &RegisterGrpc{
		client:                   client,
		airConditionerCollection: db.GetCollections(client).AirConditioners,
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
	// update ac
	upsert := true
	opts := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err = handler.airConditionerCollection.UpdateOne(handler.ctx, bson.M{
		"mac": in.Mac,
	}, bson.M{
		"$set": bson.M{
			"mac":            in.Mac,
			"uuid":           in.Uuid,
			"name":           in.Name,
			"manufacturer":   in.Manufacturer,
			"model":          in.Model,
			"profileOwnerId": profileOwnerID,
			"apiToken":       in.ApiToken,
			"createdAt":      time.Now(),
			"modifiedAt":     time.Now(),
		},
	}, &opts)

	if err != nil {
		handler.logger.Error("gRPC - Register - Cannot update db with the registered AC with mac " + in.Mac)
		return nil, err
	}

	return &register.RegisterReply{Status: "200", Message: "Inserted"}, err
}
