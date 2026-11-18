package services

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
	"github.com/microsoft/terraform-provider-msgraph/internal/dynamic"
	"github.com/microsoft/terraform-provider-msgraph/internal/retry"
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

type nullPatchCredential struct{}

func (*nullPatchCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "synthetic", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type nullPatchTransport func(*http.Request) (*http.Response, error)

func (f nullPatchTransport) Do(r *http.Request) (*http.Response, error) { return f(r) }

func nullPatchModel(t *testing.T, body string) *MSGraphResourceModel {
	t.Helper()
	value, err := dynamic.FromJSONImplied([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	emptyQuery := types.MapNull(types.ListType{ElemType: types.StringType})
	return &MSGraphResourceModel{
		Id:                    types.StringValue("child"),
		ResourceUrl:           types.StringValue("groups/child"),
		Url:                   types.StringValue("groups"),
		ApiVersion:            types.StringValue("v1.0"),
		Body:                  value,
		IgnoreMissingProperty: types.BoolValue(true),
		Retry:                 retry.NewValueNull(),
		Output:                types.DynamicUnknown(),
		CreateMethod:          types.StringNull(),
		UpdateMethod:          types.StringNull(),
		ResponseExportValues:  map[string]string{},
		CreateQueryParameters: emptyQuery,
		UpdateQueryParameters: emptyQuery,
		ReadQueryParameters:   emptyQuery,
		DeleteQueryParameters: emptyQuery,
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"update": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		})},
	}
}

func nullPatchState(t *testing.T, r *MSGraphResource, model *MSGraphResourceModel) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schema := resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	if diags := state.Set(ctx, model); diags.HasError() {
		t.Fatal(diags)
	}
	return state
}

func TestMSGraphResourceUpdateExplicitNull(t *testing.T) {
	for _, tc := range []struct {
		name, before, after, want string
		writes                    int
	}{
		{"clear", `{"description":"old"}`, `{"description":null}`, `{"description":null}`, 1},
		{"already null", `{"description":null}`, `{"description":null}`, "", 0},
		{"omitted field", `{"description":"old"}`, `{}`, "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				writes := 0
				transport := nullPatchTransport(func(req *http.Request) (*http.Response, error) {
					if req.Method == http.MethodPatch {
						writes++
						data, readErr := io.ReadAll(req.Body)
						if readErr != nil {
							t.Fatal(readErr)
						}
						if strings.TrimSpace(string(data)) != tc.want {
							t.Errorf("PATCH body = %s, want %s", data, tc.want)
						}
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader(`{"id":"child","description":null}`)),
						Request:    req,
					}, nil
				})
				client, err := clients.NewMSGraphClient(&nullPatchCredential{}, cloud.Configuration{}, &policy.ClientOptions{Transport: transport})
				if err != nil {
					t.Fatal(err)
				}
				r := &MSGraphResource{client: client}
				before := nullPatchState(t, r, nullPatchModel(t, tc.before))
				after := nullPatchState(t, r, nullPatchModel(t, tc.after))
				resp := resource.UpdateResponse{State: before}
				r.Update(context.Background(), resource.UpdateRequest{State: before, Plan: tfsdk.Plan(after)}, &resp)
				if resp.Diagnostics.HasError() {
					t.Fatal(resp.Diagnostics)
				}
				if writes != tc.writes {
					t.Fatalf("PATCH requests = %d, want %d", writes, tc.writes)
				}
			})
		})
	}
}
