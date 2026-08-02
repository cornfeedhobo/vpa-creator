package main

import (
	"testing"

	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func TestParseUpdateModeAcceptsVPAEnumValues(t *testing.T) {
	tests := map[string]vpav1.UpdateMode{
		"Off":               vpav1.UpdateModeOff,
		"Initial":           vpav1.UpdateModeInitial,
		"Recreate":          vpav1.UpdateModeRecreate,
		"InPlaceOrRecreate": vpav1.UpdateModeInPlaceOrRecreate,
		"InPlace":           vpav1.UpdateModeInPlace,
	}

	for value, expected := range tests {
		t.Run(value, func(t *testing.T) {
			actual, ok := parseUpdateMode(value)
			if !ok {
				t.Fatal("expected update mode to parse")
			}
			if actual != expected {
				t.Fatalf("expected %q, got %q", expected, actual)
			}
		})
	}
}

func TestParseUpdateModeRejectsUnknownValues(t *testing.T) {
	if actual, ok := parseUpdateMode("Unknown"); ok {
		t.Fatalf("expected unknown update mode to be rejected, got %q", actual)
	}
}

func TestParseUpdateModeRejectsDeprecatedAuto(t *testing.T) {
	if actual, ok := parseUpdateMode("Auto"); ok {
		t.Fatalf("expected deprecated Auto update mode to be rejected, got %q", actual)
	}
}
