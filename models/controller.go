package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Type string
type Type string

// ControllerType and SensorType types
const (
	ControllerType Type = "controller"
	SensorType     Type = "sensor"
)

// Status struct
type Status struct {
	Value      float32   `json:"value" bson:"value"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt" bson:"modifiedAt"`
}

// Controller struct
type Controller struct {
	// profile info
	ProfileOwnerID primitive.ObjectID `json:"profileOwnerId" bson:"profileOwnerId"`
	APIToken       string             `json:"apiToken" bson:"apiToken"`
	// device info
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	DeviceUUID   string             `json:"deviceUuid" bson:"deviceUuid"`
	Mac          string             `json:"mac" bson:"mac"`
	Model        string             `json:"model" bson:"model"`
	Manufacturer string             `json:"manufacturer" bson:"manufacturer"`
	// feature info
	FeatureUUID string `json:"featureUuid" bson:"featureUuid"`
	FeatureName string `json:"featureName" bson:"featureName"`
	Status      Status `json:"status" bson:"status"`
	// dates
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt" bson:"modifiedAt"`
}
