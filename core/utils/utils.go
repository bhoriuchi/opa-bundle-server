package utils

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"
)

// ReMarshal is a helper function to map 1 interface to another
func ReMarshal(in, out interface{}) error {
	j, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(j, out)
}

// Unmarshal unmarshals json or yaml content to an interface
func Unmarshal(content []byte, out interface{}) error {
	if bytes.HasPrefix(content, []byte("{")) && bytes.HasSuffix(content, []byte("}")) {
		if err := json.Unmarshal(content, out); err != nil {
			return err
		}
		return nil
	}

	if err := yaml.Unmarshal(content, out); err != nil {
		return err
	}

	return nil
}

// StringSliceContains returns true of the slice contains the string
func StringSliceContains(s []string, val string) bool {
	if s != nil && val != "" {
		for _, v := range s {
			if val == v {
				return true
			}
		}
	}

	return false
}
