package db

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/zap"
)

var client *mongo.Client

// Collections struct
type Collections struct {
	Controllers *mongo.Collection
}

// InitDb function
func InitDb(logger *zap.SugaredLogger) *mongo.Client {
	mongoDBUrl := os.Getenv("MONGODB_URL")
	logger.Info("InitDb - connecting to MongoDB URL = " + mongoDBUrl)

	// connect to DB
	var err error
	client, err = mongo.Connect(options.Client().ApplyURI(mongoDBUrl))
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
		Controllers: client.Database(getDbName()).Collection("controllers"),
	}
}

// getDbName function
func getDbName() string {
	if os.Getenv("ENV") == "testing" {
		return "controllers-test"
	} else {
		return "controllers"
	}
}
