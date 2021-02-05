package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/veertuinc/packer-builder-veertu-anka/builder/ankavm"
	"github.com/veertuinc/packer-builder-veertu-anka/post-processor/ankaregistry"
)

var version = "SNAPSHOT"
var commit = ""

func main() {
	if commit == "" {
		log.Printf("packer-builder-veertu-anka version: %s", version)
	} else {
		log.Printf("packer-builder-veertu-anka version: %s+%s", version, commit)
	}
	pps := plugin.NewSet()
	pps.RegisterBuilder("vm", new(ankavm.Builder))
	pps.RegisterPostProcessor("registry", new(ankaregistry.PostProcessor))
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
