package main

import (
	"github.com/1f349/voidterm"
	"github.com/1f349/voidterm/termutil"
	"log"
)

func main() {
	updateChan := make(chan struct{})

	void := voidterm.New("unix:///var/run/docker.sock", "aa5e8ebe40c4")
	void.Run(updateChan, 14, 11)
	term := termutil.New(termutil.WithShell("/usr/bin/bash"))
	err := term.Run(updateChan, 14, 11)
	if err != nil {
		log.Fatal(err)
	}
}
