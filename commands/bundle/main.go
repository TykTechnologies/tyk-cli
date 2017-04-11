package bundle

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/TykTechnologies/goverify"
	"github.com/TykTechnologies/tyk/apidef"
)

var (
	bundleOutput, privKey string
	forceInsecure         *bool
)

const defaultBundleOutput = "./bundle.zip"

func init() {
}

// Bundle will handle the bundle command calls.
func Bundle(command string, thisBundleOutput string, thisPrivKey string, thisForceInsecure *bool) (err error) {
	bundleOutput = thisBundleOutput
	privKey = thisPrivKey
	forceInsecure = thisForceInsecure

	switch command {
	case "build":
		manifestPath := "./manifest.json"
		if _, err = os.Stat(manifestPath); err == nil {
			var manifestData []byte
			manifestData, err = ioutil.ReadFile(manifestPath)

			var manifest apidef.BundleManifest
			err = json.Unmarshal(manifestData, &manifest)

			if err != nil {
				fmt.Println("Couldn't parse manifest file!")
				break
			}

			err = BundleValidateManifest(&manifest)

			if err != nil {
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

// BundleValidateManifest will validate the manifest file before building a bundle.
func BundleValidateManifest(manifest *apidef.BundleManifest) (err error) {
	// Validate manifest file list:
	for _, file := range manifest.FileList {
		if _, statErr := os.Stat(file); statErr != nil {
			err = errors.New("Referencing a nonexistent file: " + file)
			break
		}
	}

	// The file list references a nonexistent file:
	if err != nil {
		return err
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
func bundleBuild(manifest *apidef.BundleManifest) (err error) {
	var useSignature bool

	if bundleOutput == "" {
		fmt.Println("No output specified, using bundle.zip")
		bundleOutput = defaultBundleOutput
	}

	if privKey != "" {
		fmt.Println("The bundle will be signed.")
		useSignature = true
	}

	var bundleData bytes.Buffer

	for _, file := range manifest.FileList {
		var data []byte
		data, err = ioutil.ReadFile(file)

		if err != nil {
			fmt.Println("*** Error: ", err)
			return err
		}

		bundleData.Write(data)
	}

	// Update the manifest file:
	manifest.Checksum = fmt.Sprintf("%x", md5.Sum(bundleData.Bytes()))

	// If a private key is specified, sign the data:
	if useSignature {
		var signer goverify.Signer
		signer, err = goverify.LoadPrivateKeyFromFile(privKey)

		if err != nil {
			// Error: Couldn't read the private key
			return err
		}
		var signed []byte
		signed, err = signer.Sign(bundleData.Bytes())

		if err != nil {
			// Error: Couldn't sign the data.
			return err
		}

		manifest.Signature = base64.StdEncoding.EncodeToString(signed)
	} else {
		if *forceInsecure == false {
			fmt.Print("The bundle will be unsigned, type \"y\" to confirm: ")
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			if text != "y\n" {
				fmt.Println("Aborting")
				os.Exit(1)
			}
		}
	}

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
