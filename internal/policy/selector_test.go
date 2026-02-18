package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAPIList() []ResolverAPI {
	return []ResolverAPI{
		{ID: "a1b2c3d4e5f6", Name: "users-api", ListenPath: "/users/", Tags: []string{"public", "v1"}},
		{ID: "g7h8i9j0k1l2", Name: "orders-api", ListenPath: "/orders/", Tags: []string{"internal", "v1"}},
		{ID: "m3n4o5p6q7r8", Name: "payments-api", ListenPath: "/payments/", Tags: []string{"internal", "v2"}},
	}
}

func TestResolveByName(t *testing.T) {
	apis := testAPIList()

	t.Run("exact match", func(t *testing.T) {
		resolved, err := ResolveByName("users-api", apis)
		require.NoError(t, err)
		assert.Equal(t, "a1b2c3d4e5f6", resolved.ID)
		assert.Equal(t, "users-api", resolved.Name)
	})

	t.Run("not found with fuzzy suggestions", func(t *testing.T) {
		_, err := ResolveByName("inventori-api", apis)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no API found")
	})

	t.Run("ambiguous match", func(t *testing.T) {
		dupes := []ResolverAPI{
			{ID: "dup-1", Name: "api-service", ListenPath: "/svc-1/"},
			{ID: "dup-2", Name: "api-service", ListenPath: "/svc-2/"},
		}
		_, err := ResolveByName("api-service", dupes)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ambiguous")
	})
}

func TestResolveByListenPath(t *testing.T) {
	apis := testAPIList()

	t.Run("exact match", func(t *testing.T) {
		resolved, err := ResolveByListenPath("/orders/", apis)
		require.NoError(t, err)
		assert.Equal(t, "g7h8i9j0k1l2", resolved.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := ResolveByListenPath("/unknown/", apis)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no API found")
	})
}

func TestResolveByID(t *testing.T) {
	apis := testAPIList()

	t.Run("found", func(t *testing.T) {
		resolved, err := ResolveByID("a1b2c3d4e5f6", apis)
		require.NoError(t, err)
		assert.Equal(t, "users-api", resolved.Name)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := ResolveByID("nonexistent-id", apis)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no API found")
	})
}

func TestResolveByTags(t *testing.T) {
	apis := testAPIList()

	t.Run("single tag matches multiple", func(t *testing.T) {
		resolved, err := ResolveByTags([]string{"internal"}, apis)
		require.NoError(t, err)
		assert.Len(t, resolved, 2)
	})

	t.Run("multiple tags intersection", func(t *testing.T) {
		resolved, err := ResolveByTags([]string{"internal", "v1"}, apis)
		require.NoError(t, err)
		require.Len(t, resolved, 1)
		assert.Equal(t, "orders-api", resolved[0].Name)
	})

	t.Run("no match", func(t *testing.T) {
		_, err := ResolveByTags([]string{"legacy"}, apis)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no APIs matched tags")
	})
}

func TestFuzzySuggestions(t *testing.T) {
	apis := testAPIList()

	suggestions := FuzzySuggestions("inventori-api", apis, 3)
	require.NotEmpty(t, suggestions)
	// Closest should come first, and all APIs have reasonable edit distance
	assert.LessOrEqual(t, len(suggestions), 3)
	// The closest should be one of the real APIs
	assert.NotEmpty(t, suggestions[0].Name)
	assert.NotEmpty(t, suggestions[0].ID)
	assert.Greater(t, suggestions[0].Distance, 0)
}

func TestResolveAccessEntries(t *testing.T) {
	apis := testAPIList()
	entries := []ResolveRequest{
		{SelectorType: "name", Value: "users-api", Versions: []string{"v1"}},
		{SelectorType: "listenPath", Value: "/orders/", Versions: []string{"v1", "v2"}},
		{SelectorType: "id", Value: "m3n4o5p6q7r8", Versions: []string{"v1"}},
		{SelectorType: "tags", TagValues: []string{"public", "v1"}, Versions: []string{"v1"}},
	}

	resolved, errs := ResolveAccessEntries(entries, apis)
	require.Empty(t, errs, "expected no resolution errors, got: %v", errs)
	require.Len(t, resolved, 4)

	assert.Equal(t, "a1b2c3d4e5f6", resolved[0].APIID)
	assert.Equal(t, "g7h8i9j0k1l2", resolved[1].APIID)
	assert.Equal(t, "m3n4o5p6q7r8", resolved[2].APIID)
	// tags resolved to users-api (has both "public" and "v1")
	assert.Equal(t, "a1b2c3d4e5f6", resolved[3].APIID)
}

func TestResolveAccessEntries_CollectsErrors(t *testing.T) {
	apis := testAPIList()
	entries := []ResolveRequest{
		{SelectorType: "name", Value: "nonexistent-api"},
		{SelectorType: "name", Value: "users-api", Versions: []string{"v1"}}, // valid
		{SelectorType: "listenPath", Value: "/nowhere/"},
	}

	resolved, errs := ResolveAccessEntries(entries, apis)
	// The valid entry should still resolve
	assert.Len(t, resolved, 1)
	// Two errors should be collected
	assert.Len(t, errs, 2)
}
