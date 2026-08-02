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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	vpaName := fmt.Sprintf("%s-vpa", statefulSet.Name)

	// Check if VPA already exists
	var existingVPA vpav1.VerticalPodAutoscaler
	err = r.Get(ctx, client.ObjectKey{Name: vpaName, Namespace: statefulSet.Namespace}, &existingVPA)
	if err == nil {
		// VPA exists
		return ctrl.Result{}, nil
	} else if !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	vpa := NewVPA(vpaName, statefulSet.Namespace, "StatefulSet", statefulSet.Name, r.VPAConfig)

	// Set the StatefulSet as the owner of the VPA for automatic garbage collection.
	if err := controllerutil.SetControllerReference(&statefulSet, vpa, r.Scheme); err != nil {
		l.Error(err, "Failed to set controller reference")
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, vpa); err != nil {
		l.Error(err, "Failed to create VPA")
		return ctrl.Result{}, err
	}

	l.Info("Created VPA", "VPA", vpaName)
	return ctrl.Result{}, nil
}

func (r *StatefulSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.StatefulSet{}).
		Owns(&vpav1.VerticalPodAutoscaler{}).
		Complete(r)
}
