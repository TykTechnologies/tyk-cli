package db

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"

	"github.com/TykTechnologies/tyk-cli/utils"
)

type Item struct {
	id   string
	name string
}

// Record interface for all objects in the DB
type Record interface {
	Id() string
	Name() string
	BucketName() string
	Group() string
	RecordData() interface{}
	Create(bdb *bolt.DB) error
	Edit(bdb *bolt.DB, params map[string]interface{}) error
}

func (item *Item) Id() string {
	return item.id
}

func (item *Item) Name() string {
	return item.name
}

func (item *Item) BucketName() string {
	return "items"
}

func (item *Item) Group() string {
	return "Item"
}

func (item *Item) RecordData() interface{} {
	return item
}

func (item *Item) Create(bdb *bolt.DB) error {
	log.Fatal("Please implement me")
	return nil
}

// Find is a public function for finding database objects
func Find(bdb *bolt.DB, id string) (interface{}, error) {
	log.Fatal("Please implement me")
	return nil, nil
}

// Edit is a public function for editing database objects
func (item *Item) Edit(bdb *bolt.DB, params map[string]interface{}) error {
	log.Fatal("Please implement me")
	return nil
}

// OpenDB is a public function used to open the Database
func OpenDB(filename string, permission os.FileMode, readOnly bool) (*bolt.DB, error) {
	dbFile := filepath.Join(
		os.Getenv("GOPATH"),
		"src",
		"github.com",
		"TykTechnologies",
		"tyk-cli",
		"db",
		filename,
	)
	utils.MkdirPFile(dbFile)
	options := &bolt.Options{ReadOnly: readOnly}

	bdb, err := bolt.Open(dbFile, permission, options)
	return bdb, err
}

// AddRecord function adds records to BoltDB
func AddRecord(collection *bolt.Bucket, r Record) error {
	member, err := json.Marshal(r.RecordData())
	if err != nil {
		log.Fatal(err)
	}
	return collection.Put([]byte(r.Id()), member)
}
