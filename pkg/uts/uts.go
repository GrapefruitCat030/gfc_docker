package uts

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"syscall"
)

// GenerateRandomHostname 生成一个随机的主机名
func generateRandomHostname(nameLen int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, nameLen)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err) // 处理错误
		}
		b[i] = letters[num.Int64()]
	}
	return string(b)
}

func AssignHostName() error {
	hostName := generateRandomHostname(6)
	if err := syscall.Sethostname([]byte(hostName)); err != nil {
		return fmt.Errorf("error setting hostname - %s", err)
	}
	return nil
}
