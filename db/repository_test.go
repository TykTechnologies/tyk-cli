package db

import (
	"testing"
)

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

// TODO
//func TestAddRecord(t *testing.T){}
