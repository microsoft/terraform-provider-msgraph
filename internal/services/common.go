package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoft/terraform-provider-msgraph/internal/dynamic"
)

func AsMapOfString(input types.Map) map[string]string {
	result := make(map[string]string)
	diags := input.ElementsAs(context.Background(), &result, false)
	if diags.HasError() {
		tflog.Warn(context.Background(), fmt.Sprintf("failed to convert input to map of strings: %s", diags))
	}
	return result
}

func AsListOfString(input types.List) []string {
	result := make([]string, 0)
	diags := input.ElementsAs(context.Background(), &result, false)
	if diags.HasError() {
		tflog.Warn(context.Background(), fmt.Sprintf("failed to convert input to list of strings: %s", diags))
	}
	return result
}

func AsMapOfLists(input types.Map) map[string][]string {
	result := make(map[string][]string)
	diags := input.ElementsAs(context.Background(), &result, false)
	if diags.HasError() {
		tflog.Warn(context.Background(), fmt.Sprintf("failed to convert input to map of lists: %s", diags))
	}
	return result
}

func unmarshalBody(input types.Dynamic, out interface{}) error {
	if input.IsNull() || input.IsUnknown() || input.IsUnderlyingValueUnknown() {
		return nil
	}
	data, err := dynamic.ToJSON(input)
	if err != nil {
		return fmt.Errorf(`invalid dynamic value: %s, err: %+v`, input.String(), err)
	}
	if err = json.Unmarshal(data, &out); err != nil {
		return fmt.Errorf(`unmarshaling failed: value: %s, err: %+v`, string(data), err)
	}
	return nil
}
