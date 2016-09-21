package main

import(
  "fmt"
  "flag"
  "errors"
  "os"
)

// tyk-cli <module> <submodule> <command> [--options] args...

var module, submodule, command string

func init() {
}

func main() {
  fmt.Println("tyk-cli:", flag.CommandLine, os.Args)
  fmt.Println("os.Args (length) = ", len(os.Args))
  if len(os.Args) == 1 {
    fmt.Println("No module specified.")
    os.Exit(1)
  } else if len(os.Args) == 2 {
    fmt.Println("No command specified.")
    os.Exit(1)
  }


  module = os.Args[1]
  command = os.Args[2]

  fmt.Println("module =", module)
  fmt.Println("command =", command)

  var err error

  switch module {
  case "bundle":
    fmt.Println("Using bundle module.")
    err = bundle(command)
  default:
    err = errors.New("Invalid module")
  }

  if err != nil {
    fmt.Println("Error:", err)
    os.Exit(1)
  }
}

func bundle(command string) (err error) {
  fmt.Println("calling bundle w/ command: ", command)
  switch command {
  case "build":
    fmt.Println("Build bundle.")
  default:
    err = errors.New("Invalid command.")
  }
  return err
}
