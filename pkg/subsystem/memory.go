package subsystem

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type MemorySubsystem struct{}

func (s *MemorySubsystem) Name() string {
	return "memory"
}

func (s *MemorySubsystem) Set(cgroupPath string, resrcConf *ResourceConfig) error {
	path, err := getAbsCgroupPath(s.Name(), cgroupPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(path, "memory.limit_in_bytes"), []byte(resrcConf.MemLimit), 0644); err != nil {
		log.Fatalf("Error writing memory limit to cgroup %s: %s", cgroupPath, err)
		return err
	}
	// // 设置内存+swap限制
	// if err := os.WriteFile(filepath.Join(newCgroupPath, "memory.memsw.limit_in_bytes"), []byte("20m"), 0644); err != nil {
	// 	log.Fatalf("Error writing memory limit to cgroup %s: %s", testMemoryLimitCgroup, err)
	// }
	return nil
}

func (s *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	path, err := getAbsCgroupPath(s.Name(), cgroupPath)
	if err != nil {
		return err
	}
	// 将进程加入cgroup
	// 要注意, 由于 process migration 的存在, 当一个进程从一个cgroup移动到另一个cgroup时，
	// 默认情况下，该进程已经占用的内存还是统计在原来的cgroup里面，不会占用新cgroup的配额，但新分配的内存会统计到新的cgroup中
	if err := os.WriteFile(filepath.Join(path, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		log.Fatalf("Error writing pid %d to cgroup %s: %s", pid, path, err)
		return err
	}
	return nil
}

func (s *MemorySubsystem) Remove(cgroupPath string) error {
	path, err := getAbsCgroupPath(s.Name(), cgroupPath)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(path); err != nil {
		log.Fatalf("Error removing cgroup %s: %s", path, err)
	}
	return nil
}
