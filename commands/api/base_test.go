package api

import (
	"github.com/TykTechnologies/tyk-cli/db"
	"testing"
)

func TestCreate(t *testing.T) {
	apiId := "1234567890abcdef"
	API := New()
	API.Create(&db.Item{apiId, "api_name"})

	if API.Find(apiId) == nil {
		t.Fatal(`Error: API was not created`)
	}
}
