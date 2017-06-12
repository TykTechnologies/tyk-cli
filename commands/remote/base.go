package remote

import (
	"fmt"
	"io"
)

func List(w io.Writer, conf []interface{}, verbose bool) {
	for _, remote := range conf {
		remote := remote.(map[string]interface{})
		if verbose {
			fmt.Fprintf(w, "%v - %v\n", remote["alias"], remote["url"])
		} else {
			fmt.Fprintf(w, "%v\n", remote["alias"])
		}
	}
}
