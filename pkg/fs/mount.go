package fs

import "syscall"

// SetMountPrivate is a function that will set the mount namespace to private
func SetMountPrivate() error {
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
