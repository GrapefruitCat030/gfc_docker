package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// CheckPathExist is a function that will check if the path exists
func CheckPathExist(rootFSPath string) (bool, error) {
	if _, err := os.Stat(rootFSPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
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
