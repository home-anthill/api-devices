package models_test

import (
	"api-devices/models"
	"math"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestIsModeValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value float32
		want  bool
	}{
		{name: "off", value: -1.0, want: true},
		{name: "auto", value: 0.0, want: true},
		{name: "heat", value: 1.0, want: true},
		{name: "cool", value: 2.0, want: true},
		{name: "fractional", value: 1.5, want: false},
		{name: "below range", value: -2.0, want: false},
		{name: "above range", value: 3.0, want: false},
		{name: "NaN", value: float32(math.NaN()), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := models.IsModeValue(tt.value); got != tt.want {
				t.Fatalf("IsModeValue(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestControllerStatusValueDecodesBSONDouble(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	id := bson.NewObjectID()

	raw, err := bson.Marshal(bson.M{
		"_id":            id,
		"profileOwnerId": bson.NewObjectID(),
		"apiToken":       "token",
		"deviceUuid":     "08fe3a05-a3f7-4977-b943-f574f2ba43ca",
		"mac":            "1C:DB:D4:41:38:B4",
		"model":          "thermostat",
		"manufacturer":   "ks89",
		"featureUuid":    "69c1ff7a-0278-49e6-8c99-6b5617a68460",
		"featureName":    "setpoint",
		"status": bson.M{
			"value":      10.23,
			"createdAt":  now,
			"modifiedAt": now,
		},
		"createdAt":  now,
		"modifiedAt": now,
	})
	if err != nil {
		t.Fatalf("marshal controller: %v", err)
	}

	var controller models.Controller
	if err := bson.Unmarshal(raw, &controller); err != nil {
		t.Fatalf("unmarshal controller: %v", err)
	}

	if got, want := controller.Status.Value, float32(10.23); got != want {
		t.Fatalf("status.value = %v, want %v", got, want)
	}
}
