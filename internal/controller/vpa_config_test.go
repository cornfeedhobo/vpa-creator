package controller

import (
	"testing"

	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func TestNewVPADefaultsToOffWithoutResourcePolicy(t *testing.T) {
	vpa := NewVPA("app-vpa", "default", "Deployment", "app", VPAConfig{})

	if vpa.Spec.UpdatePolicy == nil || vpa.Spec.UpdatePolicy.UpdateMode == nil {
		t.Fatal("expected update policy mode to be set")
	}
	if *vpa.Spec.UpdatePolicy.UpdateMode != vpav1.UpdateModeOff {
		t.Fatalf("expected update mode %q, got %q", vpav1.UpdateModeOff, *vpa.Spec.UpdatePolicy.UpdateMode)
	}
	if vpa.Spec.ResourcePolicy != nil {
		t.Fatalf("expected no resource policy, got %#v", vpa.Spec.ResourcePolicy)
	}
}

func TestNewVPARequestsOnlySetsWildcardContainerPolicyForApplyingModes(t *testing.T) {
	vpa := NewVPA("app-vpa", "default", "Deployment", "app", VPAConfig{
		UpdateMode:       vpav1.UpdateModeRecreate,
		ControlledValues: vpav1.ContainerControlledValuesRequestsOnly,
	})

	if vpa.Spec.UpdatePolicy == nil || vpa.Spec.UpdatePolicy.UpdateMode == nil {
		t.Fatal("expected update policy mode to be set")
	}
	if *vpa.Spec.UpdatePolicy.UpdateMode != vpav1.UpdateModeRecreate {
		t.Fatalf("expected update mode %q, got %q", vpav1.UpdateModeRecreate, *vpa.Spec.UpdatePolicy.UpdateMode)
	}
	if vpa.Spec.ResourcePolicy == nil {
		t.Fatal("expected resource policy to be set")
	}
	if len(vpa.Spec.ResourcePolicy.ContainerPolicies) != 1 {
		t.Fatalf("expected one container policy, got %d", len(vpa.Spec.ResourcePolicy.ContainerPolicies))
	}

	policy := vpa.Spec.ResourcePolicy.ContainerPolicies[0]
	if policy.ContainerName != vpav1.DefaultContainerResourcePolicy {
		t.Fatalf("expected wildcard container policy, got %q", policy.ContainerName)
	}
	if policy.ControlledValues == nil {
		t.Fatal("expected controlled values to be set")
	}
	if *policy.ControlledValues != vpav1.ContainerControlledValuesRequestsOnly {
		t.Fatalf("expected controlled values %q, got %q",
			vpav1.ContainerControlledValuesRequestsOnly,
			*policy.ControlledValues)
	}
}
