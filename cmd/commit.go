package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"

	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(commitCmd)
}

type CommitConfig struct {
	containerName string
	imageName     string
}

var commitConf CommitConfig

var commitCmd = &cobra.Command{
	Use:   "commit [container-name] [image-name]",
	Short: "Commit a container into an image",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		commit(args)
	},
}

func commit(args []string) {
	containerName := args[0]
	imageName := args[1]
	imageTar := imageName + ".tar"
	cinfo, err := gfc_runinfo.GetContainerInfo(containerName)
	if err != nil {
		fmt.Println("Error getting container info - ", err)
		return
	}
	path := filepath.Join(cinfo.RootFs, runConf.UFSer.WorkSpace())
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", path, ".").CombinedOutput(); err != nil {
		fmt.Println("Error creating image tar - ", err)
		return
	}
}
