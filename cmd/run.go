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
	gfc_net "github.com/GrapefruitCat030/gfc_docker/pkg/net"
	gfc_pipe "github.com/GrapefruitCat030/gfc_docker/pkg/pipe"
	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"
	gfc_subsys "github.com/GrapefruitCat030/gfc_docker/pkg/subsystem"
	gfc_ufs "github.com/GrapefruitCat030/gfc_docker/pkg/ufs"
	gfc_uts "github.com/GrapefruitCat030/gfc_docker/pkg/uts"

	"github.com/docker/docker/pkg/reexec"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&runConf.ContainerName, "name", "n", "", "container name")
	runCmd.Flags().StringVarP(&runConf.RootFs, "rootfs", "r", "/root/project/gfc_docker/filesystem", "root filesystem path")
	runCmd.Flags().StringVarP(&runConf.MemLimit, "memory", "m", "20m", "memory limit")
	runCmd.Flags().BoolVarP(&runConf.Tty, "tty", "t", false, "tty")
	runCmd.Flags().BoolVarP(&runConf.Detach, "detach", "d", false, "detach")
	runCmd.Flags().StringVarP(&runConf.Volume, "volume", "v", "", "mount volume, format: <host_path>:<container_path>")
	runCmd.Flags().StringArrayVarP(&runConf.Env, "env", "e", []string{}, "environment variables")
	runCmd.Flags().StringVarP(&runConf.Net, "network", "w", "", "specify network")
	runCmd.Flags().StringArrayVarP(&runConf.PortMapping, "port", "p", []string{}, "port mapping, format: <host_port>:<container_port>")

	runConf.UFSer = &gfc_ufs.OverlayFS{}
	runConf.ContainerID = gfc_runinfo.GenerateRandomID(10)
	if runConf.ContainerName == "" {
		runConf.ContainerName = runConf.ContainerID
	}
	runConf.RootFs = filepath.Join(runConf.RootFs, runConf.ContainerID)
}

type RunConfig struct {
	ContainerID   string
	ContainerName string
	RootFs        string
	MemLimit      string
	Tty           bool
	Detach        bool
	UFSer         gfc_ufs.UnionFSer
	Volume        string
	Env           []string
	Net           string
	PortMapping   []string
}

var runConf RunConfig

var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Run a new container",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running command: ", args, " with config: ", runConf, " , default union filesystem: overlayfs")
		if runConf.Tty && runConf.Detach {
			fmt.Println("Error: cannot specify both -t and -d")
			os.Exit(1)
		}
		run(args)
	},
}

func run(args []string) {

	// ---- fork self process ----
	parentProc, pw := runNewProcess()          // new fork proc/self/exe  => reexec init map
	if err := parentProc.Start(); err != nil { // reexec init map => func runDetails()
		panic(err)
	}

	// ---- record container info ----
	containerInfo, err := gfc_runinfo.RecordContainerInfo(
		parentProc.Process.Pid,
		runConf.ContainerID,
		runConf.ContainerName,
		runConf.RootFs,
		runConf.Volume,
		runConf.PortMapping,
		args,
	)
	if err != nil {
		fmt.Printf("Error recording container info - %s\n", err)
	}
	fmt.Println("Container name: ", runConf.ContainerName)

	// ---- run netsetgo using default setting ----
	if err = gfc_net.ConnectEndpoint(runConf.Net, containerInfo); err != nil {
		fmt.Printf("Error connecting endpoint - %s\n", err)
	}

	// ---- set cgroup & subsys ----
	// ATTENTION: alpine does not support cgroup v2, need to comment out the following code
	cgroupManager := gfc_cgroup.NewCgroupManager("gfc_docker")
	defer cgroupManager.Remove()
	cgroupManager.Resource = &gfc_subsys.ResourceConfig{MemLimit: runConf.MemLimit}
	cgroupManager.Set()
	cgroupManager.Apply(parentProc.Process.Pid)

	// ---- write user command to pipe ----
	// Keep this behavior at the end to synchronize the parent and child processes
	if _, err := pw.WriteString(strings.Join(args, " ")); err != nil {
		fmt.Printf("Error writing to pipe - %s\n", err)
	}
	pw.Close()

	if runConf.Tty {
		if err := parentProc.Wait(); err != nil {
			fmt.Printf("Error waiting for the reexec.Command - %s\n", err)
			os.Exit(1)
		}
		if err := gfc_runinfo.DeleteContainerInfo(runConf.ContainerName); err != nil { // TODO: if detach container over?
			fmt.Printf("Error deleting container info - %s\n", err)
			os.Exit(1)
		}
		if err := gfc_ufs.DeleteWorkSpace(runConf.RootFs, runConf.Volume, runConf.UFSer); err != nil {
			fmt.Printf("Error deleting union filesystem - %s\n", err)
			os.Exit(1)
		}
	}
}

