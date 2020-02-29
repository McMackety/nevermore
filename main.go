package main

import (
	"github.com/McMackety/nevermore/config"
	"github.com/McMackety/nevermore/field"
	"log"
)

var Version string = "0.0.1" // This will be injected at build time, don't worry about it :)
var GitCommit string = "dev" // This will be injected at build time, don't worry about it :)

func main() {
	log.Printf("Starting nevermore v%s (Commit %s)", Version, GitCommit)
	config.LoadConfig()
	field.InitField()
}