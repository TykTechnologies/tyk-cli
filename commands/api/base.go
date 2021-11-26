package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	uuid "github.com/satori/go.uuid"

	"github.com/TykTechnologies/tyk-cli/db"
	"github.com/TykTechnologies/tyk/apidef"
)

type APIDef struct {
	id            string
	name          string
	item          db.Item
	APIModel      APIModel
	APIDefinition apidef.APIDefinition
}

type APIModel struct {
	schemaPath string
}

func (api *APIDef) setAPIDefinition() {
	api.APIDefinition.APIID = api.Id()
	api.APIDefinition.Name = api.Name()
}

func New(name string) *APIDef {
	api := APIDef{}
	id := uuid.NewV4()
	api.id = strings.Replace(id.String(), "-", "", -1)
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

func (api *APIDef) RecordData() interface{} {
	api.setAPIDefinition()
	type rec struct {
		APIModel      APIModel             `json:"api_model"`
		APIDefinition apidef.APIDefinition `json:"api_definition"`
	}
	return rec{api.APIModel, api.APIDefinition}
}

// Create is a public function for creating staged APIs
func (api *APIDef) Create(bdb *bolt.DB) error {
	return bdb.Update(func(tx *bolt.Tx) error {
		return db.AddRecord(tx, api)
	})
}

// Find is a public function for finding staged APIs
func (apis *APIDef) Find(bdb *bolt.DB, id string) (interface{}, error) {
	var item interface{}
	err := bdb.View(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte(apis.BucketName()))
		member := collection.Get([]byte(id))
		return json.Unmarshal(member, &item)
	})
	if item == nil {
		return nil, fmt.Errorf("API not found")
	}
	return item, err
}