func runNewProcess() (*exec.Cmd, *os.File) {
	// ---- fork self process ----
	// ATTENTION: for success running in Alpine(root),comment out the NEWUSER flag and mappings
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
	// ---- set pipe for child process ----
	pr, pw, err := gfc_pipe.NewPipe()
	if err != nil {
		fmt.Printf("Error creating pipe - %s\n", err)
		os.Exit(1)
	}
	parentProc.ExtraFiles = []*os.File{pr}
	// ---- set environment variables ----
	parentProc.Env = append(os.Environ(), runConf.Env...)
	// ---- set union filesystem ----
	if err := gfc_ufs.NewWorkSpace(runConf.RootFs, runConf.Volume, runConf.UFSer); err != nil { //TODO: each container has its own workspace
		fmt.Printf("Error setting up union filesystem - %s\n", err)
		os.Exit(1)
	}
	parentProc.Dir = filepath.Join(runConf.RootFs, runConf.UFSer.WorkSpace())
	// ---- set tty, or redirect to logfile ----
	if runConf.Tty {
		parentProc.Stdin = os.Stdin
		parentProc.Stdout = os.Stdout
		parentProc.Stderr = os.Stderr
	} else {
		logPath := "/var/log/gfc_docker/" + runConf.ContainerName + "/container.log"
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0777); err != nil {
			fmt.Printf("Error creating log directory - %s\n", err)
			os.Exit(1)
		}
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("Error opening log file - %s\n", err)
			os.Exit(1)
		}
		parentProc.Stdout = f
		parentProc.Stderr = f
	}
	return parentProc, pw
}

func runDetails() {
	// 1. read user command from pipe
	cmdArr := readUserCmd()
	// 2. setup new root filesystem
	setNewRootfs()
	// 3. setup new hostname
	if err := gfc_uts.AssignHostName(); err != nil {
		fmt.Printf("Error assigning hostname - %s\n", err)
		os.Exit(1)
	}
	// 4. setup new network
	// if err := gfc_net.WaitNetwork(); err != nil {
	// 	fmt.Printf("Error waiting for network - %s\n", err)
	// 	os.Exit(1)
	// }
	// 5. execute command
	execCommand(cmdArr)
}

// ---- helper functions ----

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

func setNewRootfs() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory - %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Current location: ", pwd)
	if err := gfc_fs.SetMountPrivate(); err != nil {
		fmt.Printf("Error mounting independent - %s\n", err)
		os.Exit(1)
	}
	if err := gfc_fs.PivotRoot(pwd); err != nil {
		fmt.Printf("PivotRoot Error: %+v\n", err)
		os.Exit(1)
	}
	if err := gfc_fs.MountProc(); err != nil {
		fmt.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}
	if err := gfc_fs.MountTmpfs(); err != nil {
		fmt.Printf("Error mounting /tmp - %s\n", err)
		os.Exit(1)
	}
}

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
	env := append(os.Environ(), "PS1=[gfc_container] # ")
	if err := syscall.Exec(cmdPath, command, env); err != nil {
		fmt.Printf("Error executing shell - %s\n", err)
		os.Exit(1)
	}
}
