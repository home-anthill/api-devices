package db

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestGetDbNameUsesControllersByDefault(t *testing.T) {
	t.Setenv("ENV", "")

	if got := getDbName(); got != "controllers" {
		t.Fatalf("db name = %q, want %q", got, "controllers")
	}
}

func TestGetDbNameUsesTestDatabase(t *testing.T) {
	t.Setenv("ENV", "testing")

	if got := getDbName(); got != "controllers-test" {
		t.Fatalf("db name = %q, want %q", got, "controllers-test")
	}
}

func TestGetCollectionsReturnsControllersCollection(t *testing.T) {
	t.Setenv("ENV", "testing")
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("create mongo client: %v", err)
	}

	collections := GetCollections(client)

	if collections.Controllers.Name() != "controllers" {
		t.Fatalf("collection name = %q, want %q", collections.Controllers.Name(), "controllers")
	}
	if collections.Controllers.Database().Name() != "controllers-test" {
		t.Fatalf("database name = %q, want %q", collections.Controllers.Database().Name(), "controllers-test")
	}
}
