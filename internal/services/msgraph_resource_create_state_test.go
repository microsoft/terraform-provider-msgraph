package services

import (
	"context"
	"io"
	"net/http"
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
		Id: types.StringValue("child"), ResourceUrl: types.StringValue("groups/child"), Url: types.StringValue("groups"), ApiVersion: types.StringValue("v1.0"), Body: value, IgnoreMissingProperty: types.BoolValue(true), Retry: retry.NewValueNull(), Output: types.DynamicUnknown(), CreateMethod: types.StringNull(), UpdateMethod: types.StringNull(),
		ResponseExportValues:  map[string]string{},
		CreateQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}), UpdateQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}), ReadQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}), DeleteQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "read": types.StringType, "delete": types.StringType})},
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
	for _, tc := range []struct {
		name, method, body, location, wantID, wantURL string
		status                                        int
	}{
		{"POST response ID", "POST", `{"id":"created-child"}`, "", "created-child", "groups/created-child", 201},
		{"POST Location", "POST", `{}`, "https://graph.microsoft.com/v1.0/groups/created-child", "created-child", "groups/created-child", 201},
		{"PUT empty response", "PUT", "", "", "target-child", "groups/target-child", 204},
		{"failed POST", "POST", `{"error":{"code":"BadRequest","message":"synthetic write failure"}}`, "", "", "", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				writes := 0
				client, err := clients.NewMSGraphClient(&createStateCredential{}, cloud.Configuration{}, &policy.ClientOptions{Transport: createStateTransport(func(req *http.Request) (*http.Response, error) {
					code, body := 400, `{"error":{"code":"BadRequest","message":"synthetic read failure"}}`
					headers := make(http.Header)
					if req.Method == tc.method {
						writes++
						code = tc.status
						body = tc.body
						headers.Set("Location", tc.location)
					}
					return &http.Response{StatusCode: code, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
				})})
				if err != nil {
					t.Fatal(err)
				}
				r := &MSGraphResource{client: client}
				model := createStateModel(t, `{"displayName":"example"}`)
				model.Id = types.StringUnknown()
				model.ResourceUrl = types.StringUnknown()
				if tc.method == "PUT" {
					model.CreateMethod = types.StringValue("PUT")
					model.Url = types.StringValue("groups/target-child")
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
