package bundle

import (
	"encoding/json"
	"fmt"
	"github.com/TykTechnologies/tyk/apidef"
	"io/ioutil"
	"os"
	"testing"
)

const (
	testManifestName  = "manifest.json"
	testPerm          = 0755
	testBundleName    = "zip"
	testBundleAltName = "myzip"
	testDummyFilename = "mymiddleware.py"
)

func TestBundleUnknownCommand(t *testing.T) {
	fmt.Println("a test", t)
	var force bool
	var err error
	err = Bundle("unknown", "", "", &force)
	if err == nil {
		t.Fatal("Must return an error when the command doesn't exist.")
	}
}

func TestBundleBuildCommand(t *testing.T) {
	var force bool = true
	var err error
	err = Bundle("build", "", "", &force)
	if err == nil {
		t.Fatal("The manifest file doesn't exist, the build command should return an error.")
	}

	noManifest := []byte("")
	ioutil.WriteFile(testManifestName, noManifest, testPerm)

	err = Bundle("build", "", "", &force)
	if err == nil {
		t.Fatal("When the manifest is invalid, the build command should return an error.")
	}

	os.Remove(testManifestName)

	emptyManifest := []byte("{}")
	ioutil.WriteFile(testManifestName, emptyManifest, testPerm)

	err = Bundle("build", "", "", &force)
	if err == nil {
		t.Fatal("The manifest doesn't have any hooks, the build command should return an error.")
	}

	os.Remove(testManifestName)
}

func TestBundleValidateManifest(t *testing.T) {
	var manifest apidef.BundleManifest
	var err error
	err = BundleValidateManifest(&manifest)
	if err == nil {
		t.Fatal("The manifest is empty (no hooks defined), the validation step should return an error.")
	}

	manifest = apidef.BundleManifest{
		FileList: []string{"nonexistentfile"},
	}
	err = BundleValidateManifest(&manifest)
	if err == nil {
		t.Fatal("The manifest file references a nonexistent file, the validation step should return an error.")
	}

	manifest = apidef.BundleManifest{
		CustomMiddleware: apidef.MiddlewareSection{
			Pre: []apidef.MiddlewareDefinition{
				apidef.MiddlewareDefinition{},
			},
		},
	}

	err = BundleValidateManifest(&manifest)

	if err == nil {
		t.Fatal("The validation step must return an error when no driver is specified.")
	}

	manifest = apidef.BundleManifest{
		CustomMiddleware: apidef.MiddlewareSection{
			Pre: []apidef.MiddlewareDefinition{
				apidef.MiddlewareDefinition{
					Name: "mymiddleware",
					Path: "./mymiddleware.py",
				},
			},
			Driver: apidef.OttoDriver,
		},
	}

	err = BundleValidateManifest(&manifest)
	if err != nil {
		t.Fatal("A minimal manifest should be valid.")
	}
}

func TestBundleBasicBuild(t *testing.T) {
	var force bool = true
	var err error
	var testManifest apidef.BundleManifest
	testManifest = apidef.BundleManifest{
		FileList: []string{"mymiddleware.py"},
		CustomMiddleware: apidef.MiddlewareSection{
			Pre: []apidef.MiddlewareDefinition{
				apidef.MiddlewareDefinition{
					Name: "mymiddleware",
					Path: "./mymiddleware.py",
				},
			},
			Driver: apidef.OttoDriver,
		},
	}

	ioutil.WriteFile(testDummyFilename, []byte(""), testPerm)

	var testManifestData []byte
	testManifestData, err = json.Marshal(&testManifest)
	if err != nil {
		t.Fatal("Couldn't marshal the test manifest.")
	}

	ioutil.WriteFile(testManifestName, testManifestData, testPerm)

	err = Bundle("build", "", "", &force)

	_, err = os.Stat(testBundleName)
	if err != nil {
		t.Fatal("The file wasn't created.")
	}

	os.Remove(testBundleName)

	// Test alternate output name:
	err = Bundle("build", testBundleAltName, "", &force)

	_, err = os.Stat(testBundleAltName)
	if err != nil {
		t.Fatal("An alt output name was specified, but the file wasn't found.")
	}

	os.Remove(testBundleAltName)
	os.Remove(testManifestName)
	os.Remove(testDummyFilename)
}
