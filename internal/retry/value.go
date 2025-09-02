package retry

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.ObjectValuable = Value{}

func NewValueNull() Value {
	return Value{
		state: attr.ValueStateNull,
	}
}

func NewValueUnknown() Value {
	return Value{
		state: attr.ValueStateUnknown,
	}
}

func NewValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing RetryValue Attribute Value",
				"While creating a RetryValue value, a missing attribute value was detected. "+
					"A RetryValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("RetryValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid RetryValue Attribute Type",
				"While creating a RetryValue value, an invalid attribute value was detected. "+
					"A RetryValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("RetryValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("RetryValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra RetryValue Attribute Value",
				"While creating a RetryValue value, an extra attribute value was detected. "+
					"A RetryValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra RetryValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewValueUnknown(), diags
	}

	errorMessageRegexAttribute, ok := attributes["error_message_regex"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`error_message_regex is missing from object`)

		return NewValueUnknown(), diags
	}

	errorMessageRegexVal, ok := errorMessageRegexAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`error_message_regex expected to be basetypes.ListValue, was: %T`, errorMessageRegexAttribute))
	}

	if diags.HasError() {
		return NewValueUnknown(), diags
	}

	return Value{
		ErrorMessageRegex: errorMessageRegexVal,
		state:             attr.ValueStateKnown,
	}, diags
}

func NewRetryValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) Value {
	object, diags := NewValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewRetryValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t Type) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewValueUnknown(), nil
	}

	if in.IsNull() {
		return NewValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)
	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)
		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewRetryValueMust(Value{}.AttributeTypes(ctx), attributes), nil
}

func (t Type) ValueType(ctx context.Context) attr.Value {
	return Value{}
}

var _ basetypes.ObjectValuable = Value{}

type Value struct {
	ErrorMessageRegex basetypes.ListValue `tfsdk:"error_message_regex"`
	state             attr.ValueState
}

func (v Value) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 5)

	var val tftypes.Value
	var err error

	attrTypes["error_message_regex"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 5)

		val, err = v.ErrorMessageRegex.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["error_message_regex"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v Value) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v Value) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v Value) String() string {
	return "RetryValue"
}

func (v Value) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	var errorMessageRegexVal basetypes.ListValue
	switch {
	case v.ErrorMessageRegex.IsUnknown():
		errorMessageRegexVal = types.ListUnknown(types.StringType)
	case v.ErrorMessageRegex.IsNull():
		errorMessageRegexVal = types.ListNull(types.StringType)
	default:
		var d diag.Diagnostics
		errorMessageRegexVal, d = types.ListValue(types.StringType, v.ErrorMessageRegex.Elements())
		diags.Append(d...)
	}

	if diags.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"error_message_regex": basetypes.ListType{
				ElemType: types.StringType,
			},
		}), diags
	}

	attributeTypes := map[string]attr.Type{
		"error_message_regex": basetypes.ListType{
			ElemType: types.StringType,
		},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"error_message_regex": errorMessageRegexVal,
		})

	return objVal, diags
}

func (v Value) Equal(o attr.Value) bool {
	other, ok := o.(Value)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.ErrorMessageRegex.Equal(other.ErrorMessageRegex) {
		return false
	}

	return true
}

func (v Value) Type(ctx context.Context) attr.Type {
	return Type{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v Value) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"error_message_regex": basetypes.ListType{
			ElemType: types.StringType,
		},
	}
}

func (v Value) GetErrorMessages() []string {
	if v.IsNull() {
		return nil
	}
	if v.IsUnknown() {
		return nil
	}
	res := make([]string, len(v.ErrorMessageRegex.Elements()))
	for i, elem := range v.ErrorMessageRegex.Elements() {
		res[i] = elem.(types.String).ValueString()
	}
	return res
}

func (v Value) GetErrorMessagesRegex() []regexp.Regexp {
	msgs := v.GetErrorMessages()
	if msgs == nil {
		return nil
	}
	res := make([]regexp.Regexp, len(msgs))
	for i, msg := range msgs {
		res[i] = *regexp.MustCompile(msg)
	}
	return res
}
