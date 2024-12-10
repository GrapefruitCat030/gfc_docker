package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logCmd)
}

var logCmd = &cobra.Command{
	Use:   "log [container-name]",
	Short: "Show the log of a container",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log(args)
	},
}

func log(args []string) {
	containerName := args[0]
	containerLogPath := "/var/log/gfc_docker/" + containerName + "/container.log"

	logFile, err := os.Open(containerLogPath)
	if err != nil {
		fmt.Println("Error: cannot open log file: ", containerLogPath, " , error: ", err)
		os.Exit(1)
	}
	defer logFile.Close()

	msg, err := io.ReadAll(logFile)
	if err != nil {
		fmt.Printf("Error reading from pipe - %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(msg))
}
