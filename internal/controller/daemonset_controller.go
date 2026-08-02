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

type DaemonSetReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	VPAConfig VPAConfig
}

func (r *DaemonSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var daemonSet appsv1.DaemonSet
	err := r.Get(ctx, req.NamespacedName, &daemonSet)
	if err != nil {
		if errors.IsNotFound(err) {
			// DaemonSet was deleted - VPA will be automatically garbage collected by Kubernetes.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	vpaName := fmt.Sprintf("%s-vpa", daemonSet.Name)

	// Check if VPA already exists
	var existingVPA vpav1.VerticalPodAutoscaler
	err = r.Get(ctx, client.ObjectKey{Name: vpaName, Namespace: daemonSet.Namespace}, &existingVPA)
	if err == nil {
		// VPA exists
		return ctrl.Result{}, nil
	} else if !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	vpa := NewVPA(vpaName, daemonSet.Namespace, "DaemonSet", daemonSet.Name, r.VPAConfig)

	// Set the DaemonSet as the owner of the VPA for automatic garbage collection.
	if err := controllerutil.SetControllerReference(&daemonSet, vpa, r.Scheme); err != nil {
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

func (r *DaemonSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		Owns(&vpav1.VerticalPodAutoscaler{}).
		Complete(r)
}
