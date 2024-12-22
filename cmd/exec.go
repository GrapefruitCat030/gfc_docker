package cmd

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/GrapefruitCat030/gfc_docker/pkg/cgo/nsenter"
	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"

	"github.com/docker/docker/pkg/reexec"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(execCmd)
}

const (
	ENV_EXEC_PID = "MYDOCKER_PID"
	ENV_EXEC_CMD = "MYDOCKER_CMD"
)

var execCmd = &cobra.Command{
	Use:   "exec [container_name] [command]",
	Short: "Execute a command in a running container",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Execute command: ", args)
		execute(args[0], args[1:])
	},
}

func execute(container_name string, command []string) {
	pid := gfc_runinfo.GetContainerPid(container_name)
	if pid == "" {
		fmt.Println("Container not found")
		return
	}

	cmdStr := strings.Join(command, " ")

	parentProc := reexec.Command("exec-boot")
	parentProc.Stdin = os.Stdin
	parentProc.Stdout = os.Stdout
	parentProc.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)
	envs, err := gfc_runinfo.GetContainerEnv(pid)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	parentProc.Env = append(os.Environ(), envs...)

	if err := parentProc.Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}

func execDetails() {
	// DO NOTHING, cmd call implement in cgo
}
