package api

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	uuid "github.com/satori/go.uuid"

	"github.com/TykTechnologies/tyk-cli/db"
	"github.com/TykTechnologies/tyk-cli/utils"
	"github.com/TykTechnologies/tyk/apidef"
)

type APIDef struct {
	id            string
	name          string
	item          db.Item
	APIModel      ApiModel
	APIDefinition apidef.APIDefinition
}

type ApiModel struct {
	schemaPath string
}

func (api *APIDef) setAPIDefinition() {
	api.APIDefinition.APIID = api.Id()
	api.APIDefinition.Name = api.Name()
}

func New(name string) *APIDef {
	api := APIDef{}
	api.id = strings.Replace(uuid.NewV4().String(), "-", "", -1)
	api.name = name
	return &api
}

func (api *APIDef) Id() string {
	return api.id
}

func (api *APIDef) Name() string {
	return api.name
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
		APIModel      ApiModel             `json:"api_model"`
		APIDefinition apidef.APIDefinition `json:"api_definition"`
	}
	return rec{api.APIModel, api.APIDefinition}
}

// Create is a public function for creating staged APIs
func (api *APIDef) Create() error {
	db_file := filepath.Join("db", "bolt.db")
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
	db_file := filepath.Join("db", "bolt.db")
	bdb, err := bolt.Open(db_file, 0666, &bolt.Options{ReadOnly: true})
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
