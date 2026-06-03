package oas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reqproof:req REQ-API-013
func TestHasTykExtensions(t *testing.T) {
	tests := []struct {
		name     string
		oasDoc   map[string]interface{}
		expected bool
	}{
		{
			name: "has extensions",
			oasDoc: map[string]interface{}{
				"openapi": "3.0.0",
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"name": "Test API"},
				},
			},
			expected: true,
		},
		{
			name: "no extensions",
			oasDoc: map[string]interface{}{
				"openapi": "3.0.0",
				"info":    map[string]interface{}{"title": "Test API"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasTykExtensions(tt.oasDoc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// reqproof:req REQ-API-013
func TestExtractAPIIDFromTykExtensions(t *testing.T) {
	tests := []struct {
		name     string
		oasDoc   map[string]interface{}
		expectedID string
		expectedFound bool
	}{
		{
			name: "has ID",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"id": "api-123"},
				},
			},
			expectedID: "api-123",
			expectedFound: true,
		},
		{
			name: "no extensions",
			oasDoc: map[string]interface{}{
				"info": map[string]interface{}{"title": "Test API"},
			},
			expectedID: "",
			expectedFound: false,
		},
		{
			name: "has extensions but no ID",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"name": "Test API"},
				},
			},
			expectedID: "",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, found := ExtractAPIIDFromTykExtensions(tt.oasDoc)
			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}

// reqproof:req REQ-API-011
func TestAddTykExtensions(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]interface{}
		expectError bool
		checkResult func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "plain OAS document",
			input: map[string]interface{}{
				"openapi": "3.0.0",
				"info": map[string]interface{}{
					"title":   "Swagger Petstore",
					"version": "1.0.0",
				},
				"servers": []interface{}{
					map[string]interface{}{"url": "http://petstore.swagger.io/v2"},
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, result map[string]interface{}) {
				// Should have extensions now
				assert.True(t, HasTykExtensions(result))
				
				// Check generated extensions
				tykExt, ok := result["x-tyk-api-gateway"].(map[string]interface{})
				require.True(t, ok)
				
				info, ok := tykExt["info"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Swagger Petstore", info["name"])
				
				upstream, ok := tykExt["upstream"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "http://petstore.swagger.io/v2", upstream["url"])
				
				server, ok := tykExt["server"].(map[string]interface{})
				require.True(t, ok)
				listenPath, ok := server["listenPath"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "/swagger-petstore/", listenPath["value"])
				assert.Equal(t, true, listenPath["strip"])
			},
		},
		{
			name: "already has extensions",
			input: map[string]interface{}{
				"openapi": "3.0.0",
				"info": map[string]interface{}{
					"title": "Test API",
				},
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"name": "Existing"},
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, result map[string]interface{}) {
				// Should keep existing extensions unchanged
				tykExt, ok := result["x-tyk-api-gateway"].(map[string]interface{})
				require.True(t, ok)
				
				info, ok := tykExt["info"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Existing", info["name"])
			},
		},
		{
			name: "missing info section",
			input: map[string]interface{}{
				"openapi": "3.0.0",
			},
			expectError: true,
		},
		{
			name: "missing title",
			input: map[string]interface{}{
				"openapi": "3.0.0",
				"info":    map[string]interface{}{"version": "1.0.0"},
			},
			expectError: true,
		},
		{
			name: "missing servers",
			input: map[string]interface{}{
				"openapi": "3.0.0",
				"info": map[string]interface{}{
					"title": "Test API",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AddTykExtensions(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, result)
			
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// reqproof:req REQ-API-012
func TestGenerateListenPath(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Swagger Petstore", "/swagger-petstore/"},
		{"My API", "/my-api/"},
		{"User Management API v2", "/user-management-api-v2/"},
		{"123 Test API", "/api-123-test-api/"},
		{"", "/api/"},
		{"Special!@#$%Characters", "/special-characters/"},
		{"API", "/api/"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := GenerateListenPath(tt.title)
			assert.Equal(t, tt.expected, result)
		})
	}
}// ===========================================================================
// MC/DC coverage of extractUpstreamURL — 3 decision points:
//   1. servers, ok := oasDoc["servers"].([]interface{}); !ok || len(servers)==0
//   2. firstServer, ok := servers[0].(map[string]interface{}); !ok
//   3. url, ok := firstServer["url"].(string); !ok
// ===========================================================================

// reqproof:req REQ-API-013
func TestExtractUpstreamURL_MCDC(t *testing.T) {
	tests := []struct {
		name    string
		oasDoc  map[string]interface{}
		want    string
		comment string
	}{
		{
			name:    "no servers key",
			oasDoc:  map[string]interface{}{"openapi": "3.0.0"},
			want:    "",
			comment: "branch 1: servers cast fails (!ok)",
		},
		{
			name:    "servers wrong type",
			oasDoc:  map[string]interface{}{"servers": "not-an-array"},
			want:    "",
			comment: "branch 1: servers exists but is not []interface{}",
		},
		{
			name:    "servers empty",
			oasDoc:  map[string]interface{}{"servers": []interface{}{}},
			want:    "",
			comment: "branch 1: len(servers)==0",
		},
		{
			name: "first server wrong type",
			oasDoc: map[string]interface{}{
				"servers": []interface{}{"not-a-map"},
			},
			want:    "",
			comment: "branch 2: firstServer cast fails",
		},
		{
			name: "url key missing",
			oasDoc: map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"description": "no url here"},
				},
			},
			want:    "",
			comment: "branch 3: url field absent",
		},
		{
			name: "url wrong type",
			oasDoc: map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"url": 12345},
				},
			},
			want:    "",
			comment: "branch 3: url present but not string",
		},
		{
			name: "happy path",
			oasDoc: map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"url": "https://upstream.example.com"},
				},
			},
			want:    "https://upstream.example.com",
			comment: "all three branches take the happy path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUpstreamURL(tt.oasDoc)
			assert.Equal(t, tt.want, got, tt.comment)
		})
	}
}

