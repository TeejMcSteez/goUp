package utils

import (
	"fmt"
)

// Takes in service data and checks for any bad responses
// Upon bad responses will fire trigger events if any and return bad responses or nil
func Check(data []ServiceData) []ServiceData {

	fmt.Println("Checking service data for errors")

	badRes := false
	var ret []ServiceData
	for _, el := range data {
		if el.ServiceHTTPResponse != "200" {
			badRes = true
			ret = append(ret, el)
		}
	}
	if badRes {
		Fire(ret)
	}
	return ret
}

