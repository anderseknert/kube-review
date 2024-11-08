package main

import (
	"log"

	"github.com/anderseknert/kube-review/cmd"
)

func main() {
	// Remove date/time from (likely fatal) log messages
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	cmd.Execute()
}
