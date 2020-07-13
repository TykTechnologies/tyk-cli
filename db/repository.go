package db

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	uuid "github.com/satori/go.uuid"

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
	Delete(bdb *bolt.DB) error
}

func New(name string) *Item {
	item := Item{}
	item.id = strings.Replace(uuid.NewV4().String(), "-", "", -1)
	item.name = name
	return &item
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

func (item *Item) Create(bdb *bolt.DB) error {
	log.Print("Please implement me")
	return nil
}

// Find is a public function for finding database objects
func Find(bdb *bolt.DB, id string) (interface{}, error) {
	log.Print("Please implement me")
	return nil, nil
}

// Edit is a public function for editing database objects
func (item *Item) Edit(bdb *bolt.DB, params map[string]interface{}) error {
	log.Print("Please implement me")
	return nil
}

// Delete is a public function for deleting database objects
func (item *Item) Delete(bdb *bolt.DB) error {
	log.Print("Please implement me")
	return nil
}

// DeleteAll is a public function for deleting all staged items
func DeleteAll(bdb *bolt.DB) error {
	return bdb.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("items"))
	})
}
