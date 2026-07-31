package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
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

type UpdateJsonOption struct {
	IgnoreCasing          bool
	IgnoreMissingProperty bool
	IgnoreNullProperty    bool
}

// isRedactedString reports whether s is a placeholder that Graph returned instead of the
// real value, and therefore must never be written back. An empty string is included
// because Graph omits or blanks out several write-only properties on read.
func isRedactedString(s string, option UpdateJsonOption) bool {
	if !option.IgnoreMissingProperty {
		return false
	}
	return regexp.MustCompile(`^\*+$`).MatchString(s) || s == "<redacted>" || s == ""
}

// UpdateObject is used to get an updated object which has same schema as old, but with new value
func UpdateObject(old interface{}, new interface{}, option UpdateJsonOption) interface{} {
	if reflect.DeepEqual(old, new) {
		return old
	}
	switch oldValue := old.(type) {
	case map[string]interface{}:
		if newMap, ok := new.(map[string]interface{}); ok {
			res := make(map[string]interface{})
			for key, value := range oldValue {
				switch {
				case value == nil && option.IgnoreNullProperty:
					res[key] = nil
				case newMap[key] != nil:
					res[key] = UpdateObject(value, newMap[key], option)
				case option.IgnoreMissingProperty || isZeroValue(value):
					res[key] = value
				}
			}
			return res
		}
	case []interface{}:
		if newArr, ok := new.([]interface{}); ok {
			if len(oldValue) == 0 {
				return new
			}

			hasIdentifier := identifierOfArrayItem(oldValue[0]) != ""
			if !hasIdentifier {
				if len(oldValue) != len(newArr) {
					return newArr
				}
				res := make([]interface{}, 0)
				for index := range oldValue {
					res = append(res, UpdateObject(oldValue[index], newArr[index], option))
				}
				return res
			}

			res := make([]interface{}, 0)
			used := make([]bool, len(newArr))

			for _, oldItem := range oldValue {
				found := false
				for index, newItem := range newArr {
					if reflect.DeepEqual(oldItem, newItem) && !used[index] {
						res = append(res, UpdateObject(oldItem, newItem, option))
						used[index] = true
						found = true
						break
					}
				}
				if found {
					continue
				}
				for index, newItem := range newArr {
					if areSameArrayItems(oldItem, newItem) && !used[index] {
						res = append(res, UpdateObject(oldItem, newItem, option))
						used[index] = true
						break
					}
				}
			}

			for index, newItem := range newArr {
				if !used[index] {
					res = append(res, newItem)
				}
			}
			return res
		}
	case string:
		if newStr, ok := new.(string); ok {
			if option.IgnoreCasing && strings.EqualFold(oldValue, newStr) {
				return oldValue
			}
			if isRedactedString(newStr, option) {
				return oldValue
			}
		}
	}
	return new
}

func areSameArrayItems(a, b interface{}) bool {
	aId := identifierOfArrayItem(a)
	bId := identifierOfArrayItem(b)
	if aId == "" || bId == "" {
		return false
	}
	return aId == bId
}

func identifierOfArrayItem(input interface{}) string {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return ""
	}
	name := inputMap["name"]
	if name == nil {
		return ""
	}
	nameValue, ok := name.(string)
	if !ok {
		return ""
	}
	return nameValue
}

func isZeroValue(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case map[string]interface{}:
		return len(v) == 0
	case []interface{}:
		return len(v) == 0
	case string:
		return len(v) == 0
	case int, int32, int64, float32, float64:
		return v == 0
	case bool:
		return !v
	}
	return false
}

// DiffObject computes a minimal patch that transforms old -> new.
// It returns:
// - nil if there are no changes
// - a map[string]interface{} listing every changed top-level property in full
// - the new value itself for arrays and scalars when they differ
//
// Values are never partially diffed. Microsoft Graph applies JSON Merge PATCH (RFC 7396)
// at the top level only: a nested object (a "complex type", e.g.
// federatedIdentityCredential.claimsMatchingExpression) is replaced wholesale, so any
// sub-field omitted from the request reverts to its default. Arrays and scalars are
// replaced the same way. Every changed property is therefore emitted in full from the
// new/config value - sending a partial complex value would drop unchanged siblings such
// as languageVersion (see issue #137).
//
// Special handling: OData metadata fields (keys starting with "@odata.") are always
// included in the result when they exist in the new object and there are other changes.
// This is required for Microsoft Graph API endpoints that use polymorphic types and
// need the discriminator field (@odata.type) to be present in PATCH requests.
func DiffObject(old interface{}, new interface{}, option UpdateJsonOption) interface{} {
	if reflect.DeepEqual(old, new) {
		return nil
	}
	switch oldValue := old.(type) {
	case map[string]interface{}:
		if newMap, ok := new.(map[string]interface{}); ok {
			res := make(map[string]interface{})
			// include keys present in new
			for key, newVal := range newMap {
				if oldVal, ok := oldValue[key]; ok {
					if d := DiffObject(oldVal, newVal, option); d != nil {
						res[key] = newVal
					}
				} else {
					// key doesn't exist in old -> create
					res[key] = newVal
				}
			}

			// If we have changes, also include any @odata.* fields from newMap
			// even if they haven't changed. These are OData metadata fields that
			// some Microsoft Graph endpoints require in PATCH requests.
			if len(res) > 0 {
				for key, newVal := range newMap {
					// Only add @odata.* fields that aren't already in res
					if strings.HasPrefix(key, "@odata.") && res[key] == nil {
						// Field exists and unchanged, but include it anyway for @odata fields
						res[key] = newVal
					}
				}
			}

			if len(res) == 0 {
				return nil
			}
			return res
		}
	case []interface{}:
		if newArr, ok := new.([]interface{}); ok {
			if reflect.DeepEqual(oldValue, newArr) {
				return nil
			}
			// For arrays, send the full new array when changed
			return newArr
		}
	case string:
		if newStr, ok := new.(string); ok {
			if option.IgnoreCasing && strings.EqualFold(oldValue, newStr) {
				return nil
			}
			if isRedactedString(newStr, option) {
				return nil
			}
		}
	}
	// primitives or differing types -> return new
	return new
}
