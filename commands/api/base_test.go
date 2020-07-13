package api

import (
	"log"
	"reflect"
	"testing"

	"github.com/TykTechnologies/tyk-cli/db"
)

func TestId(t *testing.T) {
	test := New("test")
	expected := test.id
	result := test.Id()
	if result != expected {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestName(t *testing.T) {
	test := New("test")
	expected := test.name
	result := test.Name()
	if result != expected {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestBucketName(t *testing.T) {
	test := New("test")
	expected := "apis"
	result := test.BucketName()
	if result != expected {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestGroup(t *testing.T) {
	test := New("test")
	expected := "API"
	result := test.Group()
	if result != expected {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

type setAPIDefinitionTest struct {
	expected string
	result   string
}

func TestSetAPIDefinition(t *testing.T) {
	testAPI := New("test")
	testAPI.SetAPIDefinition()
	tests := []setAPIDefinitionTest{
		{testAPI.id, testAPI.APIDefinition.APIID},
		{testAPI.name, testAPI.APIDefinition.Name},
	}
	for _, i := range tests {
		if i.expected != i.result {
			t.Fatalf("expected %v, got %v", i.expected, i.result)
		}
	}
}

func TestRecordData(t *testing.T) {
	test1 := New("test")
	test1.SetAPIDefinition()
	test2 := New("test")
	test2.id = test1.Id()
	expected := rec{test1.APIModel, test1.APIDefinition}
	result := test2.RecordData()
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestCreate(t *testing.T) {
	err := dropDB("test.db")
	if err != nil {
		log.Print(err)
	}
	bdb, err := db.OpenDB("test.db", 0600, false)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()
	apiName := "New API name"
	test := New(apiName)
	err = test.Create(bdb)
	if err != nil {
		t.Fatal(err)
	}
	apis, err := FindByName(bdb, apiName)
	if err != nil {
		log.Println(err)
	}
	if len(apis) == 0 {
		t.Fatal(`Error: API was not created`)
	}
	api := apis[0]
	if api.APIDefinition.Name != apiName || api.APIDefinition.APIID == "" {
		t.Fatal(`Error: API was not created`)
	}
}

func TestCreateError(t *testing.T) {
	err := dropDB("test.db")
	if err != nil {
		log.Print(err)
	}
	bdb, err := db.OpenDB("test.db", 0660, false)
	if err != nil {
		log.Print(err)
	}
	bdb.Close()
	apiName := "New API name"
	test := New(apiName)
	err = test.Create(bdb)
	if err == nil {
		t.Fatal(`Error: API was created`)
	}
}

func TestFind(t *testing.T) {
	err := dropDB("test.db")
	if err != nil {
		log.Print(err)
	}
	bdb, err := db.OpenDB("test.db", 0600, false)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()
	apiName := "New API name"
	test := New(apiName)
	test.id = "1234567890abcdef"
	err = test.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	api, err := Find(bdb, test.id)
	if err != nil {
		log.Println(err)
	}
	if api.APIDefinition.Name != apiName || api.APIDefinition.APIID == "" {
		t.Fatal(`Error: Wrong API found`)
	}
}

func TestFindMissing(t *testing.T) {
	err := dropDB("test.db")
	if err != nil {
		log.Print(err)
	}
	bdb, err := db.OpenDB("test.db", 0600, false)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()
	apiName := "New API name"
	test := New(apiName)
	id := "1234567890abcdef"
	err = test.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	_, err = Find(bdb, id)
	if err == nil {
		t.Fatal("Error: Should have raised error on missing API")
	}
	if err.Error() != "API not found" {
		t.Fatalf("Expected error 'API not found', got '%v'", err)
	}
}

func TestFindByOrgID(t *testing.T) {
	err := dropDB("test.db")
	if err != nil {
		log.Print(err)
	}
	bdb, err := db.OpenDB("test.db", 0600, false)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()
	apiName := "New API name"
	test := New(apiName)
	err = test.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	orgID := "12345678901234567890"
	err = test.Edit(bdb, map[string]interface{}{"org_id": orgID})
	if err != nil {
		t.Fatalf("Error: API was not updated\n%v\n", err)
	}
	_, err = FindByOrgID(bdb, orgID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFindByOrgIDWildcard(t *testing.T) {
	err := dropDB("test.db")
	if err != nil {
		log.Print(err)
	}
	bdb, err := db.OpenDB("test.db", 0600, false)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()
	test1 := New("test 1")
	test2 := New("test 2")
	err = test1.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	err = test1.Edit(bdb, map[string]interface{}{"org_id": "12345678901234567890"})
	if err != nil {
		t.Fatalf("Error: API was not updated\n%v\n", err)
	}
	err = test2.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	orgID := ""
	apis, err := FindByOrgID(bdb, orgID)
	if err != nil {
		t.Fatal(err)
	}
	if len(apis) != 2 {
		t.Fatalf("Error: expected 2 APIs, got %v", len(apis))
	}
}

func TestFindByName(t *testing.T) {
	err := dropDB("test.db")
	if err != nil {
		log.Print(err)
	}
	bdb, err := db.OpenDB("test.db", 0600, false)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()
	apiName := "New API name"
	test := New(apiName)
	err = test.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	_, err = FindByName(bdb, apiName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFindByNameWildcard(t *testing.T) {
	err := dropDB("test.db")
	if err != nil {
		log.Print(err)
	}
	bdb, err := db.OpenDB("test.db", 0600, false)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()
	test1 := New("test 1")
	test2 := New("test 2")
	err = test1.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	err = test2.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	apis, err := FindByName(bdb, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(apis) != 2 {
		t.Fatalf("Error: expected 2 APIs, got %v", len(apis))
	}
}

func TestDelete(t *testing.T) {
	err := dropDB("test.db")
	if err != nil {
		log.Print(err)
	}
	bdb, err := db.OpenDB("test.db", 0600, false)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()
	test1 := New("test 1")
	test2 := New("test 2")
	err = test1.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	err = test2.Create(bdb)
	if err != nil {
		t.Fatalf("Error: API was not created\n%v\n", err)
	}
	test1.Delete(bdb)
	apis, err := FindByName(bdb, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(apis) != 1 {
		t.Fatalf("Error: expected to find 1 API, got %v", len(apis))
	}
}

func dropDB(filename string) error {
	bdb, err := db.OpenDB("test.db", 0777, false)
	if err != nil {
		return err
	}
	defer bdb.Close()
	err = DeleteAll(bdb)
	if err != nil {
		return err
	}
	return nil
}
