package fs

import (
	"os"
	"syscall"
)

// FileLocker 封装文件锁
type FileLocker struct {
	file *os.File
}

// NewFileLocker 创建一个新的文件锁
func NewFileLocker(filePath string) (*FileLocker, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &FileLocker{file: file}, nil
}

// Lock 加锁
func (fl *FileLocker) Lock() error {
	return syscall.Flock(int(fl.file.Fd()), syscall.LOCK_EX)
}

// Unlock 解锁
func (fl *FileLocker) Unlock() error {
	return syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
}

// Close 关闭文件
func (fl *FileLocker) Close() error {
	return fl.file.Close()
}
