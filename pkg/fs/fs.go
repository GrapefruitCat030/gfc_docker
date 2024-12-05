package fs

import (
	"fmt"
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

// MountIndepent is a function that will mount the new mount namespace as independent
func MountIndepent() error {
	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以必须显式
	// 声明你要这个新的 mount namespace独立。
	// url: https://man7.org/linux/man-pages/man7/mount_namespaces.7.html#NOTES
	// systemd(1) automatically remounts all mounts as MS_SHARED on system startup. Thus, on most modern systems, the default propagation type is in practice MS_SHARED.
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return err
	}
	return nil
}

// MountProc is a function that will mount the proc filesystem to the newroot
func MountProc() error {
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		return err
	}
	return nil
}

// MountTmpfs is a function that will mount a tmpfs filesystem to the newroot
func MountTmpfs() error {
	defaultMountFlags := syscall.MS_NOSUID | syscall.MS_STRICTATIME
	defaultMountMode := "mode=755"
	if err := syscall.Mount("tmpfs", "/tmp", "tmpfs", uintptr(defaultMountFlags), defaultMountMode); err != nil {
		return err
	}
	return nil
}

// PivotRoot is a function that will pivot the root filesystem to the newRoot
func PivotRoot(newRoot string) error {

	// newroot and putold must not be on the same filesystem as the current root /
	// so we need to bind mount the new root to a new directory and pivot_root to it
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}

	pivotDir := filepath.Join(newRoot, ".pivot_root")
	if err := os.MkdirAll(pivotDir, 0777); err != nil {
		return err
	}
	fmt.Printf("putOld: %+v\n", pivotDir)
	fmt.Printf("newRoot: %+v\n", newRoot)
	if err := syscall.PivotRoot(newRoot, pivotDir); err != nil {
		return err
	}

	// ensure the current working directory is the new root
	if err := syscall.Chdir("/"); err != nil {
		return err
	}

	// unmount putOld, as it's no longer needed, and remove the directory
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return err
	}
	if err := os.RemoveAll(pivotDir); err != nil {
		return err
	}

	return nil
}
