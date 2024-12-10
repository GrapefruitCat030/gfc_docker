package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/tabwriter"

	gfc_runinfo "github.com/GrapefruitCat030/gfc_docker/pkg/runinfo"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(psCmd)
}

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all containers",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("List all containers")
		ps()
	},
}

func ps() {
	containerInfos := make([]*gfc_runinfo.ContainerInfo, 0)
	err := filepath.WalkDir(gfc_runinfo.DefaultInfoLocation, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 如果是文件，读取容器信息
		if !d.IsDir() {
			containerInfo, err := getContainerInfo(path)
			if err != nil {
				return err
			}
			containerInfos = append(containerInfos, containerInfo)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error: cannot walk through the container info directory: ", gfc_runinfo.DefaultInfoLocation, " , error: ", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintln(w, "PID\tID\tNAME\tCOMMAND\tCREATED\tSTATUS")
	for _, containerInfo := range containerInfos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			containerInfo.Pid,
			containerInfo.Id,
			containerInfo.Name,
			containerInfo.Command,
			containerInfo.CreatedTime,
			containerInfo.Status)
	}
	if err := w.Flush(); err != nil {
		fmt.Println("Error: cannot flush the tabwriter: ", err)
		os.Exit(1)
	}
}

func getContainerInfo(filePath string) (*gfc_runinfo.ContainerInfo, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	containerInfo := &gfc_runinfo.ContainerInfo{}
	if err := json.NewDecoder(f).Decode(containerInfo); err != nil {
		return nil, err
	}
	return containerInfo, nil
}
