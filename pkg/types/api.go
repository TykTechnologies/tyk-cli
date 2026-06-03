package types

import "encoding/json"

// APIResponse represents the response structure from Tyk Dashboard API
type APIResponse struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
	Meta    string `json:"Meta"`
	ID      string `json:"ID,omitempty"`
}

// OASAPIResponse represents an OAS API response from Tyk Dashboard
type OASAPIResponse struct {
	APIResponse
	API *OASAPI `json:"api"`
}

// OASAPIListResponse represents a list of OAS APIs
type OASAPIListResponse struct {
	APIResponse
	APIs []*OASAPI `json:"apis"`
}

// OASAPI represents an OAS API in Tyk Dashboard
type OASAPI struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	ListenPath       string                 `json:"listen_path"`
	DefaultVersion   string                 `json:"default_version"`
	VersionData      map[string]*APIVersion `json:"version_data"`
	OAS              map[string]interface{} `json:"oas"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	CustomDomain     string                 `json:"custom_domain,omitempty"`
	UpstreamURL      string                 `json:"upstream_url,omitempty"`
}

// APIVersion represents version data for an API
type APIVersion struct {
	Name         string                 `json:"name"`
	SetDefault   bool                   `json:"set_default"`
	OAS          map[string]interface{} `json:"oas"`
	UpstreamURL  string                 `json:"upstream_url,omitempty"`
	ListenPath   string                 `json:"listen_path,omitempty"`
	CustomDomain string                 `json:"custom_domain,omitempty"`
}

// CreateOASAPIRequest represents the request to create an OAS API
type CreateOASAPIRequest struct {
	OAS         json.RawMessage `json:"oas"`
	BaseAPIID   string          `json:"base_api_id,omitempty"`
	NewVersionName string       `json:"new_version_name,omitempty"`
	SetDefault  bool            `json:"set_default,omitempty"`
	UpstreamURL string          `json:"upstream_url,omitempty"`
	ListenPath  string          `json:"listen_path,omitempty"`
	CustomDomain string         `json:"custom_domain,omitempty"`
}

// UpdateOASAPIRequest represents the request to update an OAS API
type UpdateOASAPIRequest struct {
	OAS         json.RawMessage `json:"oas"`
	VersionName string          `json:"version_name,omitempty"`
	SetDefault  bool            `json:"set_default,omitempty"`
}

// VersionListResponse represents the response for listing API versions
type VersionListResponse struct {
	APIResponse
	Versions []string `json:"versions"`
	Default  string   `json:"default"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Status  int                    `json:"status"`
	Code    string                 `json:"code,omitempty"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// reqproof:req REQ-API-022
func (e *ErrorResponse) Error() string {
	return e.Message
}