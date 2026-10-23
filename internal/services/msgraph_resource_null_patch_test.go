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
	return &MSGraphResourceModel{
		Id: types.StringValue("child"), ResourceUrl: types.StringValue("groups/child"), Url: types.StringValue("groups"), ApiVersion: types.StringValue("v1.0"), Body: value, IgnoreMissingProperty: types.BoolValue(true), Retry: retry.NewValueNull(), Output: types.DynamicUnknown(), CreateMethod: types.StringNull(), UpdateMethod: types.StringNull(),
		ResponseExportValues:  map[string]string{},
		CreateQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}), UpdateQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}), ReadQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}), DeleteQueryParameters: types.MapNull(types.ListType{ElemType: types.StringType}),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "read": types.StringType, "delete": types.StringType})},
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
				client, err := clients.NewMSGraphClient(&nullPatchCredential{}, cloud.Configuration{}, &policy.ClientOptions{Transport: nullPatchTransport(func(req *http.Request) (*http.Response, error) {
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
					return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"child","description":null}`)), Request: req}, nil
				})})
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
