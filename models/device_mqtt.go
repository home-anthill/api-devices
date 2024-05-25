package models

// Values struct
type Values struct {
	UUID        string `json:"uuid"`
	APIToken    string `json:"apiToken"`
	On          bool   `json:"on"`
	Temperature int    `json:"temperature"`
	Mode        int    `json:"mode"`
	FanSpeed    int    `json:"fanSpeed"`
}
