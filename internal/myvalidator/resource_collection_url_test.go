package myvalidator

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestResourceCollectionURL_ValidateString(t *testing.T) {
	v := resourceCollectionURL{}

	cases := []struct {
		name      string
		value     string
		wantError bool
		wantWarn  bool
	}{
		{"valid_basic", "groups/123/members/$ref", false, false},
		{"valid_nested", "groups/123/owners/$ref", false, false},
		{"missing_ref", "groups/123/members", true, false},
		{"leading_slash", "/groups/123/members/$ref", true, false},
		{"absolute_url", "https://graph.microsoft.com/v1.0/groups/123/members/$ref", true, false},
		{"double_ref", "groups/123/members/$ref/$ref", true, false},
		{"empty_before_ref", "/$ref", true, false},
		{"double_slash", "groups//123/members/$ref", true, false},
		{"approleassignments_warn", "servicePrincipals/123/appRoleAssignments/$ref", false, true},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: basetypes.NewStringValue(c.value),
				Path:        path.Empty(),
			}
			resp := &validator.StringResponse{Diagnostics: diag.Diagnostics{}}
			v.ValidateString(context.Background(), req, resp)

			hasErr := resp.Diagnostics.HasError()
			if hasErr != c.wantError {
				t.Fatalf("error expectation mismatch: got error=%v want=%v diagnostics=%v", hasErr, c.wantError, resp.Diagnostics)
			}

			// check warning presence if requested
			foundWarn := false
			for _, d := range resp.Diagnostics {
				if d.Severity() == diag.SeverityWarning {
					foundWarn = true
					break
				}
			}
			if foundWarn != c.wantWarn {
				t.Fatalf("warning expectation mismatch: got warn=%v want=%v diagnostics=%v", foundWarn, c.wantWarn, resp.Diagnostics)
			}
		})
	}
}
