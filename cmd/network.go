package cmd

import (
	gfc_net "github.com/GrapefruitCat030/gfc_docker/pkg/net"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(networkCmd)
	networkCmd.AddCommand(networkCreateCmd)
	networkCreateCmd.Flags().StringVarP(&networkConfig.Subnet, "subnet", "s", "", "subnet cidr")
	networkCreateCmd.Flags().StringVarP(&networkConfig.Driver, "driver", "d", "", "network driver")
	networkCreateCmd.MarkFlagRequired("subnet")
	networkCreateCmd.MarkFlagRequired("driver")
}

type NetworkConfig struct {
	Subnet string
	Driver string
}

var networkConfig NetworkConfig

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage networks",
}

var networkCreateCmd = &cobra.Command{
	Use:   "create [network name]",
	Short: "Create a network, with a specified subnet and driver, and a network name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return gfc_net.CreateNetwork(networkConfig.Driver, networkConfig.Subnet, name)
	},
}

var networkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List networks",
	Run: func(cmd *cobra.Command, args []string) {
		gfc_net.ListNetworks()
	},
}

var networkRemoveCmd = &cobra.Command{
	Use:   "remove [network name]",
	Short: "Remove a network",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := gfc_net.RemoveNetwork(name); err != nil {
			panic(err)
		}
	},
}
