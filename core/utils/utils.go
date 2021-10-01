package utils

import "encoding/json"

// ReMarshal is a helper function to map 1 interface to another
func ReMarshal(in, out interface{}) error {
	j, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(j, out)
}
