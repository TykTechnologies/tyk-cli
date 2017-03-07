package bundle

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/TykTechnologies/tyk/apidef"
)

const (
	testManifestName  = "./manifest.json"
	testPerm          = 0755
	testBundleName    = "./bundle.zip"
	testBundleAltName = "./mybundle.zip"
	testDummyFilename = "./mymiddleware.py"
)

var (
	force bool
	err   error
)

func TestBundleUnknownCommand(t *testing.T) {
	force = true
	err = Bundle("unknown", "", "", &force)
	if err == nil {
		t.Fatal("Must return an error when the command doesn't exist.")
	}
}

func TestBundleBuildWithoutManifest(t *testing.T) {
	force = true
	err = Bundle("build", "", "", &force)
	if err == nil {
		t.Fatal("The manifest file doesn't exist, the build command should return an error.")
	}
}

func TestBundleBuildWithInvalidManifest(t *testing.T) {
	force = true
	noManifest := []byte("")
	ioutil.WriteFile(testManifestName, noManifest, testPerm)
	err = Bundle("build", "", "", &force)
	if err == nil {
		t.Fatal("When the manifest is invalid, the build command should return an error.")
	}
	deleteFiles([]string{testManifestName})
}

func TestBundleBuildWithEmptyManifest(t *testing.T) {
	force = true
	emptyManifest := []byte("{}")
	ioutil.WriteFile(testManifestName, emptyManifest, testPerm)
	err = Bundle("build", "", "", &force)
	if err == nil {
		t.Fatal("The manifest doesn't have any hooks, the build command should return an error.")
	}
	deleteFiles([]string{testManifestName})
}

func TestBundleValidateManifestWithEmptyFile(t *testing.T) {
	var manifest apidef.BundleManifest
	err = BundleValidateManifest(&manifest)
	if err == nil {
		t.Fatal("The manifest is empty (no hooks defined), the validation step should return an error.")
	}
}

func TestBundleValidateManifestWithNonExistentFile(t *testing.T) {
	manifest := apidef.BundleManifest{
		FileList: []string{"nonexistentfile"},
	}
	err = BundleValidateManifest(&manifest)
	if err == nil {
		t.Fatal("The manifest file references a nonexistent file, the validation step should return an error.")
	}
}

func TestBundleValidateManifestWithNoDriver(t *testing.T) {
	manifest := apidef.BundleManifest{
		CustomMiddleware: apidef.MiddlewareSection{
			Pre: []apidef.MiddlewareDefinition{
				{},
			},
		},
	}
	err = BundleValidateManifest(&manifest)
	if err == nil {
		t.Fatal("The validation step must return an error when no driver is specified.")
	}
}

func TestBundleValidateManifestMinimalFileValidation(t *testing.T) {
	manifest := validManifest()
	err = BundleValidateManifest(&manifest)
	if err != nil {
		t.Fatal("A minimal manifest should be valid.")
	}
}

func TestBundleBasicBuild(t *testing.T) {
	force = true
	testManifest := validManifest()
	testManifestData, err := json.Marshal(&testManifest)
	if err != nil {
		t.Fatalf("Couldn't marshal the test manifest.")
	}
	ioutil.WriteFile(testManifestName, testManifestData, testPerm)
	err = Bundle("build", testBundleName, "", &force)
	_, err = os.Stat(testBundleName)
	if err != nil {
		t.Fatalf("The file wasn't created.\nERROR=%v", err)
	}
	deleteFiles([]string{testBundleName, testManifestName, testDummyFilename})
}

func TestBundleBasicBuildAlternateOutputName(t *testing.T) {
	force = true
	testManifest := validManifest()
	ioutil.WriteFile(testDummyFilename, []byte(""), testPerm)
	testManifestData, err := json.Marshal(&testManifest)
	if err != nil {
		t.Fatalf("Couldn't marshal the test manifest.")
	}
	ioutil.WriteFile(testManifestName, testManifestData, testPerm)
	err = Bundle("build", testBundleAltName, "", &force)
	_, err = os.Stat(testBundleAltName)
	if err != nil {
		t.Fatal("An alt output name was specified, but the file wasn't found.")
	}
	deleteFiles([]string{testBundleAltName, testManifestName, testDummyFilename})
}

func validManifest() apidef.BundleManifest {
	return apidef.BundleManifest{
		CustomMiddleware: apidef.MiddlewareSection{
			Pre: []apidef.MiddlewareDefinition{
				{
					Name: "mymiddleware",
					Path: "./mymiddleware.py",
				},
			},
			Driver: apidef.OttoDriver,
		},
	}
}

func deleteFiles(files []string) {
	for i := range files {
		os.Remove(files[i])
	}
}
