package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	UpdateModeAnnotation           = "vpa-creator.cornfeedhobo/update-mode"
	MinReplicasAnnotation          = "vpa-creator.cornfeedhobo/min-replicas"
	ControlledResourcesAnnotation  = "vpa-creator.cornfeedhobo/controlled-resources"
	MinAllowedAnnotation           = "vpa-creator.cornfeedhobo/min-allowed"
	MaxAllowedAnnotation           = "vpa-creator.cornfeedhobo/max-allowed"
	ValidUpdateModes               = "Off, Initial, Recreate, InPlaceOrRecreate, InPlace"
	ValidContainerResourcePolicyKV = "must be a comma-separated resource list, for example: cpu=100m,memory=128Mi"
)

type VPAConfig struct {
	UpdateMode          vpav1.UpdateMode
	MinReplicas         *int32
	ControlledValues    vpav1.ContainerControlledValues
	ControlledResources *[]corev1.ResourceName
	MinAllowed          corev1.ResourceList
	MaxAllowed          corev1.ResourceList
}

func ConfigForObject(obj client.Object, config VPAConfig) (VPAConfig, error) {
	return ConfigForAnnotations(obj.GetAnnotations(), config)
}

func ConfigForObjectOrPodTemplate(obj client.Object, podTemplateAnnotations map[string]string, config VPAConfig) (VPAConfig, error) {
	annotations := map[string]string{}
	for key, value := range podTemplateAnnotations {
		annotations[key] = value
	}
	for key, value := range obj.GetAnnotations() {
		annotations[key] = value
	}

	return ConfigForAnnotations(annotations, config)
}

func ConfigForAnnotations(annotations map[string]string, config VPAConfig) (VPAConfig, error) {
	updateMode, ok := annotations[UpdateModeAnnotation]
	if ok {
		parsed, ok := ParseUpdateMode(updateMode)
		if !ok {
			return VPAConfig{}, fmt.Errorf(
				"invalid %s annotation value %q, must be one of: %s",
				UpdateModeAnnotation,
				updateMode,
				ValidUpdateModes,
			)
		}

		config.UpdateMode = parsed
	} else {
		config.UpdateMode = vpav1.UpdateModeOff
	}

	minReplicas, ok := annotations[MinReplicasAnnotation]
	if ok {
		parsed, err := ParseMinReplicas(minReplicas)
		if err != nil {
			return VPAConfig{}, err
		}

		config.MinReplicas = &parsed
	}

	controlledResources, ok := annotations[ControlledResourcesAnnotation]
	if ok {
		parsed, err := ParseControlledResources(controlledResources)
		if err != nil {
			return VPAConfig{}, err
		}

		config.ControlledResources = &parsed
	}

	minAllowed, ok := annotations[MinAllowedAnnotation]
	if ok {
		parsed, err := ParseResourceListAnnotation(MinAllowedAnnotation, minAllowed)
		if err != nil {
			return VPAConfig{}, err
		}

		config.MinAllowed = parsed
	}

	maxAllowed, ok := annotations[MaxAllowedAnnotation]
	if ok {
		parsed, err := ParseResourceListAnnotation(MaxAllowedAnnotation, maxAllowed)
		if err != nil {
			return VPAConfig{}, err
		}

		config.MaxAllowed = parsed
	}

	return config, nil
}

func ParseUpdateMode(value string) (vpav1.UpdateMode, bool) {
	switch value {
	case string(vpav1.UpdateModeOff):
		return vpav1.UpdateModeOff, true
	case string(vpav1.UpdateModeInitial):
		return vpav1.UpdateModeInitial, true
	case string(vpav1.UpdateModeRecreate):
		return vpav1.UpdateModeRecreate, true
	case string(vpav1.UpdateModeInPlaceOrRecreate):
		return vpav1.UpdateModeInPlaceOrRecreate, true
	case string(vpav1.UpdateModeInPlace):
		return vpav1.UpdateModeInPlace, true
	default:
		return "", false
	}
}

func ParseMinReplicas(value string) (int32, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid %s annotation value %q, must be a positive integer", MinReplicasAnnotation, value)
	}

	return int32(parsed), nil
}

