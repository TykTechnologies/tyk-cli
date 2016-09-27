package main

import(
  "github.com/TykTechnologies/tykcommon"
  "github.com/TykTechnologies/goverify"

  "encoding/base64"
  "encoding/json"
  "fmt"
  "flag"
  "errors"
  "io/ioutil"
  "bytes"
  "archive/zip"
  "crypto/md5"
  "strings"
  "os"
  "io"
)

// tyk-cli <module> <submodule> <command> [--options] args...

var module, submodule, command string

var bundleOutput, privKey string

const(
  defaultBundleOutput = "bundle.zip"
)

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

  flag.Parse()
}

// main is the entrypoint.
func main() {
  fmt.Println("tyk-cli:", flag.CommandLine, os.Args)

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
  // Validate manifest file list:
  for _, file := range manifest.FileList {
    if _, statErr := os.Stat(file); statErr != nil {
      err = errors.New("Referencing a nonexistent file: " + file)
      break
    }
  }

  // The custom middleware block must specify at least one hook:
  var definedHooks int
  definedHooks = len(manifest.CustomMiddleware.Pre) + len(manifest.CustomMiddleware.Post) + len(manifest.CustomMiddleware.PostKeyAuth)

  // We should count the auth check middleware (single), if it's present:
  if manifest.CustomMiddleware.AuthCheck.Name != "" {
    definedHooks++
  }

  if definedHooks == 0 {
    err = errors.New("No hooks defined!")
    return err
  }

  // The custom middleware block must specify a driver:
  if manifest.CustomMiddleware.Driver == "" {
    err = errors.New("No driver specified!")
    return err
  }

  return err
}

// bundleBuild will build and generate a bundle file.
func bundleBuild(manifest *tykcommon.BundleManifest) (err error) {
  var useSignature bool

  if bundleOutput == "" {
    fmt.Println("No output specified, using bundle.zip")
    bundleOutput = defaultBundleOutput
  }

  if privKey == "" {
    // Warning?
    fmt.Println("The bundle won't be signed.")
  } else {
    fmt.Println("The bundle will be signed.")
    useSignature = true
  }

  var signer goverify.Signer

  if useSignature {
    signer, err = goverify.LoadPrivateKeyFromFile(privKey)
    if err != nil {
      return err
    }
  }

  // Checksum and signature:

  var bundleChecksums []string
  var bundleSignatures []string

  for _, file := range manifest.FileList {
    var data []byte
    data, err = ioutil.ReadFile(file)
    if err != nil {
      fmt.Println("*** Error: ", err)
      return err
    }
    hash := fmt.Sprintf("%x", md5.Sum(data))
    bundleChecksums = append(bundleChecksums, hash)

    if useSignature {
      var signed []byte
      signed, err = signer.Sign(data)

      sig := base64.StdEncoding.EncodeToString(signed)
      bundleSignatures = append(bundleSignatures, sig)
      fmt.Printf("Signature: %v %s\n", sig, file)
    }
  }

  mergedChecksums := strings.Join(bundleChecksums, "")
  mergedSignatures := strings.Join(bundleSignatures, "")

  // Update the manifest file:

  manifest.Checksum = fmt.Sprintf("%x", md5.Sum([]byte(mergedChecksums)))
  manifest.Signature = mergedSignatures

  var newManifestData []byte
  newManifestData, err = json.Marshal(&manifest)

  // Write the bundle file:
  buf := new(bytes.Buffer)
  bundleWriter := zip.NewWriter(buf)

  for _, file := range manifest.FileList {
    var outputFile io.Writer
    outputFile, err = bundleWriter.Create(file)
    if err != nil {
      return err
    }
    var data []byte
    data, err = ioutil.ReadFile(file)

    _, err = outputFile.Write(data)

    if err != nil {
      return err
    }
  }

  // Write manifest file:
  var newManifest io.Writer
  newManifest, err = bundleWriter.Create("manifest.json")
  _, err = newManifest.Write(newManifestData)

  err = bundleWriter.Close()
  err = ioutil.WriteFile(bundleOutput, buf.Bytes(), 0755)

  return err
}
