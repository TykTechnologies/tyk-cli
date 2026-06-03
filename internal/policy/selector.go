package policy

import (
	"fmt"
	"sort"
	"strings"
)

// ResolverAPI is the API information needed for selector resolution.
// It decouples the resolver from the wire type (types.OASAPI) and adds Tags support.
type ResolverAPI struct {
	ID         string
	Name       string
	ListenPath string
	Tags       []string
}

// ResolvedAccess represents a successfully resolved access entry.
type ResolvedAccess struct {
	APIID    string
	APIName  string
	Versions []string
}

// ResolveRequest describes a single access entry to resolve.
type ResolveRequest struct {
	SelectorType string   // "id", "name", "listenPath", "tags"
	Value        string   // for id, name, listenPath
	TagValues    []string // for tags selector
	Versions     []string
}

// FuzzySuggestion represents a fuzzy match suggestion.
type FuzzySuggestion struct {
	Name     string
	ID       string
	Distance int
}

// reqproof:req REQ-POL-011
func ResolveByName(name string, apis []ResolverAPI) (ResolverAPI, error) {
	var matches []ResolverAPI
	for _, api := range apis {
		if api.Name == name {
			matches = append(matches, api)
		}
	}

	if len(matches) == 0 {
		suggestions := FuzzySuggestions(name, apis, 3)
		msg := fmt.Sprintf("no API found for name %q", name)
		if len(suggestions) > 0 {
			parts := make([]string, len(suggestions))
			for i, s := range suggestions {
				parts[i] = fmt.Sprintf("%s (%s)", s.Name, s.ID)
			}
			msg += ". Did you mean: " + strings.Join(parts, ", ")
		}
		return ResolverAPI{}, fmt.Errorf("%s", msg)
	}

	if len(matches) > 1 {
		ids := make([]string, len(matches))
		for i, m := range matches {
			ids[i] = m.ID
		}
		return ResolverAPI{}, fmt.Errorf("ambiguous: name %q matches %d APIs: %s",
			name, len(matches), strings.Join(ids, ", "))
	}

	return matches[0], nil
}

// reqproof:req REQ-POL-011
func ResolveByListenPath(path string, apis []ResolverAPI) (ResolverAPI, error) {
	var matches []ResolverAPI
	for _, api := range apis {
		if api.ListenPath == path {
			matches = append(matches, api)
		}
	}

	if len(matches) == 0 {
		return ResolverAPI{}, fmt.Errorf("no API found for listenPath %q", path)
	}

	if len(matches) > 1 {
		ids := make([]string, len(matches))
		for i, m := range matches {
			ids[i] = m.ID
		}
		return ResolverAPI{}, fmt.Errorf("ambiguous: listenPath %q matches %d APIs: %s",
			path, len(matches), strings.Join(ids, ", "))
	}

	return matches[0], nil
}

// reqproof:req REQ-POL-011
func ResolveByID(id string, apis []ResolverAPI) (ResolverAPI, error) {
	for _, api := range apis {
		if api.ID == id {
			return api, nil
		}
	}
	return ResolverAPI{}, fmt.Errorf("no API found for id %q", id)
}

// reqproof:req REQ-POL-011
func ResolveByTags(tags []string, apis []ResolverAPI) ([]ResolverAPI, error) {
	var matches []ResolverAPI
	for _, api := range apis {
		if hasAllTags(api.Tags, tags) {
			matches = append(matches, api)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no APIs matched tags %v", tags)
	}

	return matches, nil
}

// reqproof:req REQ-POL-011
func hasAllTags(apiTags, requiredTags []string) bool {
	tagSet := make(map[string]struct{}, len(apiTags))
	for _, t := range apiTags {
		tagSet[t] = struct{}{}
	}
	for _, t := range requiredTags {
		if _, ok := tagSet[t]; !ok {
			return false
		}
	}
	return true
}

// reqproof:req REQ-POL-011
// reqproof:req SW-REQ-001
// reqproof:req SW-REQ-002
// reqproof:req SW-REQ-003
func FuzzySuggestions(query string, apis []ResolverAPI, n int) []FuzzySuggestion {
	type scored struct {
		api      ResolverAPI
		distance int
	}

	var candidates []scored
	for _, api := range apis {
		d := levenshtein(query, api.Name)
		candidates = append(candidates, scored{api: api, distance: d})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})

	n = min(n, len(candidates))

	result := make([]FuzzySuggestion, n)
	for i := 0; i < n; i++ {
		result[i] = FuzzySuggestion{
			Name:     candidates[i].api.Name,
			ID:       candidates[i].api.ID,
			Distance: candidates[i].distance,
		}
	}
	return result
}

// reqproof:req REQ-POL-011
func levenshtein(a, b string) int {
	lenA, lenB := len(a), len(b)
	if lenA == 0 {
		return lenB
	}
	if lenB == 0 {
		return lenA
	}

	// Single-row optimization: only keep the previous row in memory.
	prev := make([]int, lenB+1)
	for j := 0; j <= lenB; j++ {
		prev[j] = j
	}

	for i := 1; i <= lenA; i++ {
		curr := make([]int, lenB+1)
		curr[0] = i
		for j := 1; j <= lenB; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = curr
	}
	return prev[lenB]
}

// ResolveAccessEntries resolves a batch of access entry requests against an API list.
// It collects all errors before returning. Successful resolutions and errors are returned separately.
// reqproof:req REQ-POL-011
func ResolveAccessEntries(requests []ResolveRequest, apis []ResolverAPI) ([]ResolvedAccess, []error) {
	var resolved []ResolvedAccess
	var errs []error

	for _, req := range requests {
		switch req.SelectorType {
		case "id":
			api, err := ResolveByID(req.Value, apis)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			resolved = append(resolved, ResolvedAccess{
				APIID: api.ID, APIName: api.Name, Versions: req.Versions,
			})

		case "name":
			api, err := ResolveByName(req.Value, apis)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			resolved = append(resolved, ResolvedAccess{
				APIID: api.ID, APIName: api.Name, Versions: req.Versions,
			})

		case "listenPath":
			api, err := ResolveByListenPath(req.Value, apis)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			resolved = append(resolved, ResolvedAccess{
				APIID: api.ID, APIName: api.Name, Versions: req.Versions,
			})

		case "tags":
			matches, err := ResolveByTags(req.TagValues, apis)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			for _, api := range matches {
				resolved = append(resolved, ResolvedAccess{
					APIID: api.ID, APIName: api.Name, Versions: req.Versions,
				})
			}

		default:
			errs = append(errs, fmt.Errorf("unknown selector type %q", req.SelectorType))
		}
	}

	return resolved, errs
}
