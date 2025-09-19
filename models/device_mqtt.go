package models

// Values struct
type Values struct {
	UUID        string  `json:"uuid"`
	APIToken    string  `json:"apiToken"`
	FeatureName string  `json:"featureName"`
	Value       float64 `json:"value"`
}
