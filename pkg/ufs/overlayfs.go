package ufs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	gfc_fs "github.com/GrapefruitCat030/gfc_docker/pkg/fs"
)

type OverlayFS struct{}

const (
	overlayReadOnlyDir = "busyboxfs"
	overlayWriteDir    = "upper"
	overlayWorkDir     = "work"
	overlayMountDir    = "merged"
)

func (ofs *OverlayFS) WorkSpace() string {
	return overlayMountDir
}

func (ofs *OverlayFS) CreateReadOnlyLayer(rootPath string) error {
	path := filepath.Join(filepath.Dir(rootPath), overlayReadOnlyDir)
	ok, err := gfc_fs.CheckPathExist(path)
	if err != nil {
		return err
	}
	if !ok {
		if err := os.MkdirAll(path, 0777); err != nil {
			return err
		}
	}
	return nil
}

func (ofs *OverlayFS) CreateWriteLayer(rootPath string) error {
	upperDir := filepath.Join(rootPath, overlayWriteDir)
	workDir := filepath.Join(rootPath, overlayWorkDir)
	if err := os.MkdirAll(upperDir, 0777); err != nil {
		return err
	}
	if err := os.MkdirAll(workDir, 0777); err != nil {
		return err
	}
	return nil
}

func (ofs *OverlayFS) MountUFS(rootPath string) error {
	mergedDir := filepath.Join(rootPath, overlayMountDir)
	if err := os.MkdirAll(mergedDir, 0777); err != nil {
		return err
	}
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		filepath.Join(filepath.Dir(rootPath), overlayReadOnlyDir),
		filepath.Join(rootPath, overlayWriteDir),
		filepath.Join(rootPath, overlayWorkDir),
	)
	if err := syscall.Mount("overlay", mergedDir, "overlay", 0, opts); err != nil {
		return err
	}
	return nil
}

func (ofs *OverlayFS) UMountUFS(rootPath string) error {
	mergedDir := filepath.Join(rootPath, overlayMountDir)
	if err := syscall.Unmount(mergedDir, syscall.MNT_DETACH); err != nil {
		return err
	}
	if err := os.RemoveAll(mergedDir); err != nil {
		return err
	}
	return nil
}

func (ofs *OverlayFS) DeleteWriteLayer(rootPath string) error {
	upperDir := filepath.Join(rootPath, overlayWriteDir)
	workDir := filepath.Join(rootPath, overlayWorkDir)
	if err := os.RemoveAll(upperDir); err != nil {
		return err
	}
	if err := os.RemoveAll(workDir); err != nil {
		return err
	}
	if err := os.RemoveAll(rootPath); err != nil {
		return err
	}
	return nil
}

func (ofs *OverlayFS) MountVolume(rootPath, volumeMapPath string) error {
	p := strings.Split(volumeMapPath, ":")
	if len(p) != 2 {
		return fmt.Errorf("volume map path %s invalid", volumeMapPath)
	}
	hostPath := p[0]
	containerPath := p[1]

	if ok, err := gfc_fs.CheckPathExist(hostPath); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("path %s not exist", hostPath)
	}

	if err := os.MkdirAll(filepath.Join(rootPath, overlayMountDir, containerPath), 0777); err != nil {
		return err
	}
	if err := syscall.Mount(hostPath, filepath.Join(rootPath, overlayMountDir, containerPath), "", syscall.MS_BIND, ""); err != nil {
		return err
	}
	return nil
}

func (ofs *OverlayFS) UMountVolume(rootPath, volumeMapPath string) error {
	p := strings.Split(volumeMapPath, ":")
	if len(p) != 2 {
		return fmt.Errorf("volume map path %s invalid", volumeMapPath)
	}
	containerPath := p[1]
	if err := syscall.Unmount(filepath.Join(rootPath, overlayMountDir, containerPath), syscall.MNT_DETACH); err != nil {
		return err
	}
	return nil
}
