package mqttclient

import (
	"testing"

	"api-devices/testutils"

	"github.com/google/uuid"
)

func TestSendValuesRejectsInvalidDeviceUUID(t *testing.T) {
	SetMqttClient(testutils.NewMockClient())

	token, err := SendValues("device/abc", []byte(`[]`))
	if err == nil {
		t.Fatal("expected invalid UUID error")
	}
	if token != nil {
		t.Fatal("expected nil token for invalid UUID")
	}
}

func TestSendValuesAcceptsValidDeviceUUID(t *testing.T) {
	SetMqttClient(testutils.NewMockClient())

	token, err := SendValues(uuid.NewString(), []byte(`[]`))
	if err != nil {
		t.Fatalf("expected valid UUID to publish: %v", err)
	}
	if token == nil {
		t.Fatal("expected publish token")
	}
}
