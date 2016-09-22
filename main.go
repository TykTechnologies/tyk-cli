package main

import(
  "github.com/TykTechnologies/tykcommon"

  "encoding/json"
  "fmt"
  "flag"
  "errors"
  "io/ioutil"
  "os"
)

// tyk-cli <module> <submodule> <command> [--options] args...

var module, submodule, command string

func init() {
}

// main is the entrypoint.
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

// bundle will handle the bundle command calls.
func bundle(command string) (err error) {
  switch command {
  case "build":
    var manifestPath = "./manifest.json"
    if _, err := os.Stat(manifestPath); err == nil {
      var manifestData []byte
      manifestData, err = ioutil.ReadFile(manifestPath)

      var manifest tykcommon.BundleManifest
      err = json.Unmarshal(manifestData, &manifest)

      if err != nil {
        fmt.Println("Couldn't parse manifest file!")
        break
      }

      err = bundleValidateManifest(&manifest)

      if err != nil {
        fmt.Println("Bundle validation error:")
        fmt.Println(err)
        break
      }

      // The manifest is valid, we should do the checksum and sign step at this point.
      bundleBuild(&manifest)

    } else {
      err = errors.New("Manifest file doesn't exist.")
    }
  default:
    err = errors.New("Invalid command.")
  }
  return err
}

// bundleValidateManifest will validate the manifest file before building a bundle.
func bundleValidateManifest(manifest *tykcommon.BundleManifest) (err error) {
  for _, file := range manifest.FileList {
    if _, statErr := os.Stat(file); statErr != nil {
      err = errors.New("Referencing a nonexistent file: " + file)
    }
  }
  // TODO: validate the custom middleware block.
  return err
}

func bundleBuild(manifest *tykcommon.BundleManifest) (err error) {
  return err
}
