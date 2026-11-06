package services

import (
	"context"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/microsoft/terraform-provider-msgraph/internal/dynamic"
)

func TestResolveResourceID(t *testing.T) {
	tests := []struct {
		name      string
		body      interface{}
		location  string
		want      string
		wantError *regexp.Regexp
	}{
		{
			name: "uses response id field",
			body: map[string]interface{}{"id": "00000000-0000-0000-0000-000000000000"},
			want: "00000000-0000-0000-0000-000000000000",
		},
		{
			name:     "response id wins over Location",
			body:     map[string]interface{}{"id": "00000000-0000-0000-0000-000000000000"},
			location: "https://graph.microsoft.com/v1.0/collection/22222222-2222-2222-2222-222222222222",
			want:     "00000000-0000-0000-0000-000000000000",
		},
		{
			name:     "derives id from Location header when body has no id (issue #107)",
			body:     map[string]interface{}{"tenantId": "544a7a2e-697f-487c-b2b0-a13df7f346b6"},
			location: "https://graph.microsoft.com/v1.0/79ac12ac-71ff-4533-a37b-08fdc0205d50/crossTenantAccessPolicyConfigurationPartners/544a7a2e-697f-487c-b2b0-a13df7f346b6",
			want:     "544a7a2e-697f-487c-b2b0-a13df7f346b6",
		},
		{
			name:     "Location header with query string and trailing slash",
			body:     map[string]interface{}{},
			location: "https://graph.microsoft.com/v1.0/things/abc123/?foo=bar",
			want:     "abc123",
		},
		{
			name:      "errors when nothing resolvable",
			body:      map[string]interface{}{"tenantId": "544a7a2e-697f-487c-b2b0-a13df7f346b6"},
			location:  "",
			wantError: regexp.MustCompile(`unable to determine the resource ID`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveResourceID(tt.body, tt.location)
			if tt.wantError != nil {
				if err == nil {
					t.Fatalf("expected error matching %q, got nil", tt.wantError.String())
				}
				if !tt.wantError.MatchString(err.Error()) {
					t.Fatalf("expected error matching %q, got %q", tt.wantError.String(), err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestReconcileReferenceIdOrder(t *testing.T) {
	tests := []struct {
		name     string
		previous []string
		current  []string
		want     []string
	}{
		{
			name:     "current returned in previous order",
			previous: []string{"a", "b", "c", "d"},
			current:  []string{"c", "a", "d", "b"},
			want:     []string{"a", "b", "c", "d"},
		},
		{
			name:     "new remote ids appended in current order",
			previous: []string{"a", "b"},
			current:  []string{"b", "c", "a", "d"},
			want:     []string{"a", "b", "c", "d"},
		},
		{
			name:     "removed remote ids dropped, previous order kept",
			previous: []string{"a", "b", "c"},
			current:  []string{"c", "a"},
			want:     []string{"a", "c"},
		},
		{
			name:     "case-insensitive match preserves current casing",
			previous: []string{"AAA", "BBB"},
			current:  []string{"bbb", "aaa"},
			want:     []string{"aaa", "bbb"},
		},
		{
			name:     "empty previous keeps current order",
			previous: nil,
			current:  []string{"b", "a", "c"},
			want:     []string{"b", "a", "c"},
		},
		{
			name:     "empty current yields empty result",
			previous: []string{"a", "b"},
			current:  nil,
			want:     []string{},
		},
		{
			name:     "duplicate in previous does not duplicate output",
			previous: []string{"a", "a", "b"},
			current:  []string{"b", "a"},
			want:     []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reconcileReferenceIdOrder(tt.previous, tt.current)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %#v, got %#v", tt.want, got)
			}
		})
	}
}

func TestRelationshipRefObjectID(t *testing.T) {
	mustDynamic := func(jsonBody string) types.Dynamic {
		v, err := dynamic.FromJSONImplied([]byte(jsonBody))
		if err != nil {
			t.Fatalf("failed to build dynamic value from %q: %s", jsonBody, err)
		}
		return v
	}

	tests := []struct {
		name string
		body types.Dynamic
		want string
	}{
		{
			name: "directoryObjects form",
			body: mustDynamic(`{"@odata.id":"https://graph.microsoft.com/v1.0/directoryObjects/00000000-0000-0000-0000-000000000001"}`),
			want: "00000000-0000-0000-0000-000000000001",
		},
		{
			name: "users form resolves to same id",
			body: mustDynamic(`{"@odata.id":"https://graph.microsoft.com/v1.0/users/00000000-0000-0000-0000-000000000001"}`),
			want: "00000000-0000-0000-0000-000000000001",
		},
		{
			name: "null body yields empty",
			body: types.DynamicNull(),
			want: "",
		},
		{
			name: "body without @odata.id yields empty",
			body: mustDynamic(`{"displayName":"not a ref"}`),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := relationshipRefObjectID(tt.body); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestMSGraphResourceURLRequiresReplacement(t *testing.T) {
	ctx := context.Background()
	r := &MSGraphResource{}
	response := resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, &response)
	url := response.Schema.Attributes["url"].(schema.StringAttribute)
	rawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"url": tftypes.String}}
	rawURL := func(v types.String) tftypes.Value {
		if v.IsUnknown() {
			return tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
		}
		return tftypes.NewValue(tftypes.String, v.ValueString())
	}

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
			stateValue := tftypes.NewValue(rawType, map[string]tftypes.Value{"url": rawURL(tc.old)})
			planValue := tftypes.NewValue(rawType, map[string]tftypes.Value{"url": rawURL(tc.next)})
			if tc.create {
				stateValue = tftypes.NewValue(rawType, nil)
			}
			if tc.destroy {
				planValue = tftypes.NewValue(rawType, nil)
			}
			req := planmodifier.StringRequest{
				Path:        path.Root("url"),
				StateValue:  tc.old,
				PlanValue:   tc.next,
				ConfigValue: tc.next,
				State:       tfsdk.State{Raw: stateValue},
				Plan:        tfsdk.Plan{Raw: planValue},
			}
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
