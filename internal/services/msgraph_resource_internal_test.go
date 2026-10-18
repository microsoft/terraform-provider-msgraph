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

type createStateCredential struct{}

func (*createStateCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "synthetic", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type createStateTransport func(*http.Request) (*http.Response, error)

func (f createStateTransport) Do(r *http.Request) (*http.Response, error) { return f(r) }

func createStateModel(t *testing.T, body string) *MSGraphResourceModel {
	t.Helper()
	value, err := dynamic.FromJSONImplied([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	return &MSGraphResourceModel{
		Id:                    types.StringUnknown(),
		ResourceUrl:           types.StringUnknown(),
		Url:                   types.StringValue("groups"),
		ApiVersion:            types.StringValue("v1.0"),
		Body:                  value,
		IgnoreMissingProperty: types.BoolValue(true),
		Retry:                 retry.NewValueNull(),
		Output:                types.DynamicUnknown(),
		CreateMethod:          types.StringNull(),
		UpdateMethod:          types.StringNull(),
		ResponseExportValues:  map[string]string{},
		CreateQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}),
		UpdateQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}),
		ReadQueryParameters:   types.MapNull(types.ListType{ElemType: types.StringType}),
		DeleteQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}),
		Timeouts:              timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "read": types.StringType, "delete": types.StringType})},
	}
}

func createStateState(t *testing.T, r *MSGraphResource, model *MSGraphResourceModel) tfsdk.State {
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

func TestMSGraphResourceCreatePreservesIdentityAfterReadFailure(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		url          string
		requestBody  string
		responseBody string
		location     string
		status       int
		wantID       string
		wantURL      string
	}{
		{
			name:         "POST response ID",
			method:       "POST",
			url:          "groups",
			requestBody:  `{"displayName":"example"}`,
			responseBody: `{"id":"created-child"}`,
			status:       201,
			wantID:       "created-child",
			wantURL:      "groups/created-child",
		},
		{
			name:         "POST Location",
			method:       "POST",
			url:          "groups",
			requestBody:  `{"displayName":"example"}`,
			responseBody: `{}`,
			location:     "https://graph.microsoft.com/v1.0/groups/created-child",
			status:       201,
			wantID:       "created-child",
			wantURL:      "groups/created-child",
		},
		{
			name:        "PUT empty response",
			method:      "PUT",
			url:         "groups/target-child",
			requestBody: `{"displayName":"example"}`,
			status:      204,
			wantID:      "target-child",
			wantURL:     "groups/target-child",
		},
		{
			name:        "POST $ref",
			method:      "POST",
			url:         "groups/parent/members/$ref",
			requestBody: `{"@odata.id":"https://graph.microsoft.com/v1.0/directoryObjects/child-id"}`,
			status:      204,
			wantID:      "child-id",
			wantURL:     "groups/parent/members/child-id",
		},
		{
			name:         "failed POST",
			method:       "POST",
			url:          "groups",
			requestBody:  `{"displayName":"example"}`,
			responseBody: `{"error":{"code":"BadRequest","message":"synthetic write failure"}}`,
			status:       400,
		},
		{
			name:        "POST $ref without @odata.id",
			method:      "POST",
			url:         "groups/parent/members/$ref",
			requestBody: `{"displayName":"not a ref"}`,
			status:      204,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				writes := 0
				client, err := clients.NewMSGraphClient(&createStateCredential{}, cloud.Configuration{}, &policy.ClientOptions{Transport: createStateTransport(func(req *http.Request) (*http.Response, error) {
					code, body := 400, `{"error":{"code":"BadRequest","message":"synthetic read failure"}}`
					headers := make(http.Header)
					if req.Method == tc.method {
						writes++
						code = tc.status
						body = tc.responseBody
						headers.Set("Location", tc.location)
					}
					return &http.Response{StatusCode: code, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
				})})
				if err != nil {
					t.Fatal(err)
				}
				r := &MSGraphResource{client: client}
				model := createStateModel(t, tc.requestBody)
				model.Url = types.StringValue(tc.url)
				if tc.method == "PUT" {
					model.CreateMethod = types.StringValue("PUT")
				}
				plan := createStateState(t, r, model)
				empty := createStateState(t, r, nil)
				resp := resource.CreateResponse{State: empty}
				r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(plan)}, &resp)
				if writes != 1 || !resp.Diagnostics.HasError() {
					t.Fatalf("writes = %d, diagnostics = %v", writes, resp.Diagnostics)
				}
				if tc.wantID == "" {
					if !resp.State.Raw.IsNull() {
						t.Fatal("failed create unexpectedly stored state")
					}
					return
				}
				if resp.State.Raw.IsNull() {
					t.Fatal("successful write followed by failed read left null state")
				}
				var got MSGraphResourceModel
				if diags := resp.State.Get(context.Background(), &got); diags.HasError() {
					t.Fatal(diags)
				}
				if got.Id.ValueString() != tc.wantID || got.ResourceUrl.ValueString() != tc.wantURL {
					t.Fatalf("state identity = %q, %q; want %q, %q", got.Id.ValueString(), got.ResourceUrl.ValueString(), tc.wantID, tc.wantURL)
				}
				if got.Output.IsUnknown() {
					t.Fatal("error state contains unknown output")
				}
			})
		})
	}
}
