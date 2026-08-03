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

type DeploymentReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	VPAConfig VPAConfig
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var deploy appsv1.Deployment
	err := r.Get(ctx, req.NamespacedName, &deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			// Deployment was deleted — VPA will be automatically garbage collected by Kubernetes
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	vpaConfig, err := ConfigForObject(&deploy, r.VPAConfig)
	if err != nil {
		l.Error(err, "Invalid VPA configuration")
		return ctrl.Result{}, err
	}

	vpaName := fmt.Sprintf("%s-vpa", deploy.Name)
	vpa := NewVPA(vpaName, deploy.Namespace, "Deployment", deploy.Name, vpaConfig)

	created, updated, err := EnsureVPA(ctx, r.Client, r.Scheme, &deploy, vpa)
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

func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Owns(&vpav1.VerticalPodAutoscaler{}).
		Complete(r)
}
