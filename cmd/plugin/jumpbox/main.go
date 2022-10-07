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
)

var (
	c             kubernetes.Interface
	dynamicClient dynamic.Interface

	createCmd  *cobra.Command
	powerOnCmd *cobra.Command
)

func init() {

}

var pluginDescriptor = cliv1alpha1.PluginDescriptor{
	Name:        "jumpbox",
	Description: "tanzu cli plugin for jumpox management (tanzu vm service)",
	Version:     "v0.0.1",
	Group:       cliv1alpha1.ManageCmdGroup, // set group
}

func main() {
	ctx := context.Background()

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
		newPowerOnCmd(ctx),
		newCreateCmd(ctx),
		newSshCmd(ctx),
		// Add commands.go
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func newPowerOnCmd(ctx context.Context) *cobra.Command {
	powerOnCmd = &cobra.Command{
		Use:   "power-on",
		Short: "Power On VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return powerOn(ctx, parseArgs(args))
		}}
	return powerOnCmd
}

func newCreateCmd(ctx context.Context) *cobra.Command {
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create Jumpbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createJumpBox(ctx, parseArgs(args))
		}}
	return createCmd
}

func newSshCmd(ctx context.Context) *cobra.Command {
	createCmd = &cobra.Command{
		Use:   "ssh",
		Short: "ssh Jumpbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sshJumpbox(ctx, parseSshArgs(args))
		}}
	return createCmd
}

func parseArgs(args []string) *VMOptions {
	vmOptions := &VMOptions{
		Name:             "jumpbox-2",
		Namespace:        "vms",
		StorageClassName: "vc01cl01-t0compute",
	}
	vmOptions.PVCName = vmOptions.Name + "-pvc"
	vmOptions.ConfigName = vmOptions.Name + "-cm"
	return vmOptions

}

func parseSshArgs(args []string) *sshOptions {
	return &sshOptions{
		vmName:     "jumpbox-test",
		namespace:  "vms",
		sshKeyPath: "/Users/vpupim/.ssh/id_rsa",
	}
}
