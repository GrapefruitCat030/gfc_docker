package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	gfc_fs "github.com/GrapefruitCat030/gfc_docker/pkg/fs"
	gfc_uts "github.com/GrapefruitCat030/gfc_docker/pkg/uts"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	fmt.Printf("args=%+v\n", os.Args)
	reexec.Register("init-func", initFunc)
	if reexec.Init() {
		os.Exit(0)
	}
}

func initFunc() {
	// newRoot := os.Args[1]
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
	runShell()
}

func runShell() {
	cmd := exec.Command("sh")
	cmd.Env = []string{"PS1=-[namespace-process]-# "}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func main() {
	cmd := reexec.Command("init-func", "/tmp/ns-proc/rootfs") // reexec fork /proc/self/exe
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil { // non-blocking
		panic(err)
	}

	// run netsetgo using default args
	// gfc_net.SetNetwork(cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for the reexec.Command - %s\n", err)
		os.Exit(1)
	}
}
