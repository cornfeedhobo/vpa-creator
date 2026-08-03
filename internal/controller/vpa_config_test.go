package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	cpuResourceAnnotation              = "cpu"
	cpuMemoryResourcesAnnotation       = "cpu,memory"
	cpuMemoryMinAllowedAnnotationValue = "cpu=100m,memory=128Mi"
	cpuMemoryMaxAllowedAnnotationValue = "cpu=2,memory=4Gi"
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

func TestNewVPAUsesMinReplicas(t *testing.T) {
	minReplicas := int32(1)
	vpa := NewVPA("app-vpa", "default", "Deployment", "app", VPAConfig{
		UpdateMode:  vpav1.UpdateModeInPlaceOrRecreate,
		MinReplicas: &minReplicas,
	})

	if vpa.Spec.UpdatePolicy == nil || vpa.Spec.UpdatePolicy.MinReplicas == nil {
		t.Fatal("expected update policy min replicas to be set")
	}
	if *vpa.Spec.UpdatePolicy.MinReplicas != minReplicas {
		t.Fatalf("expected min replicas %d, got %d", minReplicas, *vpa.Spec.UpdatePolicy.MinReplicas)
	}
}

func TestNewVPAUsesContainerResourcePolicyAnnotations(t *testing.T) {
	controlledResources := []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory}
	vpa := NewVPA("app-vpa", "default", "Deployment", "app", VPAConfig{
		UpdateMode:          vpav1.UpdateModeInPlaceOrRecreate,
		ControlledResources: &controlledResources,
		MinAllowed: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		MaxAllowed: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
	})

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
	if policy.ControlledResources == nil {
		t.Fatal("expected controlled resources to be set")
	}
	assertResourceNames(t, *policy.ControlledResources, controlledResources)
	assertResourceListQuantity(t, policy.MinAllowed, corev1.ResourceCPU, "100m")
	assertResourceListQuantity(t, policy.MinAllowed, corev1.ResourceMemory, "128Mi")
	assertResourceListQuantity(t, policy.MaxAllowed, corev1.ResourceCPU, "2")
	assertResourceListQuantity(t, policy.MaxAllowed, corev1.ResourceMemory, "4Gi")
}

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
			actual, ok := ParseUpdateMode(value)
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
	if actual, ok := ParseUpdateMode("Unknown"); ok {
		t.Fatalf("expected unknown update mode to be rejected, got %q", actual)
	}
}

func TestParseUpdateModeRejectsDeprecatedAuto(t *testing.T) {
	if actual, ok := ParseUpdateMode("Auto"); ok {
		t.Fatalf("expected deprecated Auto update mode to be rejected, got %q", actual)
	}
}

func TestParseMinReplicasAcceptsPositiveInteger(t *testing.T) {
	actual, err := ParseMinReplicas("1")
	if err != nil {
		t.Fatal(err)
	}
	if actual != 1 {
		t.Fatalf("expected min replicas 1, got %d", actual)
	}
}

func TestParseMinReplicasRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"0", "-1", "nope"} {
		t.Run(value, func(t *testing.T) {
			if actual, err := ParseMinReplicas(value); err == nil {
				t.Fatalf("expected min replicas to be rejected, got %d", actual)
			}
		})
	}
}

func TestParseControlledResourcesAcceptsCommaSeparatedResources(t *testing.T) {
	actual, err := ParseControlledResources("cpu, memory")
	if err != nil {
		t.Fatal(err)
	}

	assertResourceNames(t, actual, []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory})
}

func TestParseControlledResourcesRejectsEmptyValues(t *testing.T) {
	for _, value := range []string{"", " , "} {
		t.Run(value, func(t *testing.T) {
			if actual, err := ParseControlledResources(value); err == nil {
				t.Fatalf("expected controlled resources to be rejected, got %#v", actual)
			}
		})
	}
}

func TestParseResourceListAnnotationAcceptsCommaSeparatedResources(t *testing.T) {
	actual, err := ParseResourceListAnnotation(MinAllowedAnnotation, "cpu=100m, memory=128Mi")
	if err != nil {
		t.Fatal(err)
	}

	assertResourceListQuantity(t, actual, corev1.ResourceCPU, "100m")
	assertResourceListQuantity(t, actual, corev1.ResourceMemory, "128Mi")
}

func TestParseResourceListAnnotationRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"", cpuResourceAnnotation, "cpu=", "=100m", "cpu=not-a-quantity"} {
		t.Run(value, func(t *testing.T) {
			if actual, err := ParseResourceListAnnotation(MinAllowedAnnotation, value); err == nil {
				t.Fatalf("expected resource list annotation to be rejected, got %#v", actual)
			}
		})
	}
}

