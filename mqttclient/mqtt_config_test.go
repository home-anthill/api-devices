package mqttclient

import (
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestInitMqttReturnsTLSConfigError(t *testing.T) {
	t.Setenv("MQTT_TLS", "true")
	t.Setenv("MQTT_CA_FILE", "/path/does/not/exist")

	err := InitMqtt(zap.NewNop().Sugar())
	if err == nil {
		t.Fatal("expected TLS config error")
	}
	if !strings.Contains(err.Error(), "TLS configuration failed") {
		t.Fatalf("error = %q, want TLS configuration failed", err)
	}
}

func TestInitMqttConfiguresClientWithoutTLS(t *testing.T) {
	t.Setenv("MQTT_URL", "tcp://localhost")
	t.Setenv("MQTT_PORT", "1883")
	t.Setenv("MQTT_CLIENT_ID", "api-devices-test")
	t.Setenv("MQTT_AUTH", "true")
	t.Setenv("MQTT_USER", "user")
	t.Setenv("MQTT_PASSWORD", "password")
	t.Setenv("MQTT_TLS", "false")

	if err := InitMqtt(zap.NewNop().Sugar()); err != nil {
		t.Fatalf("init mqtt: %v", err)
	}
	if mqttClient == nil {
		t.Fatal("expected mqtt client")
	}
}

func TestNewTLSConfigRejectsMissingCAFile(t *testing.T) {
	t.Setenv("MQTT_CA_FILE", "/path/does/not/exist")

	_, err := newTLSConfig()
	if err == nil {
		t.Fatal("expected missing CA file error")
	}
	if !strings.Contains(err.Error(), "cannot read CA file") {
		t.Fatalf("error = %q, want cannot read CA file", err)
	}
}

func TestNewTLSConfigRejectsInvalidCAFile(t *testing.T) {
	caFile, err := os.CreateTemp(t.TempDir(), "ca.pem")
	if err != nil {
		t.Fatalf("create temp CA file: %v", err)
	}
	if _, writeErr := caFile.WriteString("not a certificate"); writeErr != nil {
		t.Fatalf("write temp CA file: %v", writeErr)
	}
	if closeErr := caFile.Close(); closeErr != nil {
		t.Fatalf("close temp CA file: %v", closeErr)
	}
	t.Setenv("MQTT_CA_FILE", caFile.Name())

	_, err = newTLSConfig()
	if err == nil {
		t.Fatal("expected invalid CA file error")
	}
	if !strings.Contains(err.Error(), "failed to append CA certificates") {
		t.Fatalf("error = %q, want failed to append CA certificates", err)
	}
}
