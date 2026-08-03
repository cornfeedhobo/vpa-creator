package controller

import (
	"context"
	"fmt"

	v1 "k8s.io/api/autoscaling/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	UpdateModeAnnotation = "vpa-creator.cornfeedhobo/update-mode"
	ValidUpdateModes     = "Off, Initial, Recreate, InPlaceOrRecreate, InPlace"
)

type VPAConfig struct {
	UpdateMode       vpav1.UpdateMode
	ControlledValues vpav1.ContainerControlledValues
}

func ConfigForObject(obj client.Object, config VPAConfig) (VPAConfig, error) {
	return ConfigForAnnotations(obj.GetAnnotations(), config)
}

func ConfigForObjectOrPodTemplate(obj client.Object, podTemplateAnnotations map[string]string, config VPAConfig) (VPAConfig, error) {
	if _, ok := obj.GetAnnotations()[UpdateModeAnnotation]; ok {
		return ConfigForObject(obj, config)
	}

	return ConfigForAnnotations(podTemplateAnnotations, config)
}

func ConfigForAnnotations(annotations map[string]string, config VPAConfig) (VPAConfig, error) {
	updateMode, ok := annotations[UpdateModeAnnotation]
	if !ok {
		config.UpdateMode = vpav1.UpdateModeOff
		return config, nil
	}

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
				UpdateMode: &config.UpdateMode,
			},
		},
	}

	if config.ControlledValues == vpav1.ContainerControlledValuesRequestsOnly {
		controlledValues := vpav1.ContainerControlledValuesRequestsOnly
		vpa.Spec.ResourcePolicy = &vpav1.PodResourcePolicy{
			ContainerPolicies: []vpav1.ContainerResourcePolicy{
				{
					ContainerName:    vpav1.DefaultContainerResourcePolicy,
					ControlledValues: &controlledValues,
				},
			},
		}
	}

	return vpa
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