func TestConfigForObjectDefaultsUpdateModeToOffWithoutAnnotation(t *testing.T) {
	deploy := &appsv1.Deployment{}

	config, err := ConfigForObject(deploy, VPAConfig{
		UpdateMode:       vpav1.UpdateModeRecreate,
		ControlledValues: vpav1.ContainerControlledValuesRequestsOnly,
	})
	if err != nil {
		t.Fatal(err)
	}

	if config.UpdateMode != vpav1.UpdateModeOff {
		t.Fatalf("expected update mode %q, got %q", vpav1.UpdateModeOff, config.UpdateMode)
	}
	if config.ControlledValues != vpav1.ContainerControlledValuesRequestsOnly {
		t.Fatalf("expected controlled values %q, got %q",
			vpav1.ContainerControlledValuesRequestsOnly,
			config.ControlledValues,
		)
	}
}

func TestConfigForObjectUsesUpdateModeAnnotation(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				UpdateModeAnnotation: string(vpav1.UpdateModeInPlaceOrRecreate),
			},
		},
	}

	config, err := ConfigForObject(deploy, VPAConfig{})
	if err != nil {
		t.Fatal(err)
	}

	if config.UpdateMode != vpav1.UpdateModeInPlaceOrRecreate {
		t.Fatalf("expected update mode %q, got %q", vpav1.UpdateModeInPlaceOrRecreate, config.UpdateMode)
	}
}

func TestConfigForObjectUsesMinReplicasAnnotation(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				MinReplicasAnnotation: "1",
			},
		},
	}

	config, err := ConfigForObject(deploy, VPAConfig{})
	if err != nil {
		t.Fatal(err)
	}

	if config.MinReplicas == nil {
		t.Fatal("expected min replicas to be set")
	}
	if *config.MinReplicas != 1 {
		t.Fatalf("expected min replicas 1, got %d", *config.MinReplicas)
	}
}

func TestConfigForObjectUsesContainerResourcePolicyAnnotations(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				ControlledResourcesAnnotation: cpuMemoryResourcesAnnotation,
				MinAllowedAnnotation:          cpuMemoryMinAllowedAnnotationValue,
				MaxAllowedAnnotation:          cpuMemoryMaxAllowedAnnotationValue,
			},
		},
	}

	config, err := ConfigForObject(deploy, VPAConfig{})
	if err != nil {
		t.Fatal(err)
	}

	if config.ControlledResources == nil {
		t.Fatal("expected controlled resources to be set")
	}
	assertResourceNames(t, *config.ControlledResources, []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory})
	assertResourceListQuantity(t, config.MinAllowed, corev1.ResourceCPU, "100m")
	assertResourceListQuantity(t, config.MinAllowed, corev1.ResourceMemory, "128Mi")
	assertResourceListQuantity(t, config.MaxAllowed, corev1.ResourceCPU, "2")
	assertResourceListQuantity(t, config.MaxAllowed, corev1.ResourceMemory, "4Gi")
}

func TestConfigForObjectOrPodTemplateUsesPodTemplateAnnotation(t *testing.T) {
	deploy := &appsv1.Deployment{}
	podTemplateAnnotations := map[string]string{
		UpdateModeAnnotation:          string(vpav1.UpdateModeInPlaceOrRecreate),
		MinReplicasAnnotation:         "1",
		ControlledResourcesAnnotation: cpuMemoryResourcesAnnotation,
		MinAllowedAnnotation:          cpuMemoryMinAllowedAnnotationValue,
		MaxAllowedAnnotation:          cpuMemoryMaxAllowedAnnotationValue,
	}

	config, err := ConfigForObjectOrPodTemplate(deploy, podTemplateAnnotations, VPAConfig{})
	if err != nil {
		t.Fatal(err)
	}

	if config.UpdateMode != vpav1.UpdateModeInPlaceOrRecreate {
		t.Fatalf("expected update mode %q, got %q", vpav1.UpdateModeInPlaceOrRecreate, config.UpdateMode)
	}
	if config.MinReplicas == nil {
		t.Fatal("expected min replicas to be set")
	}
	if *config.MinReplicas != 1 {
		t.Fatalf("expected min replicas 1, got %d", *config.MinReplicas)
	}
	if config.ControlledResources == nil {
		t.Fatal("expected controlled resources to be set")
	}
	assertResourceNames(t, *config.ControlledResources, []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory})
	assertResourceListQuantity(t, config.MinAllowed, corev1.ResourceCPU, "100m")
	assertResourceListQuantity(t, config.MinAllowed, corev1.ResourceMemory, "128Mi")
	assertResourceListQuantity(t, config.MaxAllowed, corev1.ResourceCPU, "2")
	assertResourceListQuantity(t, config.MaxAllowed, corev1.ResourceMemory, "4Gi")
}

