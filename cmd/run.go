package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	gfc_cgroup "github.com/GrapefruitCat030/gfc_docker/pkg/cgroup"
	gfc_fs "github.com/GrapefruitCat030/gfc_docker/pkg/fs"
	gfc_pipe "github.com/GrapefruitCat030/gfc_docker/pkg/pipe"
	gfc_subsys "github.com/GrapefruitCat030/gfc_docker/pkg/subsystem"
	gfc_uts "github.com/GrapefruitCat030/gfc_docker/pkg/uts"

	"github.com/docker/docker/pkg/reexec"
	"github.com/spf13/cobra"
)

func init() {
	runCmd.Flags().StringVarP(&runConf.RootFs, "rootfs", "r", "filesystem/busyboxfs", "root filesystem path")
	runCmd.Flags().StringVarP(&runConf.MemLimit, "memory", "m", "20m", "memory limit")
	runCmd.Flags().BoolVarP(&runConf.Tty, "tty", "t", false, "tty")
	rootCmd.AddCommand(runCmd)
}

type RunConfig struct {
	RootFs   string
	MemLimit string
	Tty      bool
}

var runConf RunConfig

var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Run a new container",
	// 	Long: `Run a new container with the specified root filesystem.
	// [rootfs] is a required argument that specifies the path to the root filesystem.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running command: ", args, " with config: ", runConf)
		fmt.Println("Running command len", len(args))
		run(args)
	},
}

func run(args []string) {

	// ---- fork self process ----
	parentProc, pw := runNewProcess() // new fork proc/self/exe  => reexec init map => func runDetails()
	if err := parentProc.Start(); err != nil {
		panic(err)
	}

	// ---- run netsetgo using default setting ----
	// gfc_net.SetNetwork(cmd.Process.Pid)

	// ---- set cgroup & subsys ----
	cgroupManager := gfc_cgroup.NewCgroupManager("gfc_docker")
	defer cgroupManager.Remove()
	cgroupManager.Resource = &gfc_subsys.ResourceConfig{MemLimit: runConf.MemLimit}
	cgroupManager.Set()
	cgroupManager.Apply(parentProc.Process.Pid)

	// ---- write user command to pipe ----
	// Keep this behavior at the end to synchronize the parent and child processes
	if _, err := pw.WriteString(strings.Join(args, " ")); err != nil {
		fmt.Printf("Error writing to pipe - %s\n", err)
		os.Exit(1)
	}
	pw.Close()

	if err := parentProc.Wait(); err != nil {
		fmt.Printf("Error waiting for the reexec.Command - %s\n", err)
		os.Exit(1)
	}
}

func runNewProcess() (*exec.Cmd, *os.File) {
	// ---- create pipe ----
	pr, pw, err := gfc_pipe.NewPipe()
	if err != nil {
		fmt.Printf("Error creating pipe - %s\n", err)
		os.Exit(1)
	}

	// ---- fork self process ----
	parentProc := reexec.Command("run-boot") // command: /proc/self/exe run-boot [...]
	parentProc.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID:      os.Getuid(),
			Size:        1,
		}},
		GidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID:      os.Getgid(),
			Size:        1,
		}},
	}
	parentProc.ExtraFiles = []*os.File{pr}
	parentProc.Dir = filepath.Join(os.Getenv("PWD"), runConf.RootFs)
	if runConf.Tty {
		parentProc.Stdin = os.Stdin
		parentProc.Stdout = os.Stdout
		parentProc.Stderr = os.Stderr
	}
	return parentProc, pw
}

func runDetails() {

	// ---- read user command from pipe ----
	cmdArr := readUserCmd()

	// ---- setup new root filesystem ----
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory - %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Current location: ", pwd)

	if err := gfc_fs.PivotRoot(pwd); err != nil {
		fmt.Printf("PivotRoot Error: %+v\n", err)
		os.Exit(1)
	}

	// ---- mount proc ----
	if err := gfc_fs.MountProc(); err != nil {
		fmt.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}

	// ---- mount tmpfs ----
	if err := gfc_fs.MountTmpfs(); err != nil {
		fmt.Printf("Error mounting /tmp - %s\n", err)
		os.Exit(1)
	}

	// ---- setup new hostname ----
	if err := gfc_uts.AssignHostName(); err != nil {
		fmt.Printf("Error assigning hostname - %s\n", err)
		os.Exit(1)
	}

	// if err := gfc_net.WaitNetwork(); err != nil {
	// 	fmt.Printf("Error waiting for network - %s\n", err)
	// 	os.Exit(1)
	// }

	// ---- execute command ----
	execCommand(cmdArr)
}

// ---- helper functions ----

func execCommand(command []string) {
	if len(command) == 0 {
		fmt.Println("Command is empty")
		os.Exit(1)
	}
	cmdPath, err := exec.LookPath(command[0])
	if err != nil {
		fmt.Printf("Error finding shell - %s\n", err)
		os.Exit(1)
	}
	env := []string{"PS1=-[gfc_docker]- # "}
	if err := syscall.Exec(cmdPath, command, env); err != nil {
		fmt.Printf("Error executing shell - %s\n", err)
		os.Exit(1)
	}
}

func readUserCmd() []string {
	pr := os.NewFile(3, "pipe")
	defer pr.Close()

	msg, err := io.ReadAll(pr)
	if err != nil {
		fmt.Printf("Error reading from pipe - %s\n", err)
		os.Exit(1)
	}
	return strings.Fields(string(msg))
}
