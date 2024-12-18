package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"
	gfc_ufs "github.com/GrapefruitCat030/gfc_docker/pkg/ufs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop [container_name]",
	Short: "Stop a running container",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stop(args[0])
	},
}

func stop(container_name string) {
	cinfo, err := gfc_runinfo.GetContainerInfo(container_name)
	if err != nil {
		fmt.Printf("Error getting container info - %s\n", err)
		os.Exit(1)
	}
	if cinfo.Status != gfc_runinfo.STATUS_RUNNING {
		fmt.Printf("Container is not running\n")
		os.Exit(1)
	}

	pid := cinfo.Pid
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return
	}
	proc, err := os.FindProcess(pidInt)
	if err != nil {
		return
	}
	proc.Signal(os.Interrupt)

	if err := gfc_runinfo.DeleteContainerInfo(container_name); err != nil { // TODO: if detach container over?
		fmt.Printf("Error deleting container info - %s\n", err)
		os.Exit(1)
	}
	if err := gfc_ufs.DeleteWorkSpace(runConf.RootFs, runConf.Volume, runConf.UFSer); err != nil {
		fmt.Printf("Error deleting union filesystem - %s\n", err)
		os.Exit(1)
	}

	cinfo.Status = gfc_runinfo.STATUS_STOPPED
	cinfo.Pid = ""
	jsonBytes, err := json.Marshal(cinfo)
	if err != nil {
		fmt.Printf("Error marshalling container info - %s\n", err)
		os.Exit(1)
	}
	path := filepath.Join(gfc_runinfo.DefaultInfoLocation, container_name, gfc_runinfo.ConfigName)
	if err := os.WriteFile(path, jsonBytes, 0777); err != nil {
		fmt.Printf("Error writing container info - %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Container stopped")
}
