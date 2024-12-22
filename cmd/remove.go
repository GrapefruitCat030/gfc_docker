package cmd

import (
	"fmt"
	"os"

	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"
	gfc_ufs "github.com/GrapefruitCat030/gfc_docker/pkg/ufs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "rm [container_name]",
	Short: "Remove a stopped container",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remove(args[0])
	},
}

func remove(container_name string) {
	cinfo, err := gfc_runinfo.GetContainerInfo(container_name)
	if err != nil {
		fmt.Printf("Error getting container info - %s\n", err)
		os.Exit(1)
	}

	if cinfo.Status != gfc_runinfo.STATUS_STOPPED {
		fmt.Printf("Container is not stopped\n")
		os.Exit(1)
	}

	if err := gfc_runinfo.DeleteContainerInfo(cinfo.Name); err != nil { // TODO: if detach container over?
		fmt.Printf("Error deleting container info - %s\n", err)
		os.Exit(1)
	}
	if err := gfc_ufs.DeleteWorkSpace(cinfo.RootFs, cinfo.Volume, runConf.UFSer); err != nil {
		fmt.Printf("Error deleting union filesystem - %s\n", err)
		os.Exit(1)
	}
}
