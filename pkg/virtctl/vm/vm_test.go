package vm_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	"kubevirt.io/kubevirt/tests"

	"k8s.io/client-go/kubernetes/fake"

	cdifake "kubevirt.io/client-go/generated/containerized-data-importer/clientset/versioned/fake"
)

var _ = Describe("VirtualMachine", func() {

	const vmName = "testvm"
	var vmInterface *kubecli.MockVirtualMachineInterface
	var vmiInterface *kubecli.MockVirtualMachineInstanceInterface
	var ctrl *gomock.Controller

	running := true
	notRunning := false
	runStrategyAlways := v1.RunStrategyAlways
	runStrategyManual := v1.RunStrategyManual
	runStrategyHalted := v1.RunStrategyHalted

	startOpts := v1.StartOptions{Paused: false, DryRun: nil}
	stopOpts := v1.StopOptions{DryRun: nil}
	restartOpts := v1.RestartOptions{DryRun: nil}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		kubecli.GetKubevirtClientFromClientConfig = kubecli.GetMockKubevirtClientFromClientConfig
		kubecli.MockKubevirtClientInstance = kubecli.NewMockKubevirtClient(ctrl)
		vmInterface = kubecli.NewMockVirtualMachineInterface(ctrl)
		vmiInterface = kubecli.NewMockVirtualMachineInstanceInterface(ctrl)
	})

	Context("With missing input parameters", func() {
		It("should fail a start", func() {
			cmd := tests.NewRepeatableVirtctlCommand("start")
			Expect(cmd()).NotTo(BeNil())
		})
		It("should fail a stop", func() {
			cmd := tests.NewRepeatableVirtctlCommand("stop")
			Expect(cmd()).NotTo(BeNil())
		})
		It("should fail a restart", func() {
			cmd := tests.NewRepeatableVirtctlCommand("restart")
			Expect(cmd()).NotTo(BeNil())
		})
		It("should fail a migrate", func() {
			cmd := tests.NewRepeatableVirtctlCommand("migrate")
			Expect(cmd()).NotTo(BeNil())
		})
	})

	Context("should patch VM", func() {
		It("with spec:running:true", func() {
			vm := kubecli.NewMinimalVM(vmName)
			vm.Spec.Running = &notRunning

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			vmInterface.EXPECT().Start(vm.Name, &startOpts).Return(nil).Times(1)

			cmd := tests.NewVirtctlCommand("start", vmName)
			Expect(cmd.Execute()).To(BeNil())
		})

		It("with spec:running:false", func() {
			vm := kubecli.NewMinimalVM(vmName)
			vm.Spec.Running = &running

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			vmInterface.EXPECT().Stop(vm.Name, &stopOpts).Return(nil).Times(1)

			cmd := tests.NewVirtctlCommand("stop", vmName)
			Expect(cmd.Execute()).To(BeNil())
		})

		It("with spec:running:false when it's false already ", func() {
			vm := kubecli.NewMinimalVM(vmName)
			vm.Spec.Running = &notRunning

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			vmInterface.EXPECT().Stop(vm.Name, &stopOpts).Return(nil).Times(1)

			cmd := tests.NewVirtctlCommand("stop", vmName)
			Expect(cmd.Execute()).To(BeNil())
		})

		Context("Using RunStrategy", func() {
			It("with spec:runStrategy:running", func() {
				vm := kubecli.NewMinimalVM(vmName)
				vm.Spec.RunStrategy = &runStrategyHalted

				kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
				vmInterface.EXPECT().Start(vm.Name, &startOpts).Return(nil).Times(1)

				cmd := tests.NewVirtctlCommand("start", vmName)
				Expect(cmd.Execute()).To(BeNil())
			})

			It("with spec:runStrategy:halted", func() {
				vm := kubecli.NewMinimalVM(vmName)
				vm.Spec.RunStrategy = &runStrategyAlways

				kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
				vmInterface.EXPECT().Stop(vm.Name, &stopOpts).Return(nil).Times(1)

				cmd := tests.NewVirtctlCommand("stop", vmName)
				Expect(cmd.Execute()).To(BeNil())
			})

			It("with spec:runStrategy:halted when it's false already ", func() {
				vm := kubecli.NewMinimalVM(vmName)
				vm.Spec.RunStrategy = &runStrategyHalted

				kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
				vmInterface.EXPECT().Stop(vm.Name, &stopOpts).Return(nil).Times(1)

				cmd := tests.NewVirtctlCommand("stop", vmName)
				Expect(cmd.Execute()).To(BeNil())
			})
		})
		Context("With --paused flag", func() {
			It("should start paused if --paused true", func() {
				vm := kubecli.NewMinimalVM(vmName)

				kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
				vmInterface.EXPECT().Start(vm.Name, &v1.StartOptions{Paused: true, DryRun: nil}).Return(nil).Times(1)

				cmd := tests.NewVirtctlCommand("start", vmName, "--paused")
				Expect(cmd.Execute()).To(BeNil())
			})
			It("should start if --paused false", func() {
				vm := kubecli.NewMinimalVM(vmName)

				kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
				vmInterface.EXPECT().Start(vm.Name, &startOpts).Return(nil).Times(1)

				cmd := tests.NewVirtctlCommand("start", vmName, "--paused=false")
				Expect(cmd.Execute()).To(BeNil())
			})
		})

	})

	Context("with migrate VM cmd", func() {
		table.DescribeTable("should migrate a vm according to options", func(migrateOptions *v1.MigrateOptions) {
			vm := kubecli.NewMinimalVM(vmName)

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			vmInterface.EXPECT().Migrate(vm.Name, migrateOptions).Return(nil).Times(1)

			var cmd *cobra.Command
			if len(migrateOptions.DryRun) == 0 {
				cmd = tests.NewVirtctlCommand("migrate", vmName)
			} else {
				cmd = tests.NewVirtctlCommand("migrate", "--dry-run", vmName)
			}
			Expect(cmd.Execute()).To(BeNil())
		},
			table.Entry("with default", &v1.MigrateOptions{}),
			table.Entry("with dry-run option", &v1.MigrateOptions{DryRun: []string{k8smetav1.DryRunAll}}),
		)
	})

	Context("with setcpus and setmem VM cmd", func() {
		table.DescribeTable("should setvcpu a vmi according to options", func(vcpus uint32) {
			vm := kubecli.NewMinimalVM(vmName)

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachineInstance(k8smetav1.NamespaceDefault).Return(vmiInterface).Times(1)
			vmiInterface.EXPECT().SetVCpus(vm.Name, vcpus).Return(nil).Times(1)

			var cmd *cobra.Command
			cmd = tests.NewVirtctlCommand("setvcpus", vmName, "--vcpus=16")
			Expect(cmd.Execute()).To(BeNil())
		},
			table.Entry("with default", uint32(16)),
		)

		table.DescribeTable("should setmem a vmi according to options", func(memory string) {
			vm := kubecli.NewMinimalVM(vmName)

			q := resource.MustParse(memory)
			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachineInstance(k8smetav1.NamespaceDefault).Return(vmiInterface).Times(1)
			vmiInterface.EXPECT().SetMemory(vm.Name, &q).Return(nil).Times(1)

			var cmd *cobra.Command
			cmd = tests.NewVirtctlCommand("setmem", vmName, "--mem=16Gi")
			Expect(cmd.Execute()).To(BeNil())
		},
			table.Entry("with default", "16Gi"),
		)
	})

	Context("with restart VM cmd", func() {
		It("should restart vm", func() {
			vm := kubecli.NewMinimalVM(vmName)

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			vmInterface.EXPECT().Restart(vm.Name, &restartOpts).Return(nil).Times(1)

			cmd := tests.NewVirtctlCommand("restart", vmName)
			Expect(cmd.Execute()).To(BeNil())
		})

		Context("using RunStrategy", func() {
			It("should restart a vm with runStrategy:manual", func() {
				vm := kubecli.NewMinimalVM(vmName)
				vm.Spec.RunStrategy = &runStrategyManual

				kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
				vmInterface.EXPECT().Restart(vm.Name, &restartOpts).Return(nil).Times(1)

				cmd := tests.NewVirtctlCommand("restart", vmName)
				Expect(cmd.Execute()).To(BeNil())
			})
		})

		It("should force restart vm", func() {
			vm := kubecli.NewMinimalVM(vmName)

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			gracePeriod := int64(0)
			restartOptions := v1.RestartOptions{
				GracePeriodSeconds: &gracePeriod,
				DryRun:             nil,
			}
			vmInterface.EXPECT().ForceRestart(vm.Name, &restartOptions).Return(nil).Times(1)

			cmd := tests.NewVirtctlCommand("restart", vmName, "--force", "--grace-period=0")
			Expect(cmd.Execute()).To(BeNil())
		})

		It("should force delete vm", func() {
			vm := kubecli.NewMinimalVM(vmName)

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			gracePeriod := int64(0)
			stopOptions := v1.StopOptions{
				GracePeriod: &gracePeriod,
				DryRun:      nil,
			}
			vmInterface.EXPECT().ForceStop(vm.Name, &stopOptions).Return(nil).Times(1)

			cmd := tests.NewVirtctlCommand("stop", vmName, "--force", "--grace-period=0")
			Expect(cmd.Execute()).To(BeNil())
		})
	})

	Context("with dry-run parameter", func() {

		It("should not start VM", func() {
			vm := kubecli.NewMinimalVM(vmName)
			vm.Spec.Running = &notRunning

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			vmInterface.EXPECT().Start(vm.Name, &v1.StartOptions{DryRun: []string{k8smetav1.DryRunAll}}).Return(nil).Times(1)

			cmd := tests.NewVirtctlCommand("start", vmName, "--dry-run")
			Expect(cmd.Execute()).To(BeNil())
		})

		It("should not restart VM", func() {
			vm := kubecli.NewMinimalVM(vmName)
			vm.Spec.Running = &running

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			vmInterface.EXPECT().Restart(vm.Name, &v1.RestartOptions{DryRun: []string{k8smetav1.DryRunAll}}).Return(nil).Times(1)

			cmd := tests.NewVirtctlCommand("restart", vmName, "--dry-run")
			Expect(cmd.Execute()).To(BeNil())
		})

		It("should not stop VM", func() {
			vm := kubecli.NewMinimalVM(vmName)
			vm.Spec.Running = &running

			kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
			vmInterface.EXPECT().Stop(vm.Name, &v1.StopOptions{DryRun: []string{k8smetav1.DryRunAll}}).Return(nil).Times(1)

			cmd := tests.NewVirtctlCommand("stop", vmName, "--dry-run")
			Expect(cmd.Execute()).To(BeNil())
		})

	})

	Context("guest agent", func() {

		It("should return guest agent data", func() {
			vm := kubecli.NewMinimalVM(vmName)
			guestOSInfo := v1.VirtualMachineInstanceGuestAgentInfo{
				GAVersion: "3.1.0",
			}

			kubecli.MockKubevirtClientInstance.
				EXPECT().
				VirtualMachineInstance(k8smetav1.NamespaceDefault).
				Return(vmiInterface).
				Times(1)

			vmiInterface.EXPECT().GuestOsInfo(vm.Name).Return(guestOSInfo, nil).Times(1)

			cmd := tests.NewVirtctlCommand("guestosinfo", vm.Name)
			Expect(cmd.Execute()).To(BeNil())
		})

		It("should return filesystem  data", func() {
			vm := kubecli.NewMinimalVM(vmName)
			fsList := v1.VirtualMachineInstanceFileSystemList{
				Items: []v1.VirtualMachineInstanceFileSystem{
					{
						DiskName: "TEST",
					},
				},
			}

			kubecli.MockKubevirtClientInstance.
				EXPECT().
				VirtualMachineInstance(k8smetav1.NamespaceDefault).
				Return(vmiInterface).
				Times(1)

			vmiInterface.EXPECT().FilesystemList(vm.Name).Return(fsList, nil).Times(1)

			cmd := tests.NewVirtctlCommand("fslist", vm.Name)
			Expect(cmd.Execute()).To(BeNil())
		})

		It("should return userlist  data", func() {
			vm := kubecli.NewMinimalVM(vmName)
			userList := v1.VirtualMachineInstanceGuestOSUserList{
				Items: []v1.VirtualMachineInstanceGuestOSUser{
					{
						UserName: "TEST",
					},
				},
			}

			kubecli.MockKubevirtClientInstance.
				EXPECT().
				VirtualMachineInstance(k8smetav1.NamespaceDefault).
				Return(vmiInterface).
				Times(1)

			vmiInterface.EXPECT().UserList(vm.Name).Return(userList, nil).Times(1)

			cmd := tests.NewVirtctlCommand("userlist", vm.Name)
			Expect(cmd.Execute()).To(BeNil())
		})
	})

	Context("hotplug volume", func() {
		var (
			cdiClient  *cdifake.Clientset
			coreClient *fake.Clientset
		)
		createTestDataVolume := func() *v1beta1.DataVolume {
			return &v1beta1.DataVolume{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name: "testvolume",
				},
			}
		}

		createTestPVC := func() *corev1.PersistentVolumeClaim {
			return &corev1.PersistentVolumeClaim{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name: "testvolume",
				},
			}
		}

		verifyVolumeSource := func(volumeOptions interface{}, useDv bool) {
			if useDv {
				Expect(volumeOptions.(*v1.AddVolumeOptions).VolumeSource.DataVolume).ToNot(BeNil())
				Expect(volumeOptions.(*v1.AddVolumeOptions).VolumeSource.PersistentVolumeClaim).To(BeNil())
			} else {
				Expect(volumeOptions.(*v1.AddVolumeOptions).VolumeSource.DataVolume).To(BeNil())
				Expect(volumeOptions.(*v1.AddVolumeOptions).VolumeSource.PersistentVolumeClaim).ToNot(BeNil())
			}
		}

		expectVMIEndpointAddVolume := func(vmiName, volumeName string, useDv bool) {
			kubecli.MockKubevirtClientInstance.
				EXPECT().
				VirtualMachineInstance(k8smetav1.NamespaceDefault).
				Return(vmiInterface).
				Times(1)
			vmiInterface.EXPECT().AddVolume(vmiName, gomock.Any()).DoAndReturn(func(arg0, arg1 interface{}) interface{} {
				Expect(arg1.(*v1.AddVolumeOptions).Name).To(Equal(volumeName))
				verifyVolumeSource(arg1, useDv)
				return nil
			})
		}

		expectVMIEndpointRemoveVolume := func(vmiName, volumeName string, useDv bool) {
			kubecli.MockKubevirtClientInstance.
				EXPECT().
				VirtualMachineInstance(k8smetav1.NamespaceDefault).
				Return(vmiInterface).
				Times(1)
			vmiInterface.EXPECT().RemoveVolume(vmiName, gomock.Any()).DoAndReturn(func(arg0, arg1 interface{}) interface{} {
				Expect(arg1.(*v1.RemoveVolumeOptions).Name).To(Equal(volumeName))
				return nil
			})
		}

		expectVMIEndpointRemoveVolumeError := func(vmiName, volumeName string) {
			kubecli.MockKubevirtClientInstance.
				EXPECT().
				VirtualMachineInstance(k8smetav1.NamespaceDefault).
				Return(vmiInterface).
				Times(1)
			vmiInterface.EXPECT().RemoveVolume(vmiName, gomock.Any()).DoAndReturn(func(arg0, arg1 interface{}) interface{} {
				Expect(arg1.(*v1.RemoveVolumeOptions).Name).To(Equal(volumeName))
				return fmt.Errorf("error removing")
			})
		}

		expectVMEndpointAddVolume := func(vmiName, volumeName string, useDv bool) {
			kubecli.MockKubevirtClientInstance.
				EXPECT().
				VirtualMachine(k8smetav1.NamespaceDefault).
				Return(vmInterface).
				Times(1)
			vmInterface.EXPECT().AddVolume(vmiName, gomock.Any()).DoAndReturn(func(arg0, arg1 interface{}) interface{} {
				Expect(arg1.(*v1.AddVolumeOptions).Name).To(Equal(volumeName))
				verifyVolumeSource(arg1, useDv)
				return nil
			})
		}

		expectVMEndpointRemoveVolume := func(vmiName, volumeName string, useDv bool) {
			kubecli.MockKubevirtClientInstance.
				EXPECT().
				VirtualMachine(k8smetav1.NamespaceDefault).
				Return(vmInterface).
				Times(1)
			vmInterface.EXPECT().RemoveVolume(vmiName, gomock.Any()).DoAndReturn(func(arg0, arg1 interface{}) interface{} {
				Expect(arg1.(*v1.RemoveVolumeOptions).Name).To(Equal(volumeName))
				return nil
			})
		}

		BeforeEach(func() {
			cdiClient = cdifake.NewSimpleClientset()
			coreClient = fake.NewSimpleClientset()
		})

		table.DescribeTable("should fail with missing required or invalid parameters", func(commandName, errorString string, args ...string) {
			commandAndArgs := []string{commandName}
			commandAndArgs = append(commandAndArgs, args...)
			cmdAdd := tests.NewRepeatableVirtctlCommand(commandAndArgs...)
			res := cmdAdd()
			Expect(res).NotTo(BeNil())
			Expect(res.Error()).To(ContainSubstring(errorString))
		},
			table.Entry("addvolume no args", "addvolume", "argument validation failed"),
			table.Entry("addvolume name, missing required volume-name", "addvolume", "required flag(s)", "testvmi"),
			table.Entry("addvolume name, invalid extra parameter", "addvolume", "unknown flag", "testvmi", "--volume-name=blah", "--invalid=test"),
			table.Entry("removevolume no args", "removevolume", "argument validation failed"),
			table.Entry("removevolume name, missing required volume-name", "removevolume", "required flag(s)", "testvmi"),
			table.Entry("removevolume name, invalid extra parameter", "removevolume", "unknown flag", "testvmi", "--volume-name=blah", "--invalid=test"),
		)
		table.DescribeTable("should fail addvolume when no source is found according to option", func(isDryRun bool) {
			kubecli.MockKubevirtClientInstance.EXPECT().CdiClient().Return(cdiClient)
			kubecli.MockKubevirtClientInstance.EXPECT().CoreV1().Return(coreClient.CoreV1())
			commandAndArgs := []string{"addvolume", "testvmi", "--volume-name=testvolume"}
			if isDryRun {
				commandAndArgs = append(commandAndArgs, "--dry-run")
			}
			cmdAdd := tests.NewRepeatableVirtctlCommand(commandAndArgs...)
			res := cmdAdd()
			Expect(res).ToNot(BeNil())
			Expect(res.Error()).To(ContainSubstring("Volume testvolume is not a DataVolume or PersistentVolumeClaim"))
		},
			table.Entry("with default", false),
			table.Entry("with dry-run arg", true),
		)

		table.DescribeTable("should call correct endpoint", func(commandName, vmiName, volumeName string, useDv bool, expectFunc func(vmiName, volumeName string, useDv bool), args ...string) {
			if commandName == "addvolume" {
				kubecli.MockKubevirtClientInstance.EXPECT().CdiClient().Return(cdiClient)
				if useDv {
					cdiClient.CdiV1beta1().DataVolumes(k8smetav1.NamespaceDefault).Create(context.Background(), createTestDataVolume(), k8smetav1.CreateOptions{})
				} else {
					kubecli.MockKubevirtClientInstance.EXPECT().CoreV1().Return(coreClient.CoreV1())
					coreClient.CoreV1().PersistentVolumeClaims(k8smetav1.NamespaceDefault).Create(context.Background(), createTestPVC(), k8smetav1.CreateOptions{})
				}
			}
			expectFunc(vmiName, volumeName, useDv)
			commandAndArgs := []string{commandName, vmiName, fmt.Sprintf("--volume-name=%s", volumeName)}
			commandAndArgs = append(commandAndArgs, args...)
			cmd := tests.NewVirtctlCommand(commandAndArgs...)
			Expect(cmd.Execute()).To(BeNil())
		},
			table.Entry("addvolume dv, no persist should call VMI endpoint", "addvolume", "testvmi", "testvolume", true, expectVMIEndpointAddVolume),
			table.Entry("addvolume pvc, no persist should call VMI endpoint", "addvolume", "testvmi", "testvolume", false, expectVMIEndpointAddVolume),
			table.Entry("addvolume dv, with persist should call VM endpoint", "addvolume", "testvmi", "testvolume", true, expectVMEndpointAddVolume, "--persist"),
			table.Entry("addvolume pvc, with persist should call VM endpoint", "addvolume", "testvmi", "testvolume", false, expectVMEndpointAddVolume, "--persist"),
			table.Entry("removevolume dv, no persist should call VMI endpoint", "removevolume", "testvmi", "testvolume", true, expectVMIEndpointRemoveVolume),
			table.Entry("removevolume pvc, no persist should call VMI endpoint", "removevolume", "testvmi", "testvolume", false, expectVMIEndpointRemoveVolume),
			table.Entry("removevolume dv, with persist should call VM endpoint", "removevolume", "testvmi", "testvolume", true, expectVMEndpointRemoveVolume, "--persist"),
			table.Entry("removevolume pvc, with persist should call VM endpoint", "removevolume", "testvmi", "testvolume", false, expectVMEndpointRemoveVolume, "--persist"),

			table.Entry("addvolume dv, no persist with dry-run should call VMI endpoint", "addvolume", "testvmi", "testvolume", true, expectVMIEndpointAddVolume, "--dry-run"),
			table.Entry("addvolume pvc, no persist with dry-run should call VMI endpoint", "addvolume", "testvmi", "testvolume", false, expectVMIEndpointAddVolume, "--dry-run"),
			table.Entry("addvolume dv, with persist with dry-run should call VM endpoint", "addvolume", "testvmi", "testvolume", true, expectVMEndpointAddVolume, "--persist", "--dry-run"),
			table.Entry("addvolume pvc, with persist with dry-run should call VM endpoint", "addvolume", "testvmi", "testvolume", false, expectVMEndpointAddVolume, "--persist", "--dry-run"),
			table.Entry("removevolume dv, no persist with dry-run should call VMI endpoint", "removevolume", "testvmi", "testvolume", true, expectVMIEndpointRemoveVolume, "--dry-run"),
			table.Entry("removevolume pvc, no persist with dry-run should call VMI endpoint", "removevolume", "testvmi", "testvolume", false, expectVMIEndpointRemoveVolume, "--dry-run"),
			table.Entry("removevolume dv, with persist with dry-run should call VM endpoint", "removevolume", "testvmi", "testvolume", true, expectVMEndpointRemoveVolume, "--persist", "--dry-run"),
			table.Entry("removevolume pvc, with persist with dry-run should call VM endpoint", "removevolume", "testvmi", "testvolume", false, expectVMEndpointRemoveVolume, "--persist", "--dry-run"),
		)

		table.DescribeTable("removevolume should report error if call returns error according to option", func(isDryRun bool) {
			expectVMIEndpointRemoveVolumeError("testvmi", "testvolume")
			commandAndArgs := []string{"removevolume", "testvmi", "--volume-name=testvolume"}
			if isDryRun {
				commandAndArgs = append(commandAndArgs, "--dry-run")
			}
			cmdAdd := tests.NewRepeatableVirtctlCommand(commandAndArgs...)
			res := cmdAdd()
			Expect(res).ToNot(BeNil())
			Expect(res.Error()).To(ContainSubstring("error removing"))
		},
			table.Entry("with default", false),
			table.Entry("with dry-run arg", true),
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})
})
