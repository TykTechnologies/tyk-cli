package types

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// PolicyFile is the top-level structure users write in YAML policy files.
type PolicyFile struct {
	ID        string        `yaml:"id" json:"id"`
	Name      string        `yaml:"name" json:"name"`
	Tags      []string      `yaml:"tags,omitempty" json:"tags,omitempty"`
	RateLimit *RateLimit    `yaml:"rateLimit,omitempty" json:"rateLimit,omitempty"`
	Quota     *Quota        `yaml:"quota,omitempty" json:"quota,omitempty"`
	KeyTTL    Duration      `yaml:"keyTTL,omitempty" json:"keyTTL,omitempty"`
	Access    []AccessEntry `yaml:"access" json:"access"`
}

// RateLimit defines requests-per-window rate limiting.
type RateLimit struct {
	Requests int64    `yaml:"requests" json:"requests"`
	Per      Duration `yaml:"per" json:"per"`
}

// Quota defines quota limits over a time period.
type Quota struct {
	Limit  int64    `yaml:"limit" json:"limit"`
	Period Duration `yaml:"period" json:"period"`
}

// AccessEntry represents a single API access rule with exactly one selector.
// Exactly one of ID, Name, ListenPath, or Tags must be set.
type AccessEntry struct {
	ID         string   `yaml:"id,omitempty" json:"id,omitempty"`
	Name       string   `yaml:"name,omitempty" json:"name,omitempty"`
	ListenPath string   `yaml:"listenPath,omitempty" json:"listenPath,omitempty"`
	Tags       []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Versions   []string `yaml:"versions,omitempty" json:"versions,omitempty"`
}

// Duration is a string that accepts both string durations ("30d", "1m") and
// plain integers (interpreted as seconds) from YAML input. Internally stored
// as the string representation.
type Duration string

// reqproof:req REQ-POL-010
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	*d = Duration(value.Value)
	return nil
}

// DashboardPolicy represents the wire format returned by the Tyk Dashboard API.
type DashboardPolicy struct {
	MID              string                  `json:"_id,omitempty"`
	ID               string                  `json:"id"`
	Name             string                  `json:"name"`
	OrgID            string                  `json:"org_id,omitempty"`
	Rate             int64                   `json:"rate"`
	Per              int64                   `json:"per"`
	QuotaMax         int64                   `json:"quota_max"`
	QuotaRenewalRate int64                   `json:"quota_renewal_rate"`
	KeyExpiresIn     int64                   `json:"key_expires_in"`
	Tags             []string                `json:"tags,omitempty"`
	AccessRights     map[string]*AccessRight `json:"access_rights,omitempty"`
	Active           bool                    `json:"active"`
	IsInactive       bool                    `json:"is_inactive"`
}

// AccessRight represents per-API access configuration in the Dashboard wire format.
type AccessRight struct {
	APIID       string          `json:"api_id"`
	APIName     string          `json:"api_name"`
	Versions    []string        `json:"versions"`
	AllowedURLs []AllowedURL    `json:"allowed_urls"`
	Limit       *RateQuotaLimit `json:"limit"`
}

// AllowedURL represents a URL-level access restriction (Phase 1: unused).
type AllowedURL struct {
	URL     string   `json:"url"`
	Methods []string `json:"methods"`
}

// RateQuotaLimit represents per-API rate and quota limits (Phase 1: unused).
type RateQuotaLimit struct {
	Rate             int64 `json:"rate"`
	Per              int64 `json:"per"`
	QuotaMax         int64 `json:"quota_max"`
	QuotaRenewalRate int64 `json:"quota_renewal_rate"`
}

// DashboardPolicyListResponse represents the paginated list response from
// the Dashboard policy API.
type DashboardPolicyListResponse struct {
	Data       []DashboardPolicy `json:"Data"`
	Pages      int               `json:"Pages"`
	StatusCode int               `json:"StatusCode"`
}

// ValidationError describes a single validation failure on a policy file.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Kind    string `json:"kind"` // "schema", "duration", or "selector"
}

// reqproof:req REQ-POL-006
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Kind)
}

// ValidationErrors collects multiple validation failures.
type ValidationErrors []ValidationError

// reqproof:req REQ-POL-006
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}
	return fmt.Sprintf("%d validation errors: %s (and %d more)", len(ve), ve[0].Error(), len(ve)-1)
}

// reqproof:req REQ-POL-007
func (ar *AccessRight) MarshalJSON() ([]byte, error) {
	type Alias AccessRight
	a := &struct {
		*Alias
		AllowedURLs []AllowedURL    `json:"allowed_urls"`
		Limit       json.RawMessage `json:"limit"`
	}{
		Alias: (*Alias)(ar),
	}
	if ar.AllowedURLs == nil {
		a.AllowedURLs = []AllowedURL{}
	} else {
		a.AllowedURLs = ar.AllowedURLs
	}
	if ar.Limit == nil {
		a.Limit = json.RawMessage("null")
	} else {
		b, err := json.Marshal(ar.Limit)
		if err != nil { //mcdc:ignore json.Marshal cannot fail for *RateQuotaLimit (struct of int64 fields); defensive plumbing only
			return nil, err
		}
		a.Limit = b
	}
	return json.Marshal(a)
}
