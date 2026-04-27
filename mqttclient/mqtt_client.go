package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const qos byte = 0

var (
	mu         sync.RWMutex
	mqttClient mqtt.Client
)

// InitMqtt creates and configures the MQTT client.
func InitMqtt(log *zap.SugaredLogger) error {
	opts, err := getMqttConfig(log)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	mqttClient = mqtt.NewClient(opts)
	return nil
}

// SetMqttClient is a public function used in testing to set a MQTT Mock Client
// as mqttClient local private variable.
func SetMqttClient(client mqtt.Client) {
	mu.Lock()
	defer mu.Unlock()
	mqttClient = client
}

// Connect initiates a connection to the MQTT broker.
func Connect() mqtt.Token {
	mu.RLock()
	defer mu.RUnlock()
	return mqttClient.Connect()
}

// SendValues publishes device feature values to the MQTT topic for the given device.
func SendValues(deviceUUID string, messageJSON []byte) (mqtt.Token, error) {
	if _, err := uuid.Parse(deviceUUID); err != nil {
		return nil, fmt.Errorf("invalid device UUID: %w", err)
	}

	mu.RLock()
	defer mu.RUnlock()
	return mqttClient.Publish(fmt.Sprintf("devices/%s/values", deviceUUID), qos, false, messageJSON), nil
}

func getMqttConfig(log *zap.SugaredLogger) (*mqtt.ClientOptions, error) {
	mqttURL := fmt.Sprintf("%s:%s", os.Getenv("MQTT_URL"), os.Getenv("MQTT_PORT"))
	user := os.Getenv("MQTT_USER")
	password := os.Getenv("MQTT_PASSWORD")
	clientID := os.Getenv("MQTT_CLIENT_ID")

	opts := mqtt.NewClientOptions()
	if os.Getenv("MQTT_AUTH") == "true" {
		opts.SetUsername(user)
		opts.SetPassword(password)
	}
	opts.SetKeepAlive(5 * time.Second)
	opts.SetPingTimeout(2 * time.Second)
	opts.AddBroker(mqttURL)
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Warnf("UNKNOWN TOPIC - MessageID: %d, Topic: %s", msg.MessageID(), msg.Topic())
	})

	if os.Getenv("MQTT_TLS") == "true" {
		tlsConfig, err := newTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("getMqttConfig - TLS configuration failed: %w", err)
		}
		opts.SetTLSConfig(tlsConfig)
	}
	return opts, nil
}

func newTLSConfig() (*tls.Config, error) {
	// Import trusted certificates from CAfile.pem.
	certpool := x509.NewCertPool()
	pemCerts, err := os.ReadFile(os.Getenv("MQTT_CA_FILE"))
	if err != nil {
		return nil, fmt.Errorf("newTLSConfig - cannot read CA file: %w", err)
	}
	if !certpool.AppendCertsFromPEM(pemCerts) {
		return nil, fmt.Errorf("newTLSConfig - failed to append CA certificates")
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(os.Getenv("MQTT_CERT_FILE"), os.Getenv("MQTT_KEY_FILE"))
	if err != nil {
		return nil, fmt.Errorf("newTLSConfig - cannot load key pair: %w", err)
	}

	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("newTLSConfig - cannot parse certificate: %w", err)
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyway.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		// ATTENTION!!!
		// To use "InsecureSkipVerify: false" you need to connect to MQTT using the public domain
		InsecureSkipVerify: false,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}, nil
}
