package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Type represents the category of a controller (e.g. controller, sensor).
type Type string

// ControllerType and SensorType types
const (
	ControllerType Type = "controller"
	SensorType     Type = "sensor"
)

// Status represents the current state and timestamps of a device feature.
type Status struct {
	Value      float32   `json:"value" bson:"value"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt" bson:"modifiedAt"`
}

// Controller represents a device feature document stored in MongoDB.
type Controller struct {
	// profile info
	ProfileOwnerID bson.ObjectID `json:"profileOwnerId" bson:"profileOwnerId"`
	APIToken       string        `json:"apiToken" bson:"apiToken"`
	// device info
	ID           bson.ObjectID `json:"id" bson:"_id"`
	DeviceUUID   string        `json:"deviceUuid" bson:"deviceUuid"`
	Mac          string        `json:"mac" bson:"mac"`
	Model        string        `json:"model" bson:"model"`
	Manufacturer string        `json:"manufacturer" bson:"manufacturer"`
	// feature info
	FeatureUUID string `json:"featureUuid" bson:"featureUuid"`
	FeatureName string `json:"featureName" bson:"featureName"`
	Status      Status `json:"status" bson:"status"`
	// dates
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt" bson:"modifiedAt"`
}
