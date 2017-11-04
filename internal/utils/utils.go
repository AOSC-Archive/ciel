package utils

import (
	"log"
	"math/rand"
	"os"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Lock(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func Unlock(path string) bool {
	return os.Remove(path) == nil
}

func Locked(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Mkdir(p string) {
	if err := os.Mkdir(p, 0755); err != nil {
		log.Fatal(err)
	}
}

func RandomString(length int) (result string) {
	for i := 1; i <= length; i++ {
		result += string('a' + byte(rand.Intn(26)))
	}
	return
}
