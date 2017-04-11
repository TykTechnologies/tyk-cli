package importpkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/TykTechnologies/tyk-cli/request"
	"github.com/TykTechnologies/tyk-cli/utils"
)

var apiBody = `{
  "api_model": {},
      "api_definition": {
	"id": "54b53e47eba6db5c70000002",
	"name": "Nitrous Test",
	"api_id": "39d2c98be05c424371c600bd8b3e2242",
	"org_id": "54b53d3aeba6db5c35000002",
	"use_keyless": false,
	"use_oauth2": false,
	"oauth_meta": {
	  "allowed_access_types": [],
	  "allowed_authorize_types": [
	    "token"
	  ],
	  "auth_login_redirect": ""
	},
	"auth": {
	  "auth_header_name": "authorization"
	},
	"use_basic_auth": false,
	"notifications": {
	  "shared_secret": "",
	  "oauth_on_keychange_url": ""
	},
	"enable_signature_checking": false,
	"definition": {
	  "location": "header",
	  "key": ""
	},
	"version_data": {
	  "not_versioned": true,
	  "versions": {
	    "Default": {
	      "name": "Default",
	      "expires": "",
	      "paths": {
		"ignored": [],
		"white_list": [],
		"black_list": []
	      },
	      "use_extended_paths": false,
	      "extended_paths": {
		"ignored": [],
		"white_list": [],
		"black_list": []
	      }
	    }
	  }
	},
	"proxy": {
	  "listen_path": "/39d2c98be05c424371c600bd8b3e2242/",
	  "target_url": "http://tyk.io",
	  "strip_listen_path": true
	},
	"custom_middleware": {
	  "pre": null,
	  "post": null
	},
	"session_lifetime": 0,
	"active": true,
	"auth_provider": {
	  "name": "",
	  "storage_engine": "",
	  "meta": null
	},
	"session_provider": {
	  "name": "",
	  "storage_engine": "",
	  "meta": null
	},
	"event_handlers": {
	  "events": {}
	},
	"enable_batch_request_support": false,
	"enable_ip_whitelisting": false,
	"allowed_ips": [],
	"expire_analytics_after": 0
      },
      "hook_references": []
    }`

var newApiResponse = `{
      "Status": "OK"
      "Message": "API Created"
      "Meta": "5812"
    }`

func TestPostAPI(t *testing.T) {
	req, err := http.NewRequest("POST", "http://localhost:3000/api/apis", bytes.NewBuffer([]byte(apiBody)))
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	expectedResponse := newApiResponse
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(newApiResponse))
	})
	handler.ServeHTTP(recorder, req)

	call := request.New("12345", "http://localhost", "3000")
	var apiObj interface{}
	err = json.Unmarshal([]byte(apiBody), &apiObj)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	postAPI(apiObj.(map[string]interface{}), "/api/apis", call)
	if recorder.Body.String() != expectedResponse {
		t.Errorf(
			"Handler returned unexpected response.\nGot:\n\t%v\nExpected:\n\t%v",
			recorder.Body.String(),
			expectedResponse,
		)
	}
}

func TestGenereateAPIDef(t *testing.T) {
	call := request.New("12345", "http://localhost", "3000")
	var apiObj interface{}
	err := json.Unmarshal([]byte(apiBody), &apiObj)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	apis := utils.MapToIntfSlice(apiObj.(map[string]interface{}), "apis")
	generateAPIDef(apis, call)
}

func TestAPIs(t *testing.T) {
	args := []string{"12345", "http://localhost", "3000", "./api_test.json"}
	APIs(args)
}
