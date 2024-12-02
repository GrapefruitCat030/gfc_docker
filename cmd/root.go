package cmd

import (
	"fmt"
	"os"

	"github.com/docker/docker/pkg/reexec"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gfc_docker",
	Short: "gfc_docker is a simple container runtime implementation",
}

func Init() {
	fmt.Printf("args=%+v\n", os.Args)
	reexec.Register("run-boot", runDetails) // trigger by [gfc_docker run ...]
	if reexec.Init() {
		os.Exit(0)
	}
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