func TestConfigForObjectOrPodTemplatePrefersObjectAnnotation(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				UpdateModeAnnotation:          string(vpav1.UpdateModeInitial),
				MinReplicasAnnotation:         "2",
				ControlledResourcesAnnotation: "memory",
				MinAllowedAnnotation:          "memory=256Mi",
				MaxAllowedAnnotation:          "memory=2Gi",
			},
		},
	}
	podTemplateAnnotations := map[string]string{
		UpdateModeAnnotation:          string(vpav1.UpdateModeInPlaceOrRecreate),
		MinReplicasAnnotation:         "1",
		ControlledResourcesAnnotation: cpuMemoryResourcesAnnotation,
		MinAllowedAnnotation:          cpuMemoryMinAllowedAnnotationValue,
		MaxAllowedAnnotation:          cpuMemoryMaxAllowedAnnotationValue,
	}

	config, err := ConfigForObjectOrPodTemplate(deploy, podTemplateAnnotations, VPAConfig{})
	if err != nil {
		t.Fatal(err)
	}

	if config.UpdateMode != vpav1.UpdateModeInitial {
		t.Fatalf("expected update mode %q, got %q", vpav1.UpdateModeInitial, config.UpdateMode)
	}
	if config.MinReplicas == nil {
		t.Fatal("expected min replicas to be set")
	}
	if *config.MinReplicas != 2 {
		t.Fatalf("expected min replicas 2, got %d", *config.MinReplicas)
	}
	if config.ControlledResources == nil {
		t.Fatal("expected controlled resources to be set")
	}
	assertResourceNames(t, *config.ControlledResources, []corev1.ResourceName{corev1.ResourceMemory})
	assertResourceListQuantity(t, config.MinAllowed, corev1.ResourceMemory, "256Mi")
	assertResourceListQuantity(t, config.MaxAllowed, corev1.ResourceMemory, "2Gi")
}

func TestConfigForObjectOrPodTemplateMergesAnnotationsByKey(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				UpdateModeAnnotation: string(vpav1.UpdateModeInitial),
			},
		},
	}
	podTemplateAnnotations := map[string]string{
		MinReplicasAnnotation:         "1",
		ControlledResourcesAnnotation: cpuResourceAnnotation,
	}

	config, err := ConfigForObjectOrPodTemplate(deploy, podTemplateAnnotations, VPAConfig{})
	if err != nil {
		t.Fatal(err)
	}

	if config.UpdateMode != vpav1.UpdateModeInitial {
		t.Fatalf("expected update mode %q, got %q", vpav1.UpdateModeInitial, config.UpdateMode)
	}
	if config.MinReplicas == nil {
		t.Fatal("expected min replicas to be set")
	}
	if *config.MinReplicas != 1 {
		t.Fatalf("expected min replicas 1, got %d", *config.MinReplicas)
	}
	if config.ControlledResources == nil {
		t.Fatal("expected controlled resources to be set")
	}
	assertResourceNames(t, *config.ControlledResources, []corev1.ResourceName{corev1.ResourceCPU})
}

func TestConfigForObjectRejectsInvalidUpdateModeAnnotation(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				UpdateModeAnnotation: "Auto",
			},
		},
	}

	if _, err := ConfigForObject(deploy, VPAConfig{}); err == nil {
		t.Fatal("expected invalid update mode annotation to be rejected")
	}
}

func TestConfigForObjectRejectsInvalidMinReplicasAnnotation(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				MinReplicasAnnotation: "0",
			},
		},
	}

	if _, err := ConfigForObject(deploy, VPAConfig{}); err == nil {
		t.Fatal("expected invalid min replicas annotation to be rejected")
	}
}

func TestConfigForObjectRejectsInvalidContainerResourcePolicyAnnotation(t *testing.T) {
	tests := map[string]map[string]string{
		"controlled-resources": {
			ControlledResourcesAnnotation: "",
		},
		"min-allowed": {
			MinAllowedAnnotation: cpuResourceAnnotation,
		},
		"max-allowed": {
			MaxAllowedAnnotation: "memory=not-a-quantity",
		},
	}

	for name, annotations := range tests {
		t.Run(name, func(t *testing.T) {
			deploy := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
				},
			}

			if _, err := ConfigForObject(deploy, VPAConfig{}); err == nil {
				t.Fatal("expected invalid container resource policy annotation to be rejected")
			}
		})
	}
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

func assertResourceNames(t *testing.T, actual, expected []corev1.ResourceName) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected resource names %#v, got %#v", expected, actual)
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Fatalf("expected resource names %#v, got %#v", expected, actual)
		}
	}
}

func assertResourceListQuantity(t *testing.T, actual corev1.ResourceList, name corev1.ResourceName, expected string) {
	t.Helper()

	quantity, ok := actual[name]
	if !ok {
		t.Fatalf("expected resource %q to be set in %#v", name, actual)
	}
	if quantity.Cmp(resource.MustParse(expected)) != 0 {
		t.Fatalf("expected %s %s, got %s", name, expected, quantity.String())
	}
}
