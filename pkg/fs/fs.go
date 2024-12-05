package fs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

// CheckRootFS is a function that will check if the rootFS path exists
func CheckRootFS(rootFSPath string) error {
	if _, err := os.Stat(rootFSPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("RootFS %+v path does not exist", rootFSPath)
		} else {
			return err
		}
	}
	return nil
}

// MountProc is a function that will mount the proc filesystem to the newroot
func MountProc() error {
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Fatal("Error mounting /proc: ", err)
		return err
	}
	return nil
}

func MountTmpfs() error {
	defaultMountFlags := syscall.MS_NOSUID | syscall.MS_STRICTATIME
	defaultMountMode := "mode=755"
	if err := syscall.Mount("tmpfs", "/tmp", "tmpfs", uintptr(defaultMountFlags), defaultMountMode); err != nil {
		log.Fatal("Error mounting /tmp: ", err)
		return err
	}
	return nil
}

// PivotRoot is a function that will pivot the root filesystem to the newRoot
func PivotRoot(newRoot string) error {

	// newroot and putold must not be on the same filesystem as the current root /
	// so we need to bind mount the new root to a new directory and pivot_root to it
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		log.Fatal("Error mounting newRoot: ", err)
		return err
	}

	pivotDir := filepath.Join(newRoot, ".pivot_root")
	if err := os.MkdirAll(pivotDir, 0777); err != nil {
		log.Fatal("Error creating putOld: ", err)
		return err
	}
	fmt.Printf("putOld: %+v\n", pivotDir)
	fmt.Printf("newRoot: %+v\n", newRoot)
	if err := syscall.PivotRoot(newRoot, pivotDir); err != nil {
		log.Fatal("Error pivoting root: ", err)
		return err
	}

	// ensure the current working directory is the new root
	if err := syscall.Chdir("/"); err != nil {
		log.Fatal("Error changing directory: ", err)
		return err
	}

	// unmount putOld, as it's no longer needed, and remove the directory
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		log.Fatal("Error unmounting putOld: ", err)
		return err
	}
	if err := os.RemoveAll(pivotDir); err != nil {
		log.Fatal("Error removing testPath: ", err)
		return err
	}

	return nil
}
