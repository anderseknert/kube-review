package main

import (
	"log"

	"kube-review/cmd"
)

func main() {
	// Remove date/time from (likely fatal) log messages
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	cmd.Execute()
}
