// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type E2EMhcCCSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
}

func E2EMhcCCSpec(context context.Context, inputGetter func() E2EMhcCCSpecInput) { //nolint:funlen
	var (
		input        E2EMhcCCSpecInput
		tkgCtlClient tkgctl.TKGClient
		logsDir      string
		clusterName  string
		namespace    string

		mcProxy *framework.ClusterProxy
		wcProxy *framework.ClusterProxy
	)

	BeforeEach(func() {
		var err error
		namespace = constants.DefaultNamespace
		input = inputGetter()
		logsDir = filepath.Join(input.ArtifactsFolder, "logs")

		mcClusterName := input.E2EConfig.ManagementClusterName
		mcContextName := mcClusterName + "-admin@" + mcClusterName
		mcProxy = framework.NewClusterProxy(mcClusterName, "", mcContextName)

		rand.Seed(time.Now().UnixNano())
		clusterName = input.E2EConfig.ClusterPrefix + "wc"

		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})

		Expect(err).To(BeNil())

		By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
		err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
		})
		Expect(err).To(BeNil())

		wcContextName := clusterName + "-admin@" + clusterName
		wcProxy = framework.NewClusterProxy(clusterName, "", wcContextName)
	})

	It("mhc should remediate unhealthy machine", func() {
		// Validate MHC
		By(fmt.Sprintf("Getting MHC for cluster %q", clusterName))
		mhcList, err := tkgCtlClient.GetMachineHealthCheck(tkgctl.GetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			MachineHealthCheckName: clusterName,
			Namespace:              namespace,
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(mhcList)).To(Equal(1))
		mhc := mhcList[0]
		Expect(mhc.Spec.ClusterName).To(Equal(clusterName))
		Expect(mhc.Name).To(Equal(clusterName))
		Expect(len(mhc.Spec.UnhealthyConditions)).To(Equal(2)) // nolint:gomnd

		// Delete MHC and verify if MHC is deleted
		By(fmt.Sprintf("Deleting MHC for cluster %q", clusterName))
		if tkgCtlClient == nil {
			_, _ = GinkgoWriter.Write([]byte("tkgCtlClient is nil"))
		}
		err = tkgCtlClient.DeleteMachineHealthCheck(tkgctl.DeleteMachineHealthCheckOptions{
			ClusterName:            clusterName,
			MachinehealthCheckName: clusterName,
			Namespace:              namespace,
			SkipPrompt:             true,
		})
		Expect(err).ToNot(HaveOccurred())

		mhcList, err = tkgCtlClient.GetMachineHealthCheck(tkgctl.GetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			Namespace:              namespace,
			MachineHealthCheckName: clusterName,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mhcList)).To(Equal(0))

		// Set MHC and verify if it is set
		By(fmt.Sprintf("Updating MHC for cluster %q", clusterName))
		err = tkgCtlClient.SetMachineHealthCheck(tkgctl.SetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			Namespace:              namespace,
			MachineHealthCheckName: clusterName,
			UnhealthyConditions:    fmt.Sprintf("%s:%s:%s", string(corev1.NodeReady), string(corev1.ConditionFalse), "5m"),
		})
		Expect(err).ToNot(HaveOccurred())

		// Wait for Target in MHC status to get the machine name
		waitForMhcTarget(tkgCtlClient, clusterName, namespace)
		mhcList, err = tkgCtlClient.GetMachineHealthCheck(tkgctl.GetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			Namespace:              namespace,
			MachineHealthCheckName: clusterName,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mhcList)).To(Equal(1))

		mhc = mhcList[0]
		Expect(mhc.Spec.ClusterName).To(Equal(clusterName))
		Expect(mhc.Name).To(Equal(clusterName))
		Expect(len(mhc.Spec.UnhealthyConditions)).To(Equal(1))

		// Set machine to unhealthy and see if that machine is remediated
		Expect(len(mhc.Status.Targets)).To(Equal(1))
		machine := mhc.Status.Targets[0]
		By(fmt.Sprintf("Patching Node to make it fail the MHC %q", machine))
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Context : %s \n", context)))
		patchNodeUnhealthy(context, wcProxy, machine, "", mcProxy)

		By("Waiting for the Node to be remediated")
		WaitForNodeRemediation(context, clusterName, "", mcProxy, wcProxy)
	})
}
