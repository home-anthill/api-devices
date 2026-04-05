package db

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/zap"
)

// Collections holds references to all MongoDB collections used by the service.
type Collections struct {
	Controllers *mongo.Collection
}

// InitDb connects to MongoDB and returns the client.
func InitDb(logger *zap.SugaredLogger) *mongo.Client {
	mongoDBURL := os.Getenv("MONGODB_URL")
	logger.Info("InitDb - connecting to MongoDB...")

	// connect to DB
	client, err := mongo.Connect(options.Client().ApplyURI(mongoDBURL))
	if err != nil {
		logger.Fatalf("Cannot connect to MongoDB: %s", err)
	}
	if os.Getenv("ENV") != "prod" {
		pingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err = client.Ping(pingCtx, readpref.Primary()); err != nil {
			logger.Fatalf("Cannot ping MongoDB: %s", err)
		}
	}
	logger.Info("Connected to MongoDB")

	return client
}

// GetCollections returns the MongoDB collections for the appropriate database.
func GetCollections(client *mongo.Client) *Collections {
	return &Collections{
		Controllers: client.Database(getDbName()).Collection("controllers"),
	}
}

// getDbName returns the database name based on the environment.
func getDbName() string {
	if os.Getenv("ENV") == "testing" {
		return "controllers-test"
	}
	return "controllers"
}
