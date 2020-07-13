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
	id, _ := uuid.NewV4()
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

type rec struct {
	APIModel      APIModel             `json:"api_model"`
	APIDefinition apidef.APIDefinition `json:"api_definition"`
}

func (api *APIDef) RecordData() interface{} {
	api.SetAPIDefinition()
	return rec{api.APIModel, api.APIDefinition}
}

func (api *APIDef) define(rawDef map[string]interface{}) error {
	b, err := json.Marshal(rawDef)
	if err != nil {
		return err
	}
	def, _ := parseDefinition(b)
	api.APIDefinition = *def
	return nil
}

// Create is a public function for creating staged APIs
func (api *APIDef) Create(bdb *bolt.DB) error {
	return bdb.Update(func(tx *bolt.Tx) error {
		collection, err := tx.CreateBucketIfNotExists([]byte(api.BucketName()))
		if err != nil {
			return err
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
		_, m := parseDefinition([]byte(member))
		err := api.define(m["api_definition"].(map[string]interface{}))
		if err != nil {
			return err
		}
		return nil
	})
	return api, err
}

// FindByName is a function for finding APIs with a specific name
func FindByName(bdb *bolt.DB, name string) ([]APIDef, error) {
	var list []APIDef
	err := bdb.View(func(tx *bolt.Tx) error {
		var api APIDef
		c := tx.Bucket([]byte("apis"))
		c.ForEach(func(k, v []byte) error {
			var rawDef map[string]interface{}
			err := json.Unmarshal(v, &rawDef)
			if err != nil {
				log.Fatal(err)
			}
			switch name {
			case "":
				err := api.define(rawDef["api_definition"].(map[string]interface{}))
				if err != nil {
					return err
				}
				list = append(list, api)

			default:
				if rawDef["api_definition"].(map[string]interface{})["name"].(string) == name {
					err := api.define(rawDef["api_definition"].(map[string]interface{}))
					if err != nil {
						return err
					}
					list = append(list, api)
				}
			}
			return nil
		})
		return nil
	})
	return list, err
}

// FindByOrgID is a function for listing all APIs in a given organisation
func FindByOrgID(bdb *bolt.DB, orgID string) ([]APIDef, error) {
	var list []APIDef
	err := bdb.View(func(tx *bolt.Tx) error {
		var api APIDef
		c := tx.Bucket([]byte("apis"))
		c.ForEach(func(k, v []byte) error {
			var rawDef map[string]interface{}
			err := json.Unmarshal(v, &rawDef)
			if err != nil {
				log.Fatal(err)
			}
			switch orgID {
			case "":
				err := api.define(rawDef["api_definition"].(map[string]interface{}))
				if err != nil {
					return err
				}
				list = append(list, api)

			default:
				if rawDef["api_definition"].(map[string]interface{})["org_id"].(string) == orgID {
					err := api.define(rawDef["api_definition"].(map[string]interface{}))
					if err != nil {
						return err
					}
					list = append(list, api)
				}
			}
			return nil
		})
		return nil
	})
	return list, err
}

// Edit is a public function for editing staged APIs
func (api *APIDef) Edit(bdb *bolt.DB, params map[string]interface{}) error {
	id := api.APIDefinition.APIID
	var apiInt interface{}
	m, err := json.Marshal(api)
	if err != nil {
		return err
	}
	err = json.Unmarshal(m, &apiInt)
	if err != nil {
		return err
	}
	err = api.attributes(params)
	if err != nil {
		return err
	}
	err = bdb.Update(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte(api.BucketName()))
		return collection.Delete([]byte(id))
	})
	if err != nil {
		return err
	}
	return bdb.Update(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte(api.BucketName()))
		return db.AddRecord(collection, api)
	})
}

func (api *APIDef) attributes(params map[string]interface{}) error {
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
	err = api.define(params)
	if err != nil {
		return err
	}
	if params["names"] != nil {
		api.name = params["name"].(string)
	} else {
		api.name = api.APIDefinition.Name
	}
	api.id = id
	return nil
}

// Delete is a public function for deleting staged API
func (api *APIDef) Delete(bdb *bolt.DB) error {
	return bdb.Update(func(tx *bolt.Tx) error {
		collection := tx.Bucket([]byte(api.BucketName()))
		return collection.Delete([]byte(api.Id()))
	})
}

// DeleteAll is a public function for deleting all staged APIs
func DeleteAll(bdb *bolt.DB) error {
	return bdb.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("apis"))
	})
}

func parseDefinition(apiDef []byte) (*apidef.APIDefinition, map[string]interface{}) {
	appConfig := &apidef.APIDefinition{}
	if err := json.Unmarshal(apiDef, appConfig); err != nil {
		log.Fatal("[RPC] --> Couldn't unmarshal api configuration: ", err)
	}
	// Got the structured version - now lets get a raw copy for modules
	rawConfig := make(map[string]interface{})
	json.Unmarshal(apiDef, &rawConfig)

	return appConfig, rawConfig
}
