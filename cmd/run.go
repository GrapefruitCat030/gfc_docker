package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	gfc_cgroup "github.com/GrapefruitCat030/gfc_docker/pkg/cgroup"
	gfc_fs "github.com/GrapefruitCat030/gfc_docker/pkg/fs"
	gfc_subsys "github.com/GrapefruitCat030/gfc_docker/pkg/subsystem"
	gfc_uts "github.com/GrapefruitCat030/gfc_docker/pkg/uts"

	"github.com/docker/docker/pkg/reexec"
	"github.com/spf13/cobra"
)

func init() {
	runCmd.Flags().StringVarP(&runConf.RootFs, "rootfs", "r", "", "root filesystem path")
	runCmd.Flags().StringVarP(&runConf.MemLimit, "memory", "m", "20m", "memory limit")
	runCmd.Flags().BoolVarP(&runConf.Tty, "tty", "t", true, "tty")
	rootCmd.AddCommand(runCmd)
}

type RunConfig struct {
	RootFs   string
	MemLimit string
	Tty      bool
}

var runConf RunConfig

var runCmd = &cobra.Command{
	Use:   "run [echoSomething]",
	Short: "Run a new container(shell)",
	// 	Long: `Run a new container with the specified root filesystem.
	// [rootfs] is a required argument that specifies the path to the root filesystem.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running command: ", args, " with config: ", runConf)
		run()
	},
}

func run() {
	parentProc := reexec.Command("run-boot", runConf.RootFs) // command: /proc/self/exe run-boot [...]
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
	if runConf.Tty {
		parentProc.Stdin = os.Stdin
		parentProc.Stdout = os.Stdout
		parentProc.Stderr = os.Stderr
	}

	if err := parentProc.Start(); err != nil { // new fork self proc => reexec init map => runDetails
		panic(err)
	}

	// run netsetgo using default args
	// gfc_net.SetNetwork(cmd.Process.Pid)

	cgroupManager := gfc_cgroup.NewCgroupManager("gfc_docker")
	defer cgroupManager.Remove()
	cgroupManager.Resource = &gfc_subsys.ResourceConfig{
		MemLimit: runConf.MemLimit,
	}
	cgroupManager.Set()
	cgroupManager.Apply(parentProc.Process.Pid)

	if err := parentProc.Wait(); err != nil {
		fmt.Printf("Error waiting for the reexec.Command - %s\n", err)
		os.Exit(1)
	}
}

func runDetails() {

	// ---- prepare for the new process ----

	newRoot := os.Args[1] // os.Args[1] == paramRootFs
	fmt.Println("Forking run process, newRoot:", newRoot)
	// if err := gfc_fs.PivotRoot(newRoot); err != nil {
	// 	fmt.Printf("PivotRoot Error: %+v\n", err)
	// 	os.Exit(1)
	// }
	if err := gfc_fs.MountProc(""); err != nil {
		fmt.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}
	if err := gfc_uts.AssignHostName(); err != nil {
		fmt.Printf("Error assigning hostname - %s\n", err)
		os.Exit(1)
	}
	// if err := gfc_net.WaitNetwork(); err != nil {
	// 	fmt.Printf("Error waiting for network - %s\n", err)
	// 	os.Exit(1)
	// }

	// ---- execute shell ----

	bootShell()
}

func bootShell() {
	shellPath, err := exec.LookPath("sh")
	if err != nil {
		fmt.Printf("Error finding shell - %s\n", err)
		os.Exit(1)
	}
	env := []string{"PS1=-[gfc_docker]- # "}
	if err := syscall.Exec(shellPath, []string{shellPath}, env); err != nil {
		fmt.Printf("Error executing shell - %s\n", err)
		os.Exit(1)
	}
}
