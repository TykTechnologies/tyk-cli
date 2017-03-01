package request

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var responseBody = `{
  "apis": [
    {
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
    }
  ],
  "pages": 0
}`

func requestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Authorization", r.Header["Authorization"][0])
	w.Header().Set("Content-Type", r.Header["Content-Type"][0])
	w.Write([]byte(responseBody))
}

func TestFullRequest(t *testing.T) {
	request := New("12345", "http://www.example.com", "3000")
	req, err := request.FullRequest("GET", "http://localhost:3000/api/apis", nil)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(requestHandler)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf(
			"handler returned wrong status code: got %v want %v",
			status, http.StatusOK,
		)
	}

	if recorder.Body.String() != responseBody {
		t.Errorf(
			"handler returned unexpected body.\nGot:\n\t%v\nExpected:\n\t%v",
			recorder.Body.String(),
			responseBody,
		)
	}
}

func TestFullRequestAuthorisation(t *testing.T) {
	request := New("12345", "http://www.example.com", "3000")
	req, err := request.FullRequest("GET", "http://localhost:3000/api/apis", nil)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(requestHandler)
	handler.ServeHTTP(recorder, req)
	if recorder.Header().Get("Authorization") != request.Authorisation {
		t.Errorf(
			"Handler returned unexpected header.\nGot:\n\t%v\nExpected:\n\t%v",
			recorder.Header().Get("Authorization"),
			request.Authorisation,
		)
	}
}

func TestFullRequestContentType(t *testing.T) {
	request := New("12345", "http://www.example.com", "3000")
	req, err := request.FullRequest("GET", "http://localhost:3000/api/apis", nil)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(requestHandler)
	handler.ServeHTTP(recorder, req)
	expectedContentType := "application/json"
	if recorder.Header().Get("Content-Type") != expectedContentType {
		t.Errorf(
			"Handler returned unexpected header.\nGot:\n\t%v\nExpected:\n\t%v",
			recorder.Header().Get("Content-Type"),
			expectedContentType,
		)
	}
}

func TestFullRequestCorrectUrl(t *testing.T) {
	request := New("12345", "http://www.example.com", "3000")
	res, err := request.FullRequest("GET", "/api/apis", nil)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	expectedURL := fmt.Sprintf("%s:%s/api/apis", request.Domain, request.Port)
	if expectedURL != res.URL.String() {
		t.Errorf(
			"Handler returned unexpected URL.\nGot:\n\t%v\nExpected:\n\t%v",
			expectedURL,
			res.URL,
		)
	}
}
