package utils

import (
	"reflect"
	"testing"
)

func TestNormalizeJson(t *testing.T) {
	testcases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "basic reorder & whitespace",
			in:   "{\n  \"b\": 2,  \n \t\"a\" : 1\n}",
			want: "{\"a\":1,\"b\":2}",
		},
		{
			name: "empty input",
			in:   "",
			want: "",
		},
		{
			name: "invalid json returns error string (non-empty)",
			in:   "{ invalid }",
			want: NormalizeJson("{ invalid }"), // deterministic comparison not needed, only non-empty
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeJson(tc.in)
			if tc.name == "invalid json returns error string (non-empty)" {
				if got == "" {
					t.Fatalf("expected non-empty error string for invalid json")
				}
				return
			}
			if got != tc.want {
				t.Fatalf("NormalizeJson() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMergeObject(t *testing.T) {
	testcases := []struct {
		name string
		old  interface{}
		newV interface{}
		want interface{}
	}{
		{
			name: "maps merged, new overrides and adds",
			old:  map[string]interface{}{"a": 1, "b": 2, "c": map[string]interface{}{"d": 4}},
			newV: map[string]interface{}{"b": 3, "c": map[string]interface{}{"e": 5}, "f": 6},
			want: map[string]interface{}{"a": 1, "b": 3, "c": map[string]interface{}{"d": 4, "e": 5}, "f": 6},
		},
		{
			name: "arrays merged by position",
			old:  []interface{}{1, map[string]interface{}{"x": 1}},
			newV: []interface{}{2, map[string]interface{}{"x": 9}},
			want: []interface{}{2, map[string]interface{}{"x": 9}},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeObject(tc.old, tc.newV)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("MergeObject() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestUpdateObject(t *testing.T) {
	testcases := []struct {
		name string
		old  interface{}
		newV interface{}
		opt  UpdateJsonOption
		want interface{}
	}{
		{
			name: "ignore missing keeps old keys",
			old:  map[string]interface{}{"a": 1, "b": 2},
			newV: map[string]interface{}{"a": 10},
			opt:  UpdateJsonOption{IgnoreMissingProperty: true},
			want: map[string]interface{}{"a": 10, "b": 2},
		},
		{
			name: "no ignore missing drops old keys",
			old:  map[string]interface{}{"a": 1, "b": 2},
			newV: map[string]interface{}{"a": 10},
			opt:  UpdateJsonOption{IgnoreMissingProperty: false},
			want: map[string]interface{}{"a": 10},
		},
		{
			name: "string casing ignored",
			old:  "Hello",
			newV: "hello",
			opt:  UpdateJsonOption{IgnoreCasing: true},
			want: "Hello",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := UpdateObject(tc.old, tc.newV, tc.opt)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("UpdateObject() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestDiffObject(t *testing.T) {
	testcases := []struct {
		name string
		old  interface{}
		newV interface{}
		opt  UpdateJsonOption
		want interface{}
	}{
		{
			name: "nested map changed + new key",
			old:  map[string]interface{}{"a": 1, "b": 2, "c": map[string]interface{}{"d": 4}},
			newV: map[string]interface{}{"a": 1, "b": 3, "c": map[string]interface{}{"d": 4, "e": 5}, "f": 6},
			opt:  UpdateJsonOption{},
			want: map[string]interface{}{"b": 3, "c": map[string]interface{}{"e": 5}, "f": 6},
		},
		{
			name: "array changed -> full array returned",
			old:  []interface{}{1, 2, 3},
			newV: []interface{}{1, 2, 3, 4},
			opt:  UpdateJsonOption{},
			want: []interface{}{1, 2, 3, 4},
		},
		{
			name: "no change -> nil",
			old:  map[string]interface{}{"a": 1},
			newV: map[string]interface{}{"a": 1},
			opt:  UpdateJsonOption{},
			want: nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := DiffObject(tc.old, tc.newV, tc.opt)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("DiffObject() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestIsEmptyObject(t *testing.T) {
	testcases := []struct {
		name string
		v    interface{}
		want bool
	}{
		{name: "nil", v: nil, want: true},
		{name: "empty map", v: map[string]interface{}{}, want: true},
		{name: "empty slice", v: []interface{}{}, want: true},
		{name: "scalar", v: "x", want: false},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsEmptyObject(tc.v)
			if got != tc.want {
				t.Fatalf("IsEmptyObject() = %v, want %v (input=%#v)", got, tc.want, tc.v)
			}
		})
	}
}
