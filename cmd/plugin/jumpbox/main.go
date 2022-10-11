package main

import (
	"context"
	"fmt"
	"github.com/aunum/log"
	"github.com/spf13/cobra"
	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"
)

var (
	c             kubernetes.Interface
	dynamicClient dynamic.Interface
	ctx           context.Context

	createCmd   *cobra.Command
	sshCmd      *cobra.Command
	powerOnCmd  *cobra.Command
	powerOffCmd *cobra.Command
	destroyCmd  *cobra.Command
)

var vmOptions = &VMOptions{}

var pluginDescriptor = cliv1alpha1.PluginDescriptor{
	Name:        "jumpbox",
	Description: "tanzu cli plugin for jumpox management (tanzu vm service)",
	Version:     "v0.0.1",
	Group:       cliv1alpha1.ManageCmdGroup,
}

func init() {
	ctx = context.Background()
	newPowerOnCmd(ctx)
	newPowerOffCmd(ctx)
	newCreateCmd(ctx)
	newSshCmd(ctx)
	newDestroyCmd(ctx)

}

func main() {
	p, err := plugin.NewPlugin(&pluginDescriptor)
	if err != nil {
		log.Fatal(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", "/Users/vpupim/.kube/config")
	if err != nil {
		log.Fatal(err)
	}

	c, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("error creating dynamic client: %v\n", err)
		os.Exit(1)
	}

	p.AddCommands(
		createCmd,
		sshCmd,
		powerOnCmd,
		powerOffCmd,
		destroyCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func newCreateCmd(ctx context.Context) *cobra.Command {
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create Jumpbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createJumpBox(ctx, parseArgs(args))
		}}

	createCmd.Flags().StringVarP(&vmOptions.Namespace, "namespace", "n", "", "vm namespace")
	createCmd.Flags().StringVarP(&vmOptions.StorageClassName, "storage-class", "", "", "vm storage class name")
	createCmd.Flags().StringVarP(&vmOptions.ImageName, "image", "i", "", "vm image from VM Service registered content library")
	createCmd.Flags().StringVarP(&vmOptions.ClassName, "class", "c", "", "vm class")
	createCmd.Flags().StringVarP(&vmOptions.NetworkType, "network-type", "", "", "Network type. `nsx-t` or `vsphere-distributed`")
	createCmd.Flags().StringVarP(&vmOptions.NetworkName, "network-name", "", "", "Network name. required if network-type = `vsphere-distributed`")
	createCmd.Flags().StringVarP(&vmOptions.SshPubPath, "ssh-pub", "", "$HOME/.ssh/id_rsa.pub", "Path to the ssh public key to include in VM authorized_keys")
	createCmd.Flags().StringVarP(&vmOptions.User, "user", "u", "operator", "User to be created in VM")
	createCmd.Flags().StringVarP(&vmOptions.Password, "password", "p", "VMware1!", "User's password for VM login")
	createCmd.Flags().BoolVarP(&vmOptions.WaitCreate, "wait", "w", true, "Wait for VM to be created")

	createCmd.MarkFlagRequired("storage-class")
	createCmd.MarkFlagRequired("image")
	createCmd.MarkFlagRequired("class")
	createCmd.MarkFlagRequired("network-type")

	return createCmd
}

func newSshCmd(ctx context.Context) *cobra.Command {
	sshCmd = &cobra.Command{
		Use:   "ssh",
		Short: "ssh Jumpbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sshJumpbox(ctx, parseArgs(args))
		}}
	sshCmd.Flags().StringVarP(&vmOptions.Namespace, "namespace", "n", "", "vm namespace")
	sshCmd.Flags().StringVarP(&vmOptions.SshKeyPath, "ssh-key", "i", "$HOME/.ssh/id_rsaa", "Path to the ssh private key to access the vm")

	return sshCmd
}

func newPowerOnCmd(ctx context.Context) *cobra.Command {
	powerOnCmd = &cobra.Command{
		Use:   "power-on",
		Short: "Power On VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return powerOn(ctx, parseArgs(args))
		}}
	powerOnCmd.Flags().StringVarP(&vmOptions.Namespace, "namespace", "n", "", "vm namespace")
	return powerOnCmd
}
func newPowerOffCmd(ctx context.Context) *cobra.Command {
	powerOffCmd = &cobra.Command{
		Use:   "power-off",
		Short: "Power Off VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return powerOff(ctx, parseArgs(args))
		}}
	powerOffCmd.Flags().StringVarP(&vmOptions.Namespace, "namespace", "n", "", "vm namespace")
	return powerOffCmd
}

func newDestroyCmd(ctx context.Context) *cobra.Command {
	destroyCmd = &cobra.Command{
		Use:   "destroy",
		Short: "Destroy VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return destroy(ctx, parseArgs(args))
		}}
	destroyCmd.Flags().StringVarP(&vmOptions.Namespace, "namespace", "n", "", "vm namespace")
	return destroyCmd
}

func parseArgs(args []string) *VMOptions {
	vmName := args[0]
	vmOptions.Name = vmName
	vmOptions.pvcName = vmName + "-pvc"
	vmOptions.configName = vmName + "-cm"
	vmOptions.svcName = vmName + "-svc"
	if vmOptions.SshKeyPath == "" {
		vmOptions.SshKeyPath = strings.Split(vmOptions.SshPubPath, ".")[0]
	}
	//fmt.Printf("vmoptions %+v\n", vmOptions)
	return vmOptions

}
