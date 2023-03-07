package main

import (
	"context"
	"fmt"
	"github.com/aunum/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

var (
	c             kubernetes.Interface
	dynamicClient dynamic.Interface
)

var pluginDescriptor = cliv1alpha1.PluginDescriptor{
	Name:        "jumpbox",
	Description: "tanzu cli plugin for jumpox management (tanzu vm service)",
	Version:     "v1.0.0",
	Group:       cliv1alpha1.ManageCmdGroup,
}

var options = &VMOptions{}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(errors.Wrap(err, "error getting user home dir"))
	}
	options.tanzuDir = filepath.Join(homeDir, ".tanzu", "jumpbox")
}
func main() {
	ctx := context.Background()

	p, err := plugin.NewPlugin(&pluginDescriptor)
	if err != nil {
		log.Fatal(err)
	}
	// todo fix kubeconfig path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homeDir, ".kube/config"))
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
		newCreateCmd(ctx),
		newSshCmd(ctx),
		newPowerOnCmd(ctx),
		newPowerOffCmd(ctx),
		newDestroyCmd(ctx),
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func newCreateCmd(ctx context.Context) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create Jumpbox",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			setup(args)
			return buildUserdata()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return CreateJumpBox(ctx)
		}}

	createCmd.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "vm namespace")
	createCmd.Flags().StringVarP(&options.StorageClassName, "storage-class", "", "", "vm storage class name")
	createCmd.Flags().StringVarP(&options.ImageName, "image", "i", "", "vm image from VM Service registered content library")
	createCmd.Flags().StringVarP(&options.ClassName, "class", "c", "", "vm class")
	createCmd.Flags().StringVarP(&options.NetworkType, "network-type", "", "", "Network type. `nsx-t` or `vsphere-distributed`")
	createCmd.Flags().StringVarP(&options.NetworkName, "network-name", "", "", "Network name. required if network-type = `vsphere-distributed`")
	createCmd.Flags().StringVarP(&options.User, "user", "u", "", "User to be created in VM")

	createCmd.MarkFlagRequired("namespace")
	createCmd.MarkFlagRequired("storage-class")
	createCmd.MarkFlagRequired("image")
	createCmd.MarkFlagRequired("class")
	createCmd.MarkFlagRequired("network-type")

	return createCmd
}

func newSshCmd(ctx context.Context) *cobra.Command {
	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "ssh Jumpbox",
		PreRun: func(cmd *cobra.Command, args []string) {
			setup(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Ssh(ctx)
		}}
	sshCmd.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "vm namespace")
	sshCmd.Flags().StringVarP(&options.sshPrivateKeyPath, "ssh-key", "i", "", "Path to the ssh private key to access the vm")
	sshCmd.Flags().StringVarP(&options.User, "user", "u", "", "user to access the vm")

	return sshCmd
}

func newPowerOnCmd(ctx context.Context) *cobra.Command {
	powerOnCmd := &cobra.Command{
		Use:   "power-on",
		Short: "Power On VM",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			setup(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return PowerOn(ctx)
		}}
	powerOnCmd.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "vm namespace")
	powerOnCmd.MarkFlagRequired("namespace")

	return powerOnCmd
}

func newPowerOffCmd(ctx context.Context) *cobra.Command {
	powerOffCmd := &cobra.Command{
		Use:   "power-off",
		Short: "Power Off VM",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			setup(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return PowerOff(ctx)
		}}
	powerOffCmd.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "vm namespace")
	powerOffCmd.MarkFlagRequired("namespace")

	return powerOffCmd
}

func newDestroyCmd(ctx context.Context) *cobra.Command {
	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy VM",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			setup(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Destroy(ctx)
		}}
	destroyCmd.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "vm namespace")
	destroyCmd.MarkFlagRequired("namespace")

	return destroyCmd
}
