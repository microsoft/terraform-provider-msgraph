package services

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestMSGraphResourceURLRequiresReplacement(t *testing.T) {
	ctx := context.Background()
	r := &MSGraphResource{}
	response := resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, &response)
	url := response.Schema.Attributes["url"].(schema.StringAttribute)
	for _, tc := range []struct {
		name                  string
		old, next             types.String
		create, destroy, want bool
	}{
		{name: "collection parent changes", old: types.StringValue("applications/old/federatedIdentityCredentials"), next: types.StringValue("applications/new/federatedIdentityCredentials"), want: true},
		{name: "relationship parent changes", old: types.StringValue("groups/old/members/$ref"), next: types.StringValue("groups/new/members/$ref"), want: true},
		{name: "PUT item URL changes", old: types.StringValue("policies/old"), next: types.StringValue("policies/new"), want: true},
		{name: "unchanged URL", old: types.StringValue("groups"), next: types.StringValue("groups")},
		{name: "unknown parent", old: types.StringValue("groups/old/members/$ref"), next: types.StringUnknown(), want: true},
		{name: "new resource", old: types.StringNull(), next: types.StringValue("groups"), create: true},
		{name: "destroy", old: types.StringValue("groups"), next: types.StringNull(), destroy: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"url": tftypes.String}}
			stateValue := tftypes.NewValue(rawType, map[string]tftypes.Value{"url": tftypes.NewValue(tftypes.String, tc.old.ValueString())})
			planValue := tftypes.NewValue(rawType, map[string]tftypes.Value{"url": tftypes.NewValue(tftypes.String, tc.next.ValueString())})
			if tc.create {
				stateValue = tftypes.NewValue(rawType, nil)
			}
			if tc.destroy {
				planValue = tftypes.NewValue(rawType, nil)
			}
			req := planmodifier.StringRequest{Path: path.Root("url"), StateValue: tc.old, PlanValue: tc.next, ConfigValue: tc.next, State: tfsdk.State{Raw: stateValue}, Plan: tfsdk.Plan{Raw: planValue}}
			replace := false
			for _, modifier := range url.PlanModifiers {
				resp := planmodifier.StringResponse{PlanValue: tc.next}
				modifier.PlanModifyString(ctx, req, &resp)
				if resp.Diagnostics.HasError() {
					t.Fatal(resp.Diagnostics)
				}
				replace = replace || resp.RequiresReplace
			}
			if replace != tc.want {
				t.Fatalf("requires replacement = %t, want %t", replace, tc.want)
			}
		})
	}
}
