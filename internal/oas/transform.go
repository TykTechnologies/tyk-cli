package oas

import (
	"fmt"
	"regexp"
	"strings"
)

// TykExtensionKey is the key for Tyk-specific extensions in OAS documents
const TykExtensionKey = "x-tyk-api-gateway"

// HasTykExtensions checks if an OAS document contains x-tyk-api-gateway extensions
func HasTykExtensions(oasDoc map[string]interface{}) bool {
	_, exists := oasDoc[TykExtensionKey]
	return exists
}

// ExtractAPIIDFromTykExtensions extracts the API ID from x-tyk-api-gateway.info.id
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

// VersioningMetadata holds versioning info from x-tyk-api-gateway.info.versioning
type VersioningMetadata struct {
	BaseAPIID     string
	VersionName   string
	SetDefault    bool
	IsVersionFile bool
}

// ExtractVersioningMetadata extracts versioning metadata from an OAS document.
func ExtractVersioningMetadata(oasDoc map[string]interface{}) VersioningMetadata {
	tykExt, ok := oasDoc[TykExtensionKey].(map[string]interface{})
	if !ok {
		return VersioningMetadata{}
	}

	info, ok := tykExt["info"].(map[string]interface{})
	if !ok {
		return VersioningMetadata{}
	}

	versioning, ok := info["versioning"].(map[string]interface{})
	if !ok {
		return VersioningMetadata{}
	}

	meta := VersioningMetadata{IsVersionFile: true}

	if v, ok := versioning["base_api_id"].(string); ok {
		meta.BaseAPIID = v
	}
	if v, ok := versioning["version_name"].(string); ok {
		meta.VersionName = v
	}
	if v, ok := versioning["set_default"].(bool); ok {
		meta.SetDefault = v
	}

	return meta
}

// AddTykExtensions adds minimal x-tyk-api-gateway extensions to a plain OAS document
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

// extractUpstreamURL extracts the upstream URL from OAS servers section
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