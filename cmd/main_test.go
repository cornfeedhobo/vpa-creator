package main

import (
	"testing"

	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func TestParseControlledValuesAcceptsVPAEnumValues(t *testing.T) {
	tests := map[string]vpav1.ContainerControlledValues{
		"RequestsAndLimits": vpav1.ContainerControlledValuesRequestsAndLimits,
		"RequestsOnly":      vpav1.ContainerControlledValuesRequestsOnly,
	}

	for value, expected := range tests {
		t.Run(value, func(t *testing.T) {
			actual, ok := parseControlledValues(value)
			if !ok {
				t.Fatal("expected controlled values to parse")
			}
			if actual != expected {
				t.Fatalf("expected %q, got %q", expected, actual)
			}
		})
	}
}

func TestParseControlledValuesRejectsUnknownValues(t *testing.T) {
	if actual, ok := parseControlledValues("Unknown"); ok {
		t.Fatalf("expected unknown controlled values to be rejected, got %q", actual)
	}
}
