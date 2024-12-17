package utils

import (
	jmes "github.com/jmespath/go-jmespath"
)

// ExtractObjectJMES is used to extract object from old using JMES path
func ExtractObjectJMES(old interface{}, pathKey, path string) interface{} {
	result := make(map[string]interface{}, 1)
	value, err := jmes.Search(path, old)
	if err != nil {
		return nil
	}
	result[pathKey] = value
	return result
}
