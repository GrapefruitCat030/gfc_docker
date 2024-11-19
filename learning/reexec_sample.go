package main

import (
	"log"
	"os"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	log.Printf("init start, os args: %+v, os path: %s\n", os.Args, os.Args[0])
	reexec.Register("child", child)
	if reexec.Init() {
		log.Println("this is child init")
		os.Exit(0)
	}
	log.Println("this is parent init")
}

func child() {
	log.Println("child start")
}

func main() {
	//打印开始
	log.Printf("main start, os.Args = %+v\n", os.Args)
	//执行子函数
	cmd := reexec.Command("child")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Panicf("failed to run command: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Panicf("failed to wait command: %s", err)
	}

	log.Println("main exit")
}
