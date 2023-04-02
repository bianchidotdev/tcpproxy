package main

import (
	"fmt"
	"log"
	"os"

	"github.com/akamensky/argparse"

	"github.com/michaeldbianchi/tcpproxy/core"
)

// main
func main() {
	parser := argparse.NewParser("proxy", "Sets up a multi-port -> multi-target tcp proxy")
	configFile := parser.String("c", "config", &argparse.Options{Default: "./config.yaml", Help: "Config file for proxy apps/ports/targets"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	config, err := core.ReadConfig(*configFile)
	if err != nil {
		log.Fatal("Failed to read config", err)
		os.Exit(1)
	}

	log.Println("Starting a proxy with the following config:")
	for _, app := range config.Apps {
		log.Println("- Name:", app.Name, "- Ports:", app.Ports, "- Targets:", app.Targets)
	}

	core.Serve(config)
}
