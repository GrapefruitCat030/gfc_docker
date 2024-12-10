package runinfo

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"created_time"`
	Status      string `json:"status"`
}

const (
	STATUS_STOPPED = "stopped"
	STATUS_RUNNING = "running"
	STATUS_EXITED  = "exited"
)

const (
	DefaultInfoLocation = "/var/run/gfc_docker/"
	ConfigName          = "config.json"
)

func RecordContainerInfo(pid int, id, name string, cmdArr []string) error {
	containerInfo := &ContainerInfo{
		Pid:         fmt.Sprintf("%d", pid),
		Id:          id,
		Name:        name,
		Command:     strings.Join(cmdArr, " "),
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:      STATUS_RUNNING,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return err
	}
	jsonStr := string(jsonBytes)

	infoLocation := filepath.Join(DefaultInfoLocation, containerInfo.Name)
	if err := os.MkdirAll(infoLocation, 0777); err != nil {
		return err
	}
	fileName := filepath.Join(infoLocation, ConfigName)
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(jsonStr); err != nil {
		return err
	}
	return nil
}

func DeleteContainerInfo(containerName string) error {
	infoLocation := filepath.Join(DefaultInfoLocation, containerName)
	return os.RemoveAll(infoLocation)
}

func GenerateRandomID(idLen int) string {
	const letters = "0123456789"
	b := make([]byte, idLen)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		b[i] = letters[num.Int64()]
	}
	return string(b)
}
