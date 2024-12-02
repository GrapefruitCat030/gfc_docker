package subsystem

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getCgroupHierarchyMount(subsys string) string {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsys {
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return ""
	}
	return ""
}

func getAbsCgroupPath(subsys, cgroupPath string) (string, error) {
	subsysPath := getCgroupHierarchyMount(subsys)
	resPath := filepath.Join(subsysPath, cgroupPath)
	if err := os.MkdirAll(resPath, 0755); err != nil {
		log.Fatalf("Error creating cgroup %s: %s", cgroupPath, err)
		return "", err
	}
	return resPath, nil
}
