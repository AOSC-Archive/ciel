package cieldir

import (
	"log"
	"os"
)

func mkdir(p string) {
	if err := os.Mkdir(p, 0755); err != nil {
		log.Fatal(err)
	}
}
