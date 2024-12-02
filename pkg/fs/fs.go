package fs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

// CheckRootFS is a function that will check if the rootFS path exists
func CheckRootFS(rootFSPath string) {
	if _, err := os.Stat(rootFSPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("RootFS %+v path does not exist\n", rootFSPath)
			os.Exit(1)
		} else {
			panic(err)
		}
	}
}

// MountProc is a function that will mount the proc filesystem to the newroot
func MountProc(newroot string) error {
	source := "proc"
	target := filepath.Join(newroot, "/proc")
	fstype := "proc"
	// flags := 0
	flags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	os.MkdirAll(target, 0755)
	if err := syscall.Mount(source, target, fstype, uintptr(flags), ""); err != nil {
		return err
	}

	return nil
}

// PivotRoot is a function that will pivot the root filesystem to the newRoot
func PivotRoot(newRoot string) error {
	// newRoot => filesystem/alpine-fs
	// putOld => filesystem/alpine-fs/.pivot_root
	preRoot := "/.pivot_root"
	putOld := filepath.Join(newRoot, preRoot)

	// newroot and putold must not be on the same filesystem as the current root /
	// so we need to bind mount the new root to a new directory and pivot_root to it
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		log.Fatal("Error mounting newRoot: ", err)
		return err
	}

	// Create putOld and pivot_root
	if err := os.MkdirAll(putOld, 0700); err != nil {
		log.Fatal("Error creating putOld: ", err)
		return err
	}
	fmt.Printf("putOld: %+v\n", putOld)
	fmt.Printf("newRoot: %+v\n", newRoot)
	if err := syscall.PivotRoot(newRoot, putOld); err != nil {
		log.Fatal("Error pivoting root: ", err)
		// unmount
		if err := syscall.Unmount(newRoot, syscall.MNT_DETACH); err != nil {
			log.Fatal("Error unmounting newRoot: ", err)
			return err
		}
		return err
	}

	// ensure the current working directory is the new root
	if err := syscall.Chdir("/"); err != nil {
		log.Fatal("Error changing directory: ", err)
		return err
	}

	// unmount putOld, as it's no longer needed, and remove the directory
	putOld = preRoot
	if err := syscall.Unmount(putOld, syscall.MNT_DETACH); err != nil {
		log.Fatal("Error unmounting putOld: ", err)
		return err
	}
	testPath := "/kksk" // TODO: use putOld
	if err := os.RemoveAll(testPath); err != nil {
		log.Fatal("Error removing testPath: ", err)
		return err
	}

	return nil
}
