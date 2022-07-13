package unit

import (
	"fmt"
	"github.com/vmware-tanzu/tanzu-framework/test/pkg/matchers"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

var _ = Describe("PCI Passthrough", func() {
	var paths []string
	var baseVal yttValues

	BeforeEach(func() {
		paths = []string{
			filepath.Join(yamlRoot, "config_default.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "overlay.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "base-template.yaml"),
			filepath.Join("./fixtures/tkr-bom-v1.21.1.yaml"),
			filepath.Join("./fixtures/tkg-bom-v1.4.0.yaml"),
			filepath.Join(yamlRoot, "ytt"),
		}

		baseVal = map[string]interface{}{
			// required fields
			"TKG_DEFAULT_BOM":    "tkg-bom-v1.4.0.yaml",
			"KUBERNETES_RELEASE": "v1.21.2---vmware.1-tkg.1",
			"CLUSTER_NAME":       "test-cluster",

			// required fields for CAPV
			"PROVIDER_TYPE":    "vsphere",
			"TKG_CLUSTER_ROLE": "management",
			"TKG_IP_FAMILY":    "ipv4",
			"SERVICE_CIDR":     "5.5.5.5/16",

			// required vsphere configurations
			"VSPHERE_USERNAME":           "user_blah",
			"VSPHERE_PASSWORD":           "pass_1234",
			"VSPHERE_SERVER":             "vmware-tanzu.com",
			"VSPHERE_DATACENTER":         "vmware-tanzu-dc.com",
			"VSPHERE_RESOURCE_POOL":      "myrp",
			"VSPHERE_FOLDER":             "ds0",
			"VSPHERE_SSH_AUTHORIZED_KEY": "ssh-rsa AAAA...+M7Q== vmware-tanzu.local",
			"VSPHERE_INSECURE":           "true",
			"CLUSTER_CIDR":               "192.168.1.0/16",
		}
	})

	When("basic values are provided", func() {
		It("renders without error", func() {
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, baseVal.toReader())
			Expect(err).NotTo(HaveOccurred())
		})

		It("has no PCI devices in worker VSphereMachineTemplate", func() {
			output, _ := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, baseVal.toReader())
			vsphereMachineTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "test-cluster-worker",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(len(vsphereMachineTemplates)).NotTo(Equal(0))

			for _, vsphereMachineTemplate := range vsphereMachineTemplates {
				Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices", ""))
			}
		})

		It("has no PCI devices in control-plane VSphereMachineTemplate", func() {
			output, _ := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, baseVal.toReader())
			vsphereMachineTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "test-cluster-control-plane",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(len(vsphereMachineTemplates)).NotTo(Equal(0))

			for _, vsphereMachineTemplate := range vsphereMachineTemplates {
				Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices", ""))
			}
		})
	})

	When("VSPHERE_WORKER_PCI_DEVICES and VSPHERE_CONTROL_PLANE_PCI_DEVICES are set", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("throws error when the input format is invalid", func() {
			invalidValues := []string{"a:b;c;d", "sometext", "1001:1001,2000:2001;400:300"}

			for _, invalidValueTestCase := range invalidValues {
				value.Set("VSPHERE_WORKER_PCI_DEVICES", invalidValueTestCase)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
				Expect(err).To(HaveOccurred())
			}
		})

		It("throws error when the vendor devices are invalid", func() {
			value.Set("VSPHERE_WORKER_PCI_DEVICES", "a:b")
			value.Set("VSPHERE CONTROL_PLANE_PCI_DEVICES", "c:d")
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).To(HaveOccurred())
		})

		It("succeeds when the vendor devices are valid", func() {
			value.Set("VSPHERE_WORKER_PCI_DEVICES", "10DE:1EB8")
			value.Set("VSPHERE_CONTROL_PLANE_PCI_DEVICES", "10DE:1EB8")

			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			clusterTypes := []string{"worker", "control-plane"}
			for _, clusterType := range clusterTypes {
				vsphereMachineTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
					"$.kind":          "VSphereMachineTemplate",
					"$.metadata.name": fmt.Sprintf("test-cluster-%s", clusterType),
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(len(vsphereMachineTemplates)).NotTo(Equal(0))

				for _, vsphereMachineTemplate := range vsphereMachineTemplates {
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[0].vendorID", "10DE"))
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[0].deviceID", "1EB8"))
				}
			}
		})

		When("VSPHERE_IGNORE_PCI_DEVICES_ALLOW_LIST is true", func() {
			It("succeeds even when the vendor devices are not valid", func() {
				value.Set("VSPHERE_WORKER_PCI_DEVICES", "a:b,c:d")
				value.Set("VSPHERE_CONTROL_PLANE_PCI_DEVICES", "e:f,g:h")
				value.Set("VSPHERE_IGNORE_PCI_DEVICES_ALLOW_LIST", true)
				output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
				Expect(err).NotTo(HaveOccurred())

				vsphereMachineTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
					"$.kind":          "VSphereMachineTemplate",
					"$.metadata.name": "test-cluster-worker",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(len(vsphereMachineTemplates)).NotTo(Equal(0))

				for _, vsphereMachineTemplate := range vsphereMachineTemplates {
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[0].vendorID", "a"))
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[0].deviceID", "b"))
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[1].vendorID", "c"))
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[1].deviceID", "d"))
				}

				vsphereMachineTemplates, err = matchers.FindDocsMatchingYAMLPath(output, map[string]string{
					"$.kind":          "VSphereMachineTemplate",
					"$.metadata.name": "test-cluster-control-plane",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(len(vsphereMachineTemplates)).NotTo(Equal(0))

				for _, vsphereMachineTemplate := range vsphereMachineTemplates {
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[0].vendorID", "e"))
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[0].deviceID", "f"))
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[1].vendorID", "g"))
					Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.pciDevices[1].deviceID", "h"))
				}
			})
		})

	})

	When("WORKER_ROLLOUT_STRATEGY is set", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("should return validation error if the input is invalid", func() {
			value.Set("WORKER_ROLLOUT_STRATEGY", "somerandomvalue")
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).To(HaveOccurred())
		})

		It("should set MachineDeployment strategy type to correct value", func() {
			validValues := []string{"OnDelete", "RollingUpdate"}
			for _, validValue := range validValues {
				value.Set("WORKER_ROLLOUT_STRATEGY", validValue)
				output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
				Expect(err).NotTo(HaveOccurred())

				machineDeploymentTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
					"$.kind": "MachineDeployment",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(len(machineDeploymentTemplates)).NotTo(Equal(0))

				for _, machineDeploymentTemplate := range machineDeploymentTemplates {
					Expect(machineDeploymentTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.strategy.type", validValue))
				}
			}
		})
	})

	When("WORKER_ROLLOUT_STRATEGY is not set", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("defaults to rolling update", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			machineDeploymentTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind": "MachineDeployment",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(len(machineDeploymentTemplates)).NotTo(Equal(0))

			for _, machineDeploymentTemplate := range machineDeploymentTemplates {
				Expect(machineDeploymentTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.strategy.type", "RollingUpdate"))
			}
		})
	})

	When("VSPHERE_CONTROL_PLANE_CUSTOM_VMX_KEYS and VSPHERE_WORKER_CUSTOM_VMX_KEYS are set", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("Should throw error if the VSPHERE_CONTROL_PLANE_CUSTOM_VMX_KEYS is not in correct format", func() {
			value.Set("VSPHERE_CONTROL_PLANE_CUSTOM_VMX_KEYS", "a,b,c")
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).To(HaveOccurred())
		})

		It("Should throw error if the VSPHERE_WORKER_CUSTOM_VMX_KEYS is not in correct format", func() {
			value.Set("VSPHERE_WORKER_CUSTOM_VMX_KEYS", "a,b,c")
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).To(HaveOccurred())
		})

		It("Should populate CustomVMXKeys if inputs are valid", func() {
			value.Set("VSPHERE_CONTROL_PLANE_CUSTOM_VMX_KEYS", "a=b,c=d")
			value.Set("VSPHERE_WORKER_CUSTOM_VMX_KEYS", "e=f,g=h")

			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			vsphereMachineTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "test-cluster-worker",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(len(vsphereMachineTemplates)).NotTo(Equal(0))

			for _, vsphereMachineTemplate := range vsphereMachineTemplates {
				Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.customVMXKeys.e", "f"))
				Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.customVMXKeys.g", "h"))
			}

			vsphereMachineTemplates, err = matchers.FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "test-cluster-control-plane",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(len(vsphereMachineTemplates)).NotTo(Equal(0))

			for _, vsphereMachineTemplate := range vsphereMachineTemplates {
				Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.customVMXKeys.a", "b"))
				Expect(vsphereMachineTemplate).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.customVMXKeys.c", "d"))
			}
		})
	})
})
