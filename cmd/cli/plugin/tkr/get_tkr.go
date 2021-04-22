// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/clusterclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"
)

var getTanzuKubernetesRleasesCmd = &cobra.Command{
	Use:   "get TKR_NAME",
	Short: "Get available Tanzu Kubernetes Releases",
	Long:  "Get available Tanzu Kubernetes Releases",
	RunE:  getKubernetesReleases,
}

func getKubernetesReleases(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting TanzuKubernetesRelease with a global server is not implemented yet")
	}

	clusterClient, err := clusterclient.NewClusterClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}
	tkrName := ""
	if len(args) != 0 {
		tkrName = args[0]
	}

	tkrs, err := clusterClient.GetTanzuKubernetesReleases(tkrName)
	if err != nil {
		return err
	}

	t := component.NewTableWriter("NAME", "VERSION", "COMPATIBLE", "UPGRADEAVAILABLE")
	for i := range tkrs {
		compatible := ""
		upgradeAvailable := ""

		for _, condition := range tkrs[i].Status.Conditions {
			if condition.Type == runv1alpha1.ConditionCompatible {
				compatible = string(condition.Status)
			}
			if condition.Type == runv1alpha1.ConditionUpgradeAvailable {
				upgradeAvailable = string(condition.Status)
			}
		}

		t.Append([]string{tkrs[i].Name, tkrs[i].Spec.Version, compatible, upgradeAvailable})
	}
	t.Render()
	return nil
}
