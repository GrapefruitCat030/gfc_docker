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
	Status      string `json:"status"`
	Command     string `json:"command"`
	CreatedTime string `json:"created_time"`
	RootFs      string `json:"rootfs"`
	Volume      string `json:"volume"`
	//TODO: network, etc.
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

func RecordContainerInfo(pid int, id, name, rootfs, volume string, cmdArr []string) error {
	containerInfo := &ContainerInfo{
		Pid:         fmt.Sprintf("%d", pid),
		Id:          id,
		Name:        name,
		Status:      STATUS_RUNNING,
		Command:     strings.Join(cmdArr, " "),
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
		RootFs:      rootfs,
		Volume:      volume,
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

func GetContainerInfo(name string) (*ContainerInfo, error) {
	dirPath := filepath.Join(DefaultInfoLocation, name)
	configPath := filepath.Join(dirPath, ConfigName)
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(bytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}

func GetContainerPid(name string) string {
	dirPath := filepath.Join(DefaultInfoLocation, name)
	configPath := filepath.Join(dirPath, ConfigName)
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(bytes, &containerInfo); err != nil {
		return ""
	}
	return containerInfo.Pid
}

func GetContainerEnv(pid string) ([]string, error) {
	filePath := fmt.Sprintf("/proc/%s/environ", pid)
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bytes), "\u0000"), nil
}
