package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	bundle "github.com/TykTechnologies/tyk-cli/bundle"
)

// tyk-cli <module> <submodule> <command> [--options] args...

var module, submodule, command string

var bundleOutput, privKey string
var forceInsecure, showVersion *bool

func init() {
	if len(os.Args) == 1 {
		fmt.Println("No module specified!")
		os.Exit(1)
	}
	if len(os.Args) == 2 {
		fmt.Println("No command specified!")
		os.Exit(1)
	}

	module = os.Args[1]
	command = os.Args[2]

	os.Args = os.Args[2:]

	flag.StringVar(&bundleOutput, "output", "", "Bundle output")
	flag.StringVar(&privKey, "key", "", "Key for bundle signature")
	forceInsecure = flag.Bool("y", false, "Skip bundle signing")
	showVersion = flag.Bool("v", false, "Show version")

	flag.Parse()
}

// main is the entrypoint.
func main() {
	if *showVersion {
		fmt.Println("tyk-cli", VERSION)
		os.Exit(0)
	}

	var err error

	switch module {
	case "bundle":
		fmt.Println("Using bundle module.")
		err = bundle.Bundle(command, bundleOutput, privKey, forceInsecure)
	default:
		err = errors.New("Invalid module")
	}

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
