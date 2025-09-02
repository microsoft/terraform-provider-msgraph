package retry

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ basetypes.ObjectTypable = Type{}

type Type struct {
	basetypes.ObjectType
}

func (t Type) Equal(o attr.Type) bool {
	other, ok := o.(Type)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t Type) String() string {
	return "RetryType"
}

func (t Type) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	errorMessageRegexAttribute, ok := attributes["error_message_regex"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`error_message_regex is missing from object`)

		return nil, diags
	}

	errorMessageRegexVal, ok := errorMessageRegexAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`error_message_regex expected to be basetypes.ListValue, was: %T`, errorMessageRegexAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return Value{
		ErrorMessageRegex: errorMessageRegexVal,
		state:             attr.ValueStateKnown,
	}, diags
}
