// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conditions

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

const (
	success = "success"
)

//nolint:funlen
func NewPodExecutionCondition(kubeClient client.Client) func(context.Context, string, string, *corev1alpha2.PodExecutionCondition) (corev1alpha2.ReadinessConditionState, string) {
	return func(ctx context.Context, readinessProviderName string, conditionName string, c *corev1alpha2.PodExecutionCondition) (corev1alpha2.ReadinessConditionState, string) {
		key := fmt.Sprintf("%s%s", readinessProviderName, conditionName)
		h := sha256.New()
		h.Write([]byte(key))
		bs := h.Sum(nil)
		podName := fmt.Sprintf("pod-%x", string(bs))[0:20]

		stateConfig := &v1.ConfigMap{}
		err := kubeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      podName,
		}, stateConfig)

		generation := 0

		if err != nil {
			fmt.Println("## Configmap does not exist; creating config map")
			ierr := kubeClient.Create(ctx, &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: "default",
				},
				Data: map[string]string{
					"lastState":  "inprogress",
					"generation": "1",
				},
			})

			if ierr != nil {
				fmt.Println("## Configmap creation error", ierr)
				return corev1alpha2.ConditionFailureState, ierr.Error()
			}

			generation = 1
		} else {
			fmt.Println("## Configmap already there; not creating new one")
			generation, _ = strconv.Atoi(stateConfig.Data["generation"])
		}

		pod := &v1.Pod{}
		err = kubeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      fmt.Sprintf("%s-%d", podName, generation),
		}, pod)

		if err != nil {
			pod = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", podName, generation),
					Namespace: "default",
				},
				Spec: c.PodSpec,
			}
			err := kubeClient.Create(ctx, pod)
			if err != nil {
				return corev1alpha2.ConditionFailureState, err.Error()
			}
		} else {
			if pod.Status.Phase == v1.PodSucceeded {
				stateConfig1 := &v1.ConfigMap{}
				err = kubeClient.Get(ctx, types.NamespacedName{
					Namespace: "default",
					Name:      podName,
				}, stateConfig1)
				if err != nil {
					return corev1alpha2.ConditionFailureState, err.Error()
				}

				stateConfig1.Data["lastState"] = success
				stateConfig1.Data["generation"] = fmt.Sprintf("%d", generation+1)

				err = kubeClient.Update(ctx, stateConfig1)
				if err != nil {
					return corev1alpha2.ConditionFailureState, err.Error()
				}
			} else if pod.Status.Phase == v1.PodFailed {
				stateConfig1 := &v1.ConfigMap{}
				err = kubeClient.Get(ctx, types.NamespacedName{
					Namespace: "default",
					Name:      podName,
				}, stateConfig1)
				if err != nil {
					return corev1alpha2.ConditionFailureState, err.Error()
				}

				stateConfig1.Data["lastState"] = "failure"
				stateConfig1.Data["generation"] = fmt.Sprintf("%d", generation+1)

				err = kubeClient.Update(ctx, stateConfig1)
				if err != nil {
					return corev1alpha2.ConditionFailureState, err.Error()
				}
			}
		}

		err = kubeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      podName,
		}, stateConfig)

		if err != nil {
			return corev1alpha2.ConditionFailureState, err.Error()
		}

		if stateConfig.Data["lastState"] == "inprogress" {
			return corev1alpha2.ConditionInProgressState, ""
		} else if stateConfig.Data["lastState"] == success {
			return corev1alpha2.ConditionSuccessState, ""
		} else {
			return corev1alpha2.ConditionFailureState, ""
		}
	}
}
