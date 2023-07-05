package main

import (
	"os"
	stevedore "stevedore/src"
)

func main() {
	err := stevedore.Parse()
	if err != nil {
		os.Exit(1)
	}
}
