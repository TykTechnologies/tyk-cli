package db

import (
	"encoding/json"
	"fmt"

	"github.com/TykTechnologies/tyk-cli/utils"
	"github.com/boltdb/bolt"
)

type Bucket struct {
	BucketName string
	Group      string
}

type Item struct {
	Id   string
	Name string
}

// Record interface for all objects in the DB
type Record interface {
	Create(item *Item) error
	Find(id string)
}

// Create is a public function for creating staged APIs
func (b *Bucket) Create(item *Item) error {
	db_file := "./db/bolt.db"
	utils.MkdirPFile(db_file)
	bdb, err := bolt.Open(db_file, 0600, nil)
	utils.LogErr(err)
	defer bdb.Close()

	err = bdb.Update(func(tx *bolt.Tx) error {
		collection, err := tx.CreateBucketIfNotExists([]byte(b.BucketName))
		utils.ReturnErr(err)
		member, err := json.Marshal(item)
		utils.ReturnErr(err)
		return collection.Put([]byte(item.Id), member)
	})
	fmt.Printf("%v %v created ID %v\n", b.Group, item.Name, item.Id)
	return utils.ReturnErr(err)
}

// Find is a public function for finding staged APIs
func (b *Bucket) Find(id string) []byte {
	bdb, err := bolt.Open("./db/bolt.db", 0666, &bolt.Options{ReadOnly: true})
	utils.LogErr(err)
	defer bdb.Close()
	var member []byte

	err = bdb.View(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte(b.BucketName))
		member = collection.Get([]byte(id))
		fmt.Println(member)
		return nil
	})
	return member
}