// ===========================================================================
// MC/DC coverage of ExtractAPIIDFromTykExtensions — 4 decision points:
//   1. HasTykExtensions(oasDoc) returns false
//   2. tykExt cast fails
//   3. info cast fails
//   4. id cast fails OR id == ""
// ===========================================================================

// reqproof:req REQ-API-013
func TestExtractAPIIDFromTykExtensions_MCDC(t *testing.T) {
	tests := []struct {
		name    string
		oasDoc  map[string]interface{}
		wantID  string
		wantOK  bool
		comment string
	}{
		{
			name:    "no x-tyk-api-gateway at all",
			oasDoc:  map[string]interface{}{"openapi": "3.0.0"},
			wantID:  "",
			wantOK:  false,
			comment: "branch 1: HasTykExtensions returns false",
		},
		{
			name: "tyk extension wrong type",
			oasDoc: map[string]interface{}{
				"openapi":           "3.0.0",
				"x-tyk-api-gateway": "not-a-map",
			},
			wantID:  "",
			wantOK:  false,
			comment: "branch 2: tykExt cast fails",
		},
		{
			name: "info missing",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{"upstream": "x"},
			},
			wantID:  "",
			wantOK:  false,
			comment: "branch 3: info absent",
		},
		{
			name: "info wrong type",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": "should-be-map",
				},
			},
			wantID:  "",
			wantOK:  false,
			comment: "branch 3: info present but wrong type",
		},
		{
			name: "id missing",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"name": "no id here"},
				},
			},
			wantID:  "",
			wantOK:  false,
			comment: "branch 4: id key absent",
		},
		{
			name: "id wrong type",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"id": 12345},
				},
			},
			wantID:  "",
			wantOK:  false,
			comment: "branch 4: id present but not string",
		},
		{
			name: "id empty string",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"id": ""},
				},
			},
			wantID:  "",
			wantOK:  false,
			comment: "branch 4: id is string but empty",
		},
		{
			name: "happy path",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"id": "abc123"},
				},
			},
			wantID:  "abc123",
			wantOK:  true,
			comment: "all branches take the happy path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := ExtractAPIIDFromTykExtensions(tt.oasDoc)
			assert.Equal(t, tt.wantID, gotID, tt.comment)
			assert.Equal(t, tt.wantOK, gotOK, tt.comment)
		})
	}
}

