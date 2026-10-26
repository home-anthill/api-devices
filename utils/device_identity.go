package utils

import (
	"fmt"
	"net"
	"strings"
)

// NormalizeMAC validates six colon-separated hexadecimal octets and returns uppercase.
func NormalizeMAC(value string) (string, error) {
	mac, err := net.ParseMAC(value)
	if err != nil || len(mac) != 6 || len(value) != 17 || !strings.Contains(value, ":") {
		return "", fmt.Errorf("mac must be in XX:XX:XX:XX:XX:XX format")
	}
	return strings.ToUpper(value), nil
}

// ValidateDeviceText applies the same identity constraints as admission and register.
func ValidateDeviceText(model, manufacturer string) error {
	for _, field := range []struct {
		name, value string
		min         int
	}{
		{"model", model, 2},
		{"manufacturer", manufacturer, 3},
	} {
		if len(field.value) < field.min || len(field.value) > 50 {
			return fmt.Errorf("%s must be between %d and 50 characters", field.name, field.min)
		}
		for _, char := range field.value {
			if char < 0x20 || char > 0x7e {
				return fmt.Errorf("%s must contain only printable ASCII characters", field.name)
			}
		}
	}
	return nil
}
