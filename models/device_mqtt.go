package models

// MqttFeatureValue struct
type MqttFeatureValue struct {
	// profile info
	APIToken string `json:"apiToken"`
	// device info
	DeviceUUID string `json:"deviceUuid"`
	Mac        string `json:"mac"`
	Model      string `json:"model"`
	// feature info
	FeatureUUID string  `json:"featureUuid"`
	FeatureName string  `json:"featureName"`
	Value       float32 `json:"value"`
}
