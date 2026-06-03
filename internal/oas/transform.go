package oas

import (
	"fmt"
	"regexp"
	"strings"
)

// TykExtensionKey is the key for Tyk-specific extensions in OAS documents
const TykExtensionKey = "x-tyk-api-gateway"

// reqproof:req REQ-API-013
func HasTykExtensions(oasDoc map[string]interface{}) bool {
	_, exists := oasDoc[TykExtensionKey]
	return exists
}

// reqproof:req REQ-API-013
func ExtractAPIIDFromTykExtensions(oasDoc map[string]interface{}) (string, bool) {
	if !HasTykExtensions(oasDoc) {
		return "", false
	}
	
	tykExt, ok := oasDoc[TykExtensionKey].(map[string]interface{})
	if !ok {
		return "", false
	}
	
	info, ok := tykExt["info"].(map[string]interface{})
	if !ok {
		return "", false
	}
	
	id, ok := info["id"].(string)
	if !ok || id == "" {
		return "", false
	}
	
	return id, true
}

// AddTykExtensions adds minimal x-tyk-api-gateway extensions to a plain OAS document
// reqproof:req REQ-API-011
func AddTykExtensions(oasDoc map[string]interface{}) (map[string]interface{}, error) {
	if HasTykExtensions(oasDoc) {
		return oasDoc, nil // Already has extensions
	}
	
	// Extract info from OAS
	info, ok := oasDoc["info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid OAS document: missing info section")
	}
	
	title, ok := info["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("invalid OAS document: missing info.title")
	}
	
	// Extract upstream URL from servers
	upstreamURL := extractUpstreamURL(oasDoc)
	if upstreamURL == "" {
		return nil, fmt.Errorf("invalid OAS document: no servers defined or server URL missing")
	}
	
	// Generate listen path from title
	listenPath := GenerateListenPath(title)
	
	// Create a copy of the document
	result := make(map[string]interface{})
	for k, v := range oasDoc {
		result[k] = v
	}
	
	// Add x-tyk-api-gateway extensions
	result[TykExtensionKey] = map[string]interface{}{
		"info": map[string]interface{}{
			"name": title,
			"state": map[string]interface{}{
				"active": true,
			},
		},
		"upstream": map[string]interface{}{
			"url": upstreamURL,
		},
		"server": map[string]interface{}{
			"listenPath": map[string]interface{}{
				"value": listenPath,
				"strip": true,
			},
		},
	}
	
	return result, nil
}

// reqproof:req REQ-API-011
func extractUpstreamURL(oasDoc map[string]interface{}) string {
	servers, ok := oasDoc["servers"].([]interface{})
	if !ok || len(servers) == 0 {
		return ""
	}
	
	firstServer, ok := servers[0].(map[string]interface{})
	if !ok {
		return ""
	}
	
	url, ok := firstServer["url"].(string)
	if !ok {
		return ""
	}
	
	return url
}

// GenerateListenPath creates a listen path from API title
// Examples: "My API" -> "/my-api/", "Swagger Petstore" -> "/swagger-petstore/"
// reqproof:req REQ-API-012
// reqproof:req SW-REQ-007
// reqproof:req SW-REQ-008
// reqproof:req SW-REQ-009
func GenerateListenPath(title string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	slug := strings.ToLower(title)
	
	// Replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	
	// Ensure it doesn't start with a number (invalid path)
	if len(slug) > 0 && slug[0] >= '0' && slug[0] <= '9' {
		slug = "api-" + slug
	}
	
	// Fallback if empty
	if slug == "" {
		slug = "api"
	}
	
	return "/" + slug + "/"
}
// ValidateOASStructure performs a lightweight structural check on an OAS
// document before it is submitted to the Dashboard. It enforces the minimum
// schema invariants that the Dashboard rejects with opaque errors: openapi
// field present and non-empty, info section with non-empty title and version.
// It does NOT perform full OpenAPI JSON Schema validation — the goal is to
// catch obvious authoring errors locally and surface a clear message before
// any network round-trip.
//
// reqproof:req REQ-API-014
func ValidateOASStructure(oasDoc map[string]interface{}) error {
	if oasDoc == nil {
		return fmt.Errorf("OAS document is empty")
	}

	openapi, ok := oasDoc["openapi"].(string)
	if !ok || openapi == "" {
		return fmt.Errorf("OAS document is missing required field 'openapi'")
	}
	if !strings.HasPrefix(openapi, "3.") {
		return fmt.Errorf("OAS document declares openapi=%q; tyk-cli only supports OpenAPI 3.x", openapi)
	}

	info, ok := oasDoc["info"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("OAS document is missing required 'info' section")
	}
	title, ok := info["title"].(string)
	if !ok || title == "" {
		return fmt.Errorf("OAS document is missing required field 'info.title'")
	}
	version, ok := info["version"].(string)
	if !ok || version == "" {
		return fmt.Errorf("OAS document is missing required field 'info.version'")
	}

	return nil
}
