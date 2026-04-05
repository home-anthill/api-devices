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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegisterGrpc implements the gRPC Registration service.
type RegisterGrpc struct {
	register.UnimplementedRegistrationServer
	controllersCollection *mongo.Collection
	logger                *zap.SugaredLogger
}

// NewRegisterGrpc creates a new RegisterGrpc handler.
func NewRegisterGrpc(logger *zap.SugaredLogger, client *mongo.Client) *RegisterGrpc {
	return &RegisterGrpc{
		controllersCollection: db.GetCollections(client).Controllers,
		logger:                logger,
	}
}

// Register upserts a device controller document in the database.
func (r *RegisterGrpc) Register(ctx context.Context, in *register.RegisterRequest) (*register.RegisterReply, error) {
	r.logger.Infof("gRPC - Register - Called for deviceUuid: %s, mac: %s", in.DeviceUuid, in.Mac)

	if in.Feature == nil {
		r.logger.Error("gRPC - Register - missing feature field")
		return nil, status.Errorf(codes.InvalidArgument, "feature is required")
	}

	profileOwnerID, err := bson.ObjectIDFromHex(in.ProfileOwnerId)
	if err != nil {
		r.logger.Errorf("gRPC - Register - Cannot update db: profileOwnerId is not a valid ObjectID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "profileOwnerId is not a valid ObjectID: %v", err)
	}

	now := time.Now()

	// query to upsert the registered controller
	setQuery := bson.M{
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
			"status.createdAt":  now,
			"status.modifiedAt": now,
			// dates
			"createdAt":  now,
			"modifiedAt": now,
		},
	}

	_, err = r.controllersCollection.UpdateOne(ctx, bson.M{
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
		r.logger.Errorf("gRPC - Register - Cannot update db with the registered device: %v", err)
		return nil, status.Errorf(codes.Internal, "cannot register controller: %v", err)
	}

	return &register.RegisterReply{Status: "200", Message: "Inserted"}, nil
}
