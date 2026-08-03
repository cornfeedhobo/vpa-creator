package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type StatefulSetReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	VPAConfig VPAConfig
}

func (r *StatefulSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var statefulSet appsv1.StatefulSet
	err := r.Get(ctx, req.NamespacedName, &statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			// StatefulSet was deleted - VPA will be automatically garbage collected by Kubernetes.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	vpaConfig, err := ConfigForObjectOrPodTemplate(
		&statefulSet,
		statefulSet.Spec.Template.GetAnnotations(),
		r.VPAConfig,
	)
	if err != nil {
		l.Error(err, "Invalid VPA configuration")
		return ctrl.Result{}, err
	}

	vpaName := fmt.Sprintf("%s-vpa", statefulSet.Name)
	vpa := NewVPA(vpaName, statefulSet.Namespace, "StatefulSet", statefulSet.Name, vpaConfig)

	created, updated, err := EnsureVPA(ctx, r.Client, r.Scheme, &statefulSet, vpa)
	if err != nil {
		l.Error(err, "Failed to ensure VPA")
		return ctrl.Result{}, err
	}

	if created {
		l.Info("Created VPA", "VPA", vpaName)
	} else if updated {
		l.Info("Updated VPA", "VPA", vpaName)
	}
	return ctrl.Result{}, nil
}

func (r *StatefulSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.StatefulSet{}).
		Owns(&vpav1.VerticalPodAutoscaler{}).
		Complete(r)
}
