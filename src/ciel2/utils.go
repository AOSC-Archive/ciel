package main

import (
	"encoding/base64"
	"math/rand"
)

func randomFilename() string {
	const SIZE = 8
	rd := make([]byte, SIZE)
	if _, err := rand.Read(rd); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(rd)
}
