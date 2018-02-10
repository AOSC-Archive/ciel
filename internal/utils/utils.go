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

func MustMkdir(p string) {
	if err := os.Mkdir(p, 0755); err != nil {
		log.Fatal(err)
	}
}
