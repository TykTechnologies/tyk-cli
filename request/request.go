package request

import (
	"bytes"
	utils "github.com/TykTechnologies/tyk-cli/utils"
	"net/http"
	"time"
)

// Request struct used to set parameters for a HTTP request object
type Request struct {
	Authorisation string
	Domain        string
	Port          string
	Client        *http.Client
}

// New function used initialise HTTP request objects
func New(auth, dom, prt string) *Request {
	return &Request{auth, utils.CheckDomain(dom), prt,
		&http.Client{Timeout: 10 * time.Second}}
}

// FullRequest function is used to generate a HTTP request with headers
func (request *Request) FullRequest(requestType string, url string, payload []byte) (*http.Request, error) {
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(payload))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", request.Authorisation)

	return req, err
}
