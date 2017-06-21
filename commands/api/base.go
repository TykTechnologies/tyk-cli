package api

import (
	"encoding/json"
	"fmt"
	"log"
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
	APIModel      APIModel             `json:"api_model"`
	APIDefinition apidef.APIDefinition `json:"api_definition"`
}

type APIModel struct {
	schemaPath string
}

func (api *APIDef) SetAPIDefinition() {
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

func (api *APIDef) RecordData() interface{} {
	api.SetAPIDefinition()
	type rec struct {
		APIModel      APIModel             `json:"api_model"`
		APIDefinition apidef.APIDefinition `json:"api_definition"`
	}
	return rec{api.APIModel, api.APIDefinition}
}

// Create is a public function for creating staged APIs
func (api *APIDef) Create(bdb *bolt.DB) error {
	return bdb.Update(func(tx *bolt.Tx) error {
		collection, err := tx.CreateBucketIfNotExists([]byte(api.BucketName()))
		if err != nil {
			log.Fatal(err)
		}
		return db.AddRecord(collection, api)
	})
}

// Find is a public function for finding staged APIs
func Find(bdb *bolt.DB, id string) (APIDef, error) {
	var api APIDef
	err := bdb.View(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte("apis"))
		member := collection.Get([]byte(id))
		if member == nil {
			return fmt.Errorf("API not found")
		}
		loader := APIDefinitionLoader{}
		_, m := loader.ParseDefinition([]byte(member))
		b, _ := json.Marshal(m["api_definition"])
		def, _ := loader.ParseDefinition(b)
		api.APIDefinition = *def
		return nil
	})
	return api, err
}

// Edit is a public function for editing staged APIs
func (api *APIDef) Edit(bdb *bolt.DB, params map[string]interface{}) error {
	id := api.APIDefinition.APIID
	var apiInt interface{}
	m, err := json.Marshal(api)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(m, &apiInt)
	if err != nil {
		log.Fatal(err)
	}
	api.attributes(params)
	err = bdb.Update(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte(api.BucketName()))
		return collection.Delete([]byte(id))
	})
	if err != nil {
		log.Fatal(err)
	}
	return bdb.Update(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte(api.BucketName()))
		return db.AddRecord(collection, api)
	})
}

func (api *APIDef) attributes(params map[string]interface{}) {
	var originalDef map[string]interface{}
	m, err := json.Marshal(api.APIDefinition)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(m, &originalDef)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range originalDef {
		if params[k] == nil {
			params[k] = v
		}
	}
	id := api.APIDefinition.APIID
	loader := APIDefinitionLoader{}
	b, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}
	def, _ := loader.ParseDefinition(b)
	api.APIDefinition = *def
	if params["names"] != nil {
		api.name = params["name"].(string)
	} else {
		api.name = api.APIDefinition.Name
	}
	api.id = id
}

// Delete is a public function for deleting staged API
func (api *APIDef) Delete(bdb *bolt.DB) error {
	return bdb.Update(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte(api.BucketName()))
		return collection.Delete([]byte(api.Id()))
	})
}

// TODO - taken from the Main Tyk package
type APIDefinitionLoader struct{}

func (a *APIDefinitionLoader) ParseDefinition(apiDef []byte) (*apidef.APIDefinition, map[string]interface{}) {
	appConfig := &apidef.APIDefinition{}
	if err := json.Unmarshal(apiDef, appConfig); err != nil {
		log.Fatal("[RPC] --> Couldn't unmarshal api configuration: ", err)
	}
	// Got the structured version - now lets get a raw copy for modules
	rawConfig := make(map[string]interface{})
	json.Unmarshal(apiDef, &rawConfig)

	return appConfig, rawConfig
}
