// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package capabilities

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

const (
	requeueInterval = 60 * time.Second
	contextTimeout  = 60 * time.Second
)

type CapabilityReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *CapabilityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctxCancel, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	log := r.Log.WithValues("readinessprovider", req.NamespacedName)
	log.Info("starting reconcile")

	capability := &corev1alpha2.NewCapability{}
	err := r.Client.Get(ctxCancel, req.NamespacedName, capability)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	readinessProviders := corev1alpha2.ReadinessProviderList{}
	err = r.Client.List(ctxCancel, &readinessProviders)
	if err != nil {
		return ctrl.Result{}, err
	}

	found := false
	var ownedReadinessProvier *corev1alpha2.ReadinessProvider
	for _, readinessProvider := range readinessProviders.Items {
		for _, ownerReference := range readinessProvider.OwnerReferences {
			if ownerReference.APIVersion != capability.APIVersion {
				continue
			}

			if ownerReference.Kind != capability.Kind {
				continue
			}

			if ownerReference.Name == capability.Name {
				found = true
				ownedReadinessProvier = &readinessProvider
			}
		}
	}

	i := 1
	conditions := []corev1alpha2.ReadinessProviderCondition{}
	for _, gvr := range capability.Spec.GVRs {
		conditions = append(conditions, corev1alpha2.ReadinessProviderCondition{
			Name: fmt.Sprintf("condition-%d", i),
			ResourceExistenceCondition: &corev1alpha2.ResourceExistenceCondition{
				APIVersion: gvr.APIVersion,
				Kind:       gvr.Kind,
				Namespace:  gvr.Namespace,
				Name:       gvr.Name,
			},
		})
		i++
	}

	if !found {
		newReadinessProvider := corev1alpha2.ReadinessProvider{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "capability-",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: capability.APIVersion,
						Kind:       capability.Kind,
						Name:       capability.Name,
						UID:        capability.UID,
					},
				},
			},
			Spec: corev1alpha2.ReadinessProviderSpec{
				CheckRefs:  []string{},
				Conditions: conditions,
			},
		}

		err = r.Client.Create(ctxCancel, &newReadinessProvider)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		ownedReadinessProvier.Spec.Conditions = conditions
		err = r.Client.Update(ctxCancel, ownedReadinessProvier)
		if err != nil {
			return ctrl.Result{}, err
		}

	}

	if found {
		if ownedReadinessProvier.Status.State == corev1alpha2.ProviderSuccessState {
			capability.Status.Present = true
		} else {
			capability.Status.Present = false
		}
	}

	log.Info("Successfully reconciled")

	return ctrl.Result{}, r.Status().Update(ctxCancel, capability)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CapabilityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha2.NewCapability{}).
		Watches(
			&source.Kind{Type: &corev1alpha2.ReadinessProvider{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForReadinessProvider),
		).
		Complete(r)
}

func (r *CapabilityReconciler) findObjectsForReadinessProvider(readinessProviderObject client.Object) []reconcile.Request {

	provider, _ := readinessProviderObject.(*corev1alpha2.ReadinessProvider)

	requests := []reconcile.Request{}
	for _, ownerReference := range provider.OwnerReferences {
		if ownerReference.APIVersion == "core.tanzu.vmware.com/v1alpha2" && ownerReference.Kind == "NewCapability" {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: ownerReference.Name,
				},
			})
		}
	}

	return requests
}
