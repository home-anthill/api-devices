package utils

import (
	"strings"
	"testing"
)

func TestNormalizeMAC(t *testing.T) {
	for _, value := range []string{"aa:bb:cc:dd:ee:ff", "aA:bB:cC:dD:eE:fF", "AA:BB:CC:DD:EE:FF"} {
		got, err := NormalizeMAC(value)
		if err != nil || got != "AA:BB:CC:DD:EE:FF" {
			t.Fatalf("NormalizeMAC(%q) = %q, %v", value, got, err)
		}
	}
	for _, value := range []string{"", "AA-BB-CC-DD-EE-FF", "aabb.ccdd.eeff", "AA:BB:CC:DD:EE:FF:00:11", "AA:BB:CC:DD:EE", "AA:BB:CC:DD:EE:GG", " AA:BB:CC:DD:EE:FF"} {
		if _, err := NormalizeMAC(value); err == nil {
			t.Fatalf("invalid MAC %q accepted", value)
		}
	}
}

func TestValidateDeviceText(t *testing.T) {
	for _, pair := range [][2]string{{"ab", "abc"}, {strings.Repeat("a", 50), strings.Repeat("b", 50)}, {"model ~", "Maker Inc."}} {
		if err := ValidateDeviceText(pair[0], pair[1]); err != nil {
			t.Fatal(err)
		}
	}
	for _, pair := range [][2]string{{"a", "abc"}, {"ab", "ab"}, {strings.Repeat("a", 51), "abc"}, {"ab", strings.Repeat("a", 51)}, {"Modèle", "abc"}, {"ab", "Maker\n"}, {"ab\x7f", "abc"}, {"ab", "Maker\x00"}} {
		if err := ValidateDeviceText(pair[0], pair[1]); err == nil {
			t.Fatalf("invalid identity accepted: %q", pair)
		}
	}
}
