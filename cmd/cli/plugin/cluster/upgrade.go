// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/clusterclient"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/constants"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
)

type upgradeClustersOptions struct {
	namespace  string
	tkrName    string
	timeout    time.Duration
	unattended bool
}

var uc = &upgradeClustersOptions{}

var upgradeClusterCmd = &cobra.Command{
	Use:   "upgrade CLUSTER_NAME",
	Short: "Upgrade a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  upgrade,
}

func init() {
	upgradeClusterCmd.Flags().StringVarP(&uc.tkrName, "tkr", "", "", "TanzuKubernetesRelease(tkr) to upgrade to")
	upgradeClusterCmd.Flags().StringVarP(&uc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified")
	upgradeClusterCmd.Flags().DurationVarP(&uc.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	upgradeClusterCmd.Flags().BoolVarP(&uc.unattended, "yes", "y", false, "Upgrade workload cluster without asking for confirmation")
}

func upgrade(cmd *cobra.Command, args []string) error {
	server, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("upgrading cluster with a global server is not implemented yet")
	}
	return upgradeCluster(server, args[0])
}

func upgradeCluster(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	clusterClient, err := clusterclient.NewClusterClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	kubernetesVersion := ""
	if uc.tkrName != "" {
		kubernetesVersion, err = getValidK8sVersionFromTkrForUpgrade(tkgctlClient, clusterClient, clusterName)
		if err != nil {
			return err
		}
		fmt.Printf("upgrading cluster to kubernetes version %q \n", kubernetesVersion)
	}

	upgradeClusterOptions := tkgctl.UpgradeClusterOptions{
		ClusterName:       clusterName,
		Namespace:         uc.namespace,
		KubernetesVersion: kubernetesVersion,
		SkipPrompt:        uc.unattended,
		Timeout:           uc.timeout,
	}

	return tkgctlClient.UpgradeCluster(upgradeClusterOptions)
}

func getValidK8sVersionFromTkrForUpgrade(tkgctlClient tkgctl.TKGClient, clusterClient clusterclient.Client, clusterName string) (string, error) {
	result, err := tkgctlClient.DescribeCluster(tkgctl.DescribeTKGClustersOptions{
		ClusterName: clusterName,
		Namespace:   uc.namespace,
	})
	if err != nil {
		return "", err
	}

	tkrName, ok := result.Cluster.Labels["tanzuKubernetesRelease"]
	if !ok {
		return "", errors.Errorf("unable to obtain TanzuKubernetesRelease for cluster %q, namespace %q", clusterName, uc.namespace)
	}

	tkr, err := getMatchingTkrForTkrName(clusterClient, tkrName)
	if err != nil {
		return "", err
	}

	upgradeMsg := ""
	for _, condition := range tkr.Status.Conditions {
		if condition.Type == runv1alpha1.ConditionUpgradeAvailable {
			upgradeMsg = condition.Message
		}
	}

	tkrAvailableUpgrades := make([]string, 0)
	if strs := strings.Split(upgradeMsg, ": "); len(strs) != 2 {
		return "", errors.Errorf("no available upgrades for cluster %q, namespace %q", clusterName, uc.namespace)
	} else {
		tkrAvailableUpgrades = strings.Split(strs[1], ",")
	}

	for _, availableUpgrade := range tkrAvailableUpgrades {
		if availableUpgrade == uc.tkrName {
			tkrForUpgrade, err := getMatchingTkrForTkrName(clusterClient, uc.tkrName)
			if err != nil {
				return "", nil
			}

			return tkrForUpgrade.Spec.Version, nil
		}
	}

	return "", errors.Errorf("cluster cannot be upgraded to %q, available upgrades %v", uc.tkrName, tkrAvailableUpgrades)
}

func getMatchingTkrForTkrName(clusterClient clusterclient.Client, tkrName string) (*runv1alpha1.TanzuKubernetesRelease, error) {
	tkrs, err := clusterClient.GetTanzuKubernetesReleases(tkrName)
	if err != nil {
		return nil, err
	}

	for _, tkr := range tkrs {
		if tkr.Name == tkrName {
			return &tkr, err
		}
	}

	return nil, errors.Errorf("could not find a matching TanzuKubernetesRelease for name %q", tkrName)
}
