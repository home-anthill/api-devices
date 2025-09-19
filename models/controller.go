package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Type string
type Type string

// Controller and Sensor types
const (
	Controller Type = "controller"
	Sensor     Type = "sensor"
)

// Status struct
type Status struct {
	Value      float32   `json:"value" bson:"value"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt" bson:"modifiedAt"`
}

// Device struct
type Device struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`
	// UUID is the feature UUID
	UUID           string             `json:"uuid" bson:"uuid"`
	Mac            string             `json:"mac" bson:"mac"`
	Name           string             `json:"name" bson:"name"`
	Manufacturer   string             `json:"manufacturer" bson:"manufacturer"`
	Model          string             `json:"model" bson:"model"`
	ProfileOwnerID primitive.ObjectID `json:"profileOwnerId" bson:"profileOwnerId"`
	APIToken       string             `json:"apiToken" bson:"apiToken"`
	Status         Status             `json:"status" bson:"status"`

	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt" bson:"modifiedAt"`
}
