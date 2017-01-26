package utils

func InterfaceToMap(input interface{}) map[string]interface{} {
	return input.(map[string]interface{})
}
