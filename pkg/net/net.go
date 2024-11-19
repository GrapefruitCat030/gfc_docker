package net

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"time"
)

func WaitNetwork() error {
	maxWait := time.Second * 60
	checkInterval := time.Second
	timeStarted := time.Now()

	for {
		fmt.Printf("status: waiting network ...\n")
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}

		if len(interfaces) < 1 {
			return nil
		}

		if time.Since(timeStarted) < maxWait {
			return fmt.Errorf("Timeout after %s waiting for network", maxWait)
		}

		time.Sleep(checkInterval)
	}
}

func SetNetwork(pid int) error {
	pidStr := fmt.Sprintf("%d", pid)
	netsetPath := "./scripts/netsettor.sh"
	netsetCmd := exec.Command("sudo", netsetPath, pidStr)
	var out bytes.Buffer
	var stderr bytes.Buffer
	netsetCmd.Stdout = &out
	netsetCmd.Stderr = &stderr
	if err := netsetCmd.Start(); err != nil {
		fmt.Printf("Error running netsetg:%s, stderr:%s, stdout:%s", fmt.Sprint(err), stderr.String(), out.String())
		return err
	}
	fmt.Printf("run netsetter: stdout:%s \n", out.String())
	return nil
}
