package api

import (
	"github.com/TykTechnologies/tyk-cli/db"
)

type APIDef struct {
	db.Item
	db.Bucket
}

func New() APIDef {
	api := APIDef{}
	api.BucketName = "apis"
	api.Group = "API"
	return api
}
