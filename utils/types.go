package utils

// InterfaceToMap function converts interface types into a map with a single key
func InterfaceToMap(input interface{}) map[string]interface{} {
	return input.(map[string]interface{})
}
