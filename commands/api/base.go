package api

import (
	"encoding/json"
	"strings"

	"github.com/TykTechnologies/tyk-cli/db"
	"github.com/TykTechnologies/tyk-cli/utils"
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/boltdb/bolt"
	uuid "github.com/satori/go.uuid"
)

type APIDef struct {
	Id            string
	Name          string
	item          db.Item
	APIModel      interface{}
	APIDefinition apidef.APIDefinition
}

func (api *APIDef) setAPIDefinition() {
	api.APIDefinition.APIID = api.GetId()
	api.APIDefinition.Name = api.GetName()
}

func New(name string) *APIDef {
	api := APIDef{}
	api.Id = strings.Replace(uuid.NewV4().String(), "-", "", -1)
	api.Name = name
	return &api
}

func (api *APIDef) GetId() string {
	return api.Id
}

func (api *APIDef) GetName() string {
	return api.Name
}

func (api *APIDef) BucketName() string {
	return "apis"
}

func (api *APIDef) Group() string {
	return "API"
}

func (api *APIDef) GetRecordData() interface{} {
	api.setAPIDefinition()
	type rec struct {
		APIModel      interface{}
		APIDefinition apidef.APIDefinition
	}
	return rec{api.APIModel, api.APIDefinition}
}

// Create is a public function for creating staged APIs
func (api *APIDef) Create() error {
	db_file := "./db/bolt.db"
	utils.MkdirPFile(db_file)
	bdb, err := bolt.Open(db_file, 0600, nil)
	utils.HandleError(err, true)
	defer bdb.Close()
	err = bdb.Update(func(tx *bolt.Tx) error {
		return db.AddRecord(tx, api)
	})
	return utils.ReturnErr(err)
}

// Find is a public function for finding staged APIs
func (apis *APIDef) Find(id string) (interface{}, error) {
	bdb, err := bolt.Open("./db/bolt.db", 0666, &bolt.Options{ReadOnly: true})
	utils.HandleError(err, true)
	defer bdb.Close()
	var item interface{}

	err = bdb.View(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte(apis.BucketName()))
		member := collection.Get([]byte(id))
		err := json.Unmarshal(member, &item)
		if err != nil {
			return err
		}
		return nil
	})
	return item, nil
}
