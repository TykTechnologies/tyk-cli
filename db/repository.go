package db

import (
	"encoding/json"
	"log"

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
	GetRecordData() interface{}
	Create() error
	Find(id string) (interface{}, error)
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

func (item *Item) GetRecordData() interface{} {
	return item
}

func (item *Item) Create() error {
	log.Fatal("Please implement me")
	return nil
}

// Find is a public function for finding staged APIs
func (item *Item) Find(id string) (interface{}, error) {
	log.Fatal("Please implement me")
	return interface{}(nil), nil
}

// AddRecord function adds records to BoltDB
func AddRecord(tx *bolt.Tx, r Record) error {
	collection, err := tx.CreateBucketIfNotExists([]byte(r.BucketName()))
	utils.HandleError(err, true)
	member, err := json.Marshal(r.GetRecordData())
	utils.HandleError(err, true)
	return collection.Put([]byte(r.Id()), member)
}
