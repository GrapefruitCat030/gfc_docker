package cgroup

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
)

const (
	cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"
)

const (
	testMemoryLimitCgroup = "test-memory-limit"
)

func TestMemoryLimit(pid int) {
	newCgroupPath := filepath.Join(cgroupMemoryHierarchyMount, testMemoryLimitCgroup)
	if err := os.Mkdir(newCgroupPath, 0755); err != nil {
		log.Fatalf("Error creating cgroup %s: %s", testMemoryLimitCgroup, err)
	}
	// 将进程加入cgroup
	// 要注意, 由于 process migration 的存在, 当一个进程从一个cgroup移动到另一个cgroup时，
	// 默认情况下，该进程已经占用的内存还是统计在原来的cgroup里面，不会占用新cgroup的配额，但新分配的内存会统计到新的cgroup中
	if err := os.WriteFile(filepath.Join(newCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		log.Fatalf("Error writing pid %d to cgroup %s: %s", pid, testMemoryLimitCgroup, err)
	}
	if err := os.WriteFile(filepath.Join(newCgroupPath, "memory.limit_in_bytes"), []byte("20m"), 0644); err != nil {
		log.Fatalf("Error writing memory limit to cgroup %s: %s", testMemoryLimitCgroup, err)
	}
	// // 设置内存+swap限制
	// if err := os.WriteFile(filepath.Join(newCgroupPath, "memory.memsw.limit_in_bytes"), []byte("20m"), 0644); err != nil {
	// 	log.Fatalf("Error writing memory limit to cgroup %s: %s", testMemoryLimitCgroup, err)
	// }
}

func FreeMemoryLimit() {
	newCgroupPath := filepath.Join(cgroupMemoryHierarchyMount, testMemoryLimitCgroup)
	if err := os.RemoveAll(newCgroupPath); err != nil {
		log.Fatalf("Error removing cgroup %s: %s", testMemoryLimitCgroup, err)
	}
}

// AllocateMemory 分配固定大小的内存并保持分配状态
func AllocateMemory(size int) []byte {
	mem := make([]byte, size)
	for i := range mem {
		mem[i] = 0
	}
	return mem
}