func ParseControlledResources(value string) ([]corev1.ResourceName, error) {
	var resources []corev1.ResourceName
	for _, part := range strings.Split(value, ",") {
		resourceName := strings.TrimSpace(part)
		if resourceName == "" {
			continue
		}

		resources = append(resources, corev1.ResourceName(resourceName))
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf(
			"invalid %s annotation value %q, must include at least one resource",
			ControlledResourcesAnnotation,
			value,
		)
	}

	return resources, nil
}

func ParseResourceListAnnotation(annotation, value string) (corev1.ResourceList, error) {
	resourceList := corev1.ResourceList{}
	for _, part := range strings.Split(value, ",") {
		resourceName, quantity, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid %s annotation value %q, %s", annotation, value, ValidContainerResourcePolicyKV)
		}

		resourceName = strings.TrimSpace(resourceName)
		quantity = strings.TrimSpace(quantity)
		if resourceName == "" || quantity == "" {
			return nil, fmt.Errorf("invalid %s annotation value %q, %s", annotation, value, ValidContainerResourcePolicyKV)
		}

		parsed, err := resource.ParseQuantity(quantity)
		if err != nil {
			return nil, fmt.Errorf("invalid %s annotation value %q: %w", annotation, value, err)
		}

		resourceList[corev1.ResourceName(resourceName)] = parsed
	}

	if len(resourceList) == 0 {
		return nil, fmt.Errorf("invalid %s annotation value %q, %s", annotation, value, ValidContainerResourcePolicyKV)
	}

	return resourceList, nil
}

func NewVPA(name, namespace, targetKind, targetName string, config VPAConfig) *vpav1.VerticalPodAutoscaler {
	if config.UpdateMode == "" {
		config.UpdateMode = vpav1.UpdateModeOff
	}

	vpa := &vpav1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef: &v1.CrossVersionObjectReference{
				Kind:       targetKind,
				Name:       targetName,
				APIVersion: "apps/v1",
			},
			UpdatePolicy: &vpav1.PodUpdatePolicy{
				UpdateMode:  &config.UpdateMode,
				MinReplicas: config.MinReplicas,
			},
		},
	}

	if needsResourcePolicy(config) {
		containerPolicy := vpav1.ContainerResourcePolicy{
			ContainerName:       vpav1.DefaultContainerResourcePolicy,
			ControlledResources: config.ControlledResources,
			MinAllowed:          config.MinAllowed,
			MaxAllowed:          config.MaxAllowed,
		}

		if config.ControlledValues == vpav1.ContainerControlledValuesRequestsOnly {
			controlledValues := vpav1.ContainerControlledValuesRequestsOnly
			containerPolicy.ControlledValues = &controlledValues
		}

		vpa.Spec.ResourcePolicy = &vpav1.PodResourcePolicy{
			ContainerPolicies: []vpav1.ContainerResourcePolicy{
				containerPolicy,
			},
		}
	}

	return vpa
}

func needsResourcePolicy(config VPAConfig) bool {
	return config.ControlledValues == vpav1.ContainerControlledValuesRequestsOnly ||
		config.ControlledResources != nil ||
		len(config.MinAllowed) > 0 ||
		len(config.MaxAllowed) > 0
}

func EnsureVPA(
	ctx context.Context,
	k8sClient client.Client,
	scheme *runtime.Scheme,
	owner client.Object,
	desired *vpav1.VerticalPodAutoscaler,
) (bool, bool, error) {
	if err := controllerutil.SetControllerReference(owner, desired, scheme); err != nil {
		return false, false, err
	}

	var existing vpav1.VerticalPodAutoscaler
	err := k8sClient.Get(ctx, client.ObjectKeyFromObject(desired), &existing)
	if errors.IsNotFound(err) {
		if err := k8sClient.Create(ctx, desired); err != nil {
			return false, false, err
		}
		return true, false, nil
	}
	if err != nil {
		return false, false, err
	}

	before := existing.DeepCopy()
	if err := controllerutil.SetControllerReference(owner, &existing, scheme); err != nil {
		return false, false, err
	}
	existing.Spec = desired.Spec

	if apiequality.Semantic.DeepEqual(before.Spec, existing.Spec) &&
		apiequality.Semantic.DeepEqual(before.OwnerReferences, existing.OwnerReferences) {
		return false, false, nil
	}

	if err := k8sClient.Update(ctx, &existing); err != nil {
		return false, false, err
	}
	return false, true, nil
}
