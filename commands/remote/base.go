package remote

import (
	"fmt"
)

func List(conf []interface{}, verbose bool) string {
	var output string
	for i := range conf {
		remote := conf[i].(map[string]interface{})
		if verbose {
			output += fmt.Sprintf("%v - %v\n", remote["alias"], remote["url"])
		} else {
			output += fmt.Sprintf("%v\n", remote["alias"])
		}
	}
	return output
}
