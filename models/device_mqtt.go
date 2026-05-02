package models

// MqttFeatureValue represents the MQTT payload for a device feature value update.
type MqttFeatureValue struct {
	// device info
	DeviceUUID string `json:"deviceUuid"`
	Mac        string `json:"mac"`
	Model      string `json:"model"`
	// feature info
	FeatureUUID string  `json:"featureUuid"`
	FeatureName string  `json:"featureName"`
	Timestamp   int64   `json:"timestamp"`
	Nonce       string  `json:"nonce"`
	Signature   string  `json:"signature"`
	Payload     Payload `json:"payload"`
}

// Payload contains the actual command value signed inside MqttFeatureValue.
type Payload struct {
	Value float32 `json:"value"`
}
