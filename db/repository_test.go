package db

import (
	"reflect"
	"testing"
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
	expected := "items"
	result := test.BucketName()
	if result != expected {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestGroup(t *testing.T) {
	test := New("test")
	expected := "Item"
	result := test.Group()
	if result != expected {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestRecordData(t *testing.T) {
	test1 := New("test")
	if !reflect.DeepEqual(test1.RecordData(), test1) {
		t.Fatalf("expected %v, got %v", test1, test1.RecordData())
	}
}

func TestOpenDB(t *testing.T) {
	bdb, err := OpenDB("test.db", 0666, false)
	if err != nil {
		t.Fatal(err)
	} else if bdb == nil {
		t.Fatal("expected db")
	}

	if bdb.IsReadOnly() {
		bdb.Close()
		t.Fatal("expected db to be writable")
	}

	if err = bdb.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestOpenDBIfReadOnly(t *testing.T) {
	dropDB("test.db")
	bdb, err := OpenDB("test.db", 0666, true)
	if err != nil {
		t.Fatal(err)
	} else if bdb == nil {
		t.Fatal("expected db")
	}

	if !bdb.IsReadOnly() {
		bdb.Close()
		t.Fatal("expected db to be read-only")
	}
	if err = bdb.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate(t *testing.T) {
	dropDB("test.db")
	test := New("test")
	bdb, err := OpenDB("test.db", 0666, false)
	if err != nil {
		t.Fatal(err)
	}
	defer bdb.Close()
	err = test.Create(bdb)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFind(t *testing.T) {
	dropDB("test.db")
	test := New("test")
	test.id = "1234567890"
	bdb, err := OpenDB("test.db", 0666, false)
	if err != nil {
		t.Fatal(err)
	}
	defer bdb.Close()
	err = test.Create(bdb)
	if err != nil {
		t.Fatal(err)
	}
	items, err := Find(bdb, test.id)
	if err != nil {
		t.Fatal(err)
	}
	if items != nil {
		t.Fatalf("Expected to find no items, found %v", items)
	}
}

func TestEdit(t *testing.T) {
	dropDB("test.db")
	test := New("test")
	bdb, err := OpenDB("test.db", 0666, false)
	if err != nil {
		t.Fatal(err)
	}
	defer bdb.Close()
	err = test.Create(bdb)
	if err != nil {
		t.Fatal(err)
	}
	err = test.Edit(bdb, map[string]interface{}{"dummy": "input"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDelete(t *testing.T) {
	dropDB("test.db")
	test := New("test")
	bdb, err := OpenDB("test.db", 0666, false)
	if err != nil {
		t.Fatal(err)
	}
	defer bdb.Close()
	err = test.Create(bdb)
	if err != nil {
		t.Fatal(err)
	}
	err = test.Delete(bdb)
	if err != nil {
		t.Fatal(err)
	}
}

func dropDB(filename string) error {
	bdb, err := OpenDB("test.db", 0777, false)
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
