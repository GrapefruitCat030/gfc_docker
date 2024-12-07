package cmd

import (
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Flags().StringVarP(&commitConf.containerRootPath, "dir", "d", "/root/project/gfc_docker/filesystem/merged", "container root path")
	rootCmd.AddCommand(commitCmd)
}

type CommitConfig struct {
	containerRootPath string
}

var commitConf CommitConfig

var commitCmd = &cobra.Command{
	Use:   "commit [container-name]",
	Short: "Commit a container into an image",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commit(args)
	},
}

func commit(args []string) {
	containerName := args[0]
	imageTar := containerName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", commitConf.containerRootPath, ".").CombinedOutput(); err != nil {
		panic(err)
	}
}
