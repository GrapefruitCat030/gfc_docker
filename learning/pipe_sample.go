package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func pipe_main() {
	if len(os.Args) > 1 && os.Args[1] == "child" {
		// 子进程
		fmt.Println("Running as child")
		// 从父进程读取数据
		r := os.NewFile(uintptr(3), "pipe")
		fmt.Println("Pipe opened for reading, r:", r)

		msg, err := io.ReadAll(r)
		if err != nil {
			fmt.Println("Error reading from pipe:", err)
			return
		}
		r.Close()
		// fmt.Println("Child process received:", os.Args[2])
		fmt.Println("Child process received:", string(msg))
		fmt.Println("Child process received len:", len(msg))
		// time.Sleep(100 * time.Second)
		return
	}

	// 父进程

	// pipe
	r, w, err := os.Pipe()
	if err != nil {
		fmt.Println("Error creating pipe:", err)
		return
	}
	fmt.Println("Pipe created, r:", r, "w:", w)

	longArg := strings.Repeat("a", 10) // 一个非常长的参数
	cmd := exec.Command("/proc/self/exe", "child")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{r}

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting the command:", err)
		return
	}
	// 向子进程发送数据
	n, err := w.Write([]byte(longArg))
	if err != nil {
		fmt.Println("Error writing to pipe:", err)
		return
	}
	fmt.Println("Wrote", n, "bytes to the pipe")
	r.Close()
	w.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Error waiting for the command:", err)
		return
	}
}
