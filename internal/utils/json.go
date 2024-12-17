package utils

import (
	"encoding/json"
	"fmt"
)

func NormalizeJson(input interface{}) string {
	if input == nil || input == "" {
		return ""
	}

	jsonString, ok := input.(string)
	if !ok {
		return ""
	}

	var j interface{}

	if err := json.Unmarshal([]byte(jsonString), &j); err != nil {
		return fmt.Sprintf("Error parsing JSON: %+v", err)
	}
	b, _ := json.Marshal(j)
	return string(b)
}

// MergeObject is used to merge object old and new, if overlaps, use new value
func MergeObject(old interface{}, new interface{}) interface{} {
	if new == nil {
		return new
	}
	switch oldValue := old.(type) {
	case map[string]interface{}:
		if newMap, ok := new.(map[string]interface{}); ok {
			res := make(map[string]interface{})
			for key, value := range oldValue {
				if _, ok := newMap[key]; ok {
					res[key] = MergeObject(value, newMap[key])
				} else {
					res[key] = value
				}
			}
			for key, newValue := range newMap {
				if res[key] == nil {
					res[key] = newValue
				}
			}
			return res
		}
	case []interface{}:
		if newArr, ok := new.([]interface{}); ok {
			if len(oldValue) != len(newArr) {
				return newArr
			}
			res := make([]interface{}, 0)
			for index := range oldValue {
				res = append(res, MergeObject(oldValue[index], newArr[index]))
			}
			return res
		}
	}
	return new
}
