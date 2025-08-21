package db

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

var client *mongo.Client

// Collections struct
type Collections struct {
	AirConditioners *mongo.Collection
}

// InitDb function
func InitDb(ctx context.Context, logger *zap.SugaredLogger) *mongo.Client {
	mongoDBUrl := os.Getenv("MONGODB_URL")
	logger.Info("InitDb - connecting to MongoDB URL = " + mongoDBUrl)

	// connect to DB
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoDBUrl))
	if err != nil {
		logger.Fatalf("Cannot connect to MongoDB: %s", err)
		panic("Cannot connect to MongoDB")
	}
	if os.Getenv("ENV") != "prod" {
		if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
			logger.Fatalf("Cannot ping MongoDB: %s", err)
			panic("Cannot ping MongoDB")
		}
	}
	logger.Info("Connected to MongoDB")

	return client
}

// GetCollections function
func GetCollections(client *mongo.Client) *Collections {
	return &Collections{
		AirConditioners: client.Database(getDbName()).Collection("airconditioners"),
	}
}

// getDbName function
func getDbName() string {
	if os.Getenv("ENV") == "testing" {
		return "api-devices-test"
	} else {
		return "api-devices"
	}
}