// ===========================================================================
// MC/DC coverage of GenerateListenPath edge branches:
//   - empty string title (fallback to "api")
//   - title sanitises to empty (all non-alphanumeric)
//   - title with leading digit ("3D Printer" -> "api-3d-printer")
// ===========================================================================

// reqproof:req REQ-API-012
func TestGenerateListenPath_MCDC(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"empty string", "", "/api/"},
		{"only special chars", "!@#$%^", "/api/"},
		{"leading digit", "3D Printer", "/api-3d-printer/"},
		{"leading number only", "42", "/api-42/"},
		{"normal", "User Service", "/user-service/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateListenPath(tt.title)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ===========================================================================
// MC/DC coverage of AddTykExtensions error branches:
//   - missing info
//   - missing title
//   - missing servers (upstream URL empty)
// ===========================================================================

// reqproof:req REQ-API-011
func TestAddTykExtensions_ErrorBranches_MCDC(t *testing.T) {
	tests := []struct {
		name    string
		oasDoc  map[string]interface{}
		wantErr string
	}{
		{
			name:    "missing info section",
			oasDoc:  map[string]interface{}{"openapi": "3.0.0"},
			wantErr: "missing info section",
		},
		{
			name: "missing title",
			oasDoc: map[string]interface{}{
				"openapi": "3.0.0",
				"info":    map[string]interface{}{"version": "1.0.0"},
			},
			wantErr: "missing info.title",
		},
		{
			name: "empty title string",
			oasDoc: map[string]interface{}{
				"openapi": "3.0.0",
				"info":    map[string]interface{}{"title": ""},
			},
			wantErr: "missing info.title",
		},
		{
			name: "missing servers",
			oasDoc: map[string]interface{}{
				"openapi": "3.0.0",
				"info":    map[string]interface{}{"title": "My API"},
			},
			wantErr: "no servers defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AddTykExtensions(tt.oasDoc)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

// reqproof:req REQ-API-014
func TestValidateOASStructure(t *testing.T) {
	tests := []struct {
		name    string
		doc     map[string]interface{}
		wantErr string
	}{
		{"nil doc", nil, "empty"},
		{"missing openapi", map[string]interface{}{"info": map[string]interface{}{"title": "x", "version": "1"}}, "missing required field 'openapi'"},
		{"openapi empty string", map[string]interface{}{"openapi": "", "info": map[string]interface{}{"title": "x", "version": "1"}}, "missing required field 'openapi'"},
		{"openapi 2.0", map[string]interface{}{"openapi": "2.0", "info": map[string]interface{}{"title": "x", "version": "1"}}, "only supports OpenAPI 3.x"},
		{"missing info", map[string]interface{}{"openapi": "3.0.0"}, "missing required 'info' section"},
		{"info missing title", map[string]interface{}{"openapi": "3.0.0", "info": map[string]interface{}{"version": "1"}}, "missing required field 'info.title'"},
		{"info empty title", map[string]interface{}{"openapi": "3.0.0", "info": map[string]interface{}{"title": "", "version": "1"}}, "missing required field 'info.title'"},
		{"info missing version", map[string]interface{}{"openapi": "3.0.0", "info": map[string]interface{}{"title": "x"}}, "missing required field 'info.version'"},
		{"info empty version", map[string]interface{}{"openapi": "3.0.0", "info": map[string]interface{}{"title": "x", "version": ""}}, "missing required field 'info.version'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOASStructure(tt.doc)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}

	t.Run("valid minimal 3.0", func(t *testing.T) {
		err := ValidateOASStructure(map[string]interface{}{
			"openapi": "3.0.0",
			"info":    map[string]interface{}{"title": "ok", "version": "1.0.0"},
		})
		assert.NoError(t, err)
	})

	t.Run("valid 3.1", func(t *testing.T) {
		err := ValidateOASStructure(map[string]interface{}{
			"openapi": "3.1.0",
			"info":    map[string]interface{}{"title": "ok", "version": "1.0.0"},
		})
		assert.NoError(t, err)
	})
}
