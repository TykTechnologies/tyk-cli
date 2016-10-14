package main

import (
	"encoding/json"
	"fmt"
	"github.com/TykTechnologies/tyk-cli/bundle"
	"github.com/TykTechnologies/tykcommon"
	"io/ioutil"
	"os"
	"testing"
)

const (
	testManifestName  = "manifest.json"
	testPerm          = 0755
	testBundleName    = "bundle.zip"
	testBundleAltName = "mybundle.zip"
	testDummyFilename = "mymiddleware.py"
)

func TestBundleUnknownCommand(t *testing.T) {
	fmt.Println("a bundle test", t)
	var force bool
	var err error
	err = bundle.Bundle("unknown", "", "", &force)
	if err == nil {
		t.Fatal("Must return an error when the command doesn't exist.")
	}
}

func TestBundleBuildCommand(t *testing.T) {
	var force bool = true
	var err error
	err = bundle.Bundle("build", "", "", &force)
	if err == nil {
		t.Fatal("The manifest file doesn't exist, the build command should return an error.")
	}

	noManifest := []byte("")
	ioutil.WriteFile(testManifestName, noManifest, testPerm)

	err = bundle.Bundle("build", "", "", &force)
	if err == nil {
		t.Fatal("When the manifest is invalid, the build command should return an error.")
	}

	os.Remove(testManifestName)

	emptyManifest := []byte("{}")
	ioutil.WriteFile(testManifestName, emptyManifest, testPerm)

	err = bundle.Bundle("build", "", "", &force)
	if err == nil {
		t.Fatal("The manifest doesn't have any hooks, the build command should return an error.")
	}

	os.Remove(testManifestName)
}

func TestBundleValidateManifest(t *testing.T) {
	var manifest tykcommon.BundleManifest
	var err error
	err = bundle.BundleValidateManifest(&manifest)
	if err == nil {
		t.Fatal("The manifest is empty (no hooks defined), the validation step should return an error.")
	}

	manifest = tykcommon.BundleManifest{
		FileList: []string{"nonexistentfile"},
	}
	err = bundle.BundleValidateManifest(&manifest)
	if err == nil {
		t.Fatal("The manifest file references a nonexistent file, the validation step should return an error.")
	}

	manifest = tykcommon.BundleManifest{
		CustomMiddleware: tykcommon.MiddlewareSection{
			Pre: []tykcommon.MiddlewareDefinition{
				tykcommon.MiddlewareDefinition{},
			},
		},
	}

	err = bundle.BundleValidateManifest(&manifest)

	if err == nil {
		t.Fatal("The bundle validation step must return an error when no driver is specified.")
	}

	manifest = tykcommon.BundleManifest{
		CustomMiddleware: tykcommon.MiddlewareSection{
			Pre: []tykcommon.MiddlewareDefinition{
				tykcommon.MiddlewareDefinition{
					Name: "mymiddleware",
					Path: "./mymiddleware.py",
				},
			},
			Driver: tykcommon.OttoDriver,
		},
	}

	err = bundle.BundleValidateManifest(&manifest)
	if err != nil {
		t.Fatal("A minimal manifest should be valid.")
	}
}

func TestBundleBasicBuild(t *testing.T) {
	var force bool = true
	var err error
	var testManifest tykcommon.BundleManifest
	testManifest = tykcommon.BundleManifest{
		FileList: []string{"mymiddleware.py"},
		CustomMiddleware: tykcommon.MiddlewareSection{
			Pre: []tykcommon.MiddlewareDefinition{
				tykcommon.MiddlewareDefinition{
					Name: "mymiddleware",
					Path: "./mymiddleware.py",
				},
			},
			Driver: tykcommon.OttoDriver,
		},
	}

	ioutil.WriteFile(testDummyFilename, []byte(""), testPerm)

	var testManifestData []byte
	testManifestData, err = json.Marshal(&testManifest)
	if err != nil {
		t.Fatal("Couldn't marshal the test manifest.")
	}

	ioutil.WriteFile(testManifestName, testManifestData, testPerm)

	err = bundle.Bundle("build", "", "", &force)

	_, err = os.Stat(testBundleName)
	if err != nil {
		t.Fatal("The bundle file wasn't created.")
	}

	os.Remove(testBundleName)

	// Test alternate output name:
	err = bundle.Bundle("build", testBundleAltName, "", &force)

	_, err = os.Stat(testBundleAltName)
	if err != nil {
		t.Fatal("An alt bundle output name was specified, but the file wasn't found.")
	}

	os.Remove(testBundleAltName)
	os.Remove(testManifestName)
	os.Remove(testDummyFilename)
}
