package myvalidator

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure interface compliance
var _ validator.String = resourceCollectionURL{}

// ResourceCollectionURL returns a validator that ensures a collection URL ends with '/$ref'
// and is a relative Microsoft Graph path (no leading slash or scheme).
func ResourceCollectionURL() validator.String { return resourceCollectionURL{} }

type resourceCollectionURL struct{}

func (v resourceCollectionURL) Description(ctx context.Context) string {
	return "Must be a relative Microsoft Graph collection reference URL ending in '/$ref' (e.g. groups/{id}/members/$ref)."
}

func (v resourceCollectionURL) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v resourceCollectionURL) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	val := req.ConfigValue.ValueString()

	if strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://") {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid URL",
			"URL must be relative (exclude https://graph.microsoft.com). Provide only the path, e.g. 'groups/{id}/members/$ref'.",
		))
		return
	}

	if strings.HasPrefix(val, "/") {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid URL",
			"URL must not start with a leading slash. Remove the leading '/'.",
		))
	}

	if !strings.HasSuffix(val, "/$ref") {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Missing '/$ref' suffix",
			"Collection URL must end with '/$ref' (e.g. 'groups/{id}/members/$ref').",
		))
		return
	}

	// ensure there's something before /$ref
	base := strings.TrimSuffix(val, "/$ref")
	if base == "" || base == "/" {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid collection path",
			"A path segment must precede '/$ref'. Example: 'groups/{id}/members/$ref'.",
		))
	}

	// simple guard: avoid double '/$ref'
	if strings.Count(val, "/$ref") > 1 {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid collection path",
			fmt.Sprintf("URL contains multiple '/$ref' segments: %s", val),
		))
	}

	// Optionally ensure no empty segments (e.g. consecutive //)
	if strings.Contains(val, "//") {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid collection path",
			"URL contains empty path segment (double '//').",
		))
	}

	// Provide warning if user appears to supply appRoleAssignments which we discourage here
	if strings.Contains(strings.ToLower(val), "approleassignments") {
		resp.Diagnostics.Append(diag.NewAttributeWarningDiagnostic(
			req.Path,
			"Potential unsupported collection",
			"'appRoleAssignments' is not intended to be managed by this resource. Ensure this is the desired usage.",
		))
	}

	// Validate that Terraform type is indeed string (defensive)
	if req.ConfigValue.Type(ctx) != types.StringType {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid type",
			"Expected a string value.",
		))
	}
}
