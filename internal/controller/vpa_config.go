package controller

import (
	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

type VPAConfig struct {
	UpdateMode       vpav1.UpdateMode
	ControlledValues vpav1.ContainerControlledValues
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

	if config.UpdateMode == vpav1.UpdateModeAuto &&
		config.ControlledValues == vpav1.ContainerControlledValuesRequestsOnly {
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
