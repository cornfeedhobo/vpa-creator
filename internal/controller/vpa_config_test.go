package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func TestNewVPARequestsOnlySetsWildcardContainerPolicyWithOffMode(t *testing.T) {
	vpa := NewVPA("app-vpa", "default", "Deployment", "app", VPAConfig{
		UpdateMode:       vpav1.UpdateModeOff,
		ControlledValues: vpav1.ContainerControlledValuesRequestsOnly,
	})

	if vpa.Spec.UpdatePolicy == nil || vpa.Spec.UpdatePolicy.UpdateMode == nil {
		t.Fatal("expected update policy mode to be set")
	}
	if *vpa.Spec.UpdatePolicy.UpdateMode != vpav1.UpdateModeOff {
		t.Fatalf("expected update mode %q, got %q", vpav1.UpdateModeOff, *vpa.Spec.UpdatePolicy.UpdateMode)
	}
	assertRequestsOnlyPolicy(t, vpa)
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
	assertRequestsOnlyPolicy(t, vpa)
}

func TestEnsureVPAUpdatesExistingSpec(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := vpav1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app",
			Namespace: "default",
			UID:       types.UID("test-deployment"),
		},
	}
	existing := NewVPA("app-vpa", "default", "Deployment", "app", VPAConfig{
		UpdateMode: vpav1.UpdateModeOff,
	})
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(deploy, existing).
		Build()

	desired := NewVPA("app-vpa", "default", "Deployment", "app", VPAConfig{
		UpdateMode:       vpav1.UpdateModeOff,
		ControlledValues: vpav1.ContainerControlledValuesRequestsOnly,
	})

	created, updated, err := EnsureVPA(ctx, k8sClient, scheme, deploy, desired)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("expected existing VPA to be updated, not created")
	}
	if !updated {
		t.Fatal("expected existing VPA to be updated")
	}

	var actual vpav1.VerticalPodAutoscaler
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "app-vpa", Namespace: "default"}, &actual); err != nil {
		t.Fatal(err)
	}
	assertRequestsOnlyPolicy(t, &actual)
	if len(actual.OwnerReferences) != 1 {
		t.Fatalf("expected one owner reference, got %d", len(actual.OwnerReferences))
	}
}

func assertRequestsOnlyPolicy(t *testing.T, vpa *vpav1.VerticalPodAutoscaler) {
	t.Helper()

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
