package services_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/microsoft/terraform-provider-msgraph/internal/dynamic"
	"github.com/microsoft/terraform-provider-msgraph/internal/services"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance/check"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
	"github.com/microsoft/terraform-provider-msgraph/internal/utils"
)

type MSGraphTestUpdateResource struct{}

func TestAcc_UpdateResourceBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_update_resource", "test")

	r := MSGraphTestUpdateResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic("Demo App Updated"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("id").IsUUID(),
			),
		},
	})
}

func TestAcc_UpdateResourceUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_update_resource", "test")

	r := MSGraphTestUpdateResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic("Demo App Updated"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
		{
			Config: r.basic("Demo App Updated Again"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
	})
}

func TestAcc_UpdateResourceTimeouts_Update(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_update_resource", "test")
	r := MSGraphTestUpdateResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withUpdateTimeout("Demo App"),
		},
		{
			Config:      r.withUpdateTimeout("Demo App Updated"),
			ExpectError: regexp.MustCompile(`context deadline exceeded`),
		},
	})
}

func TestAcc_UpdateResourceTimeouts_Create(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_update_resource", "test")
	r := MSGraphTestUpdateResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:      r.withCreateTimeout("Demo App Updated"),
			ExpectError: regexp.MustCompile(`context deadline exceeded`),
		},
	})
}

func TestAcc_UpdateResourceTimeouts_Read(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_update_resource", "test")
	r := MSGraphTestUpdateResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:      r.withReadTimeout("Demo App Updated"),
			ExpectError: regexp.MustCompile(`context deadline exceeded`),
		},
	})
}

func TestAcc_UpdateResourceRetry(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_update_resource", "test")

	r := MSGraphTestUpdateResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withRetry(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
	})
}

func TestAcc_UpdateResource_GroupOwnerBind_UpdateDisplayName(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_update_resource", "test")
	r := MSGraphTestUpdateResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.groupWithOwnerUpdate("My Group Owners Bind 2"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
		{
			Config: r.groupWithOwnerUpdate("My Group Owners Bind 3"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
	})
}

func TestAcc_UpdateResourceWithPutUpdateMethod(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_update_resource", "test")

	r := MSGraphTestUpdateResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.updateMethod("Example Policy"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
		{
			Config: r.updateMethod("Updated Example Policy"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
	})
}

func (r MSGraphTestUpdateResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	apiVersion := state.Attributes["api_version"]
	url := state.Attributes["url"]

	_, err := client.MSGraphClient.Read(ctx, url, apiVersion, clients.DefaultRequestOptions())
	if err == nil {
		b := true
		return &b, nil
	}
	if utils.ResponseErrorWasNotFound(err) {
		b := false
		return &b, nil
	}
	return nil, fmt.Errorf("checking for presence of existing %s(api_version=%s) resource: %w", state.ID, apiVersion, err)
}

func (r MSGraphTestUpdateResource) basic(displayName string) string {
	return fmt.Sprintf(`
resource "msgraph_resource" "application" {
  url = "applications"
  body = {
    displayName = "Demo App"
  }

  lifecycle {
    ignore_changes = [body.displayName]
  }
}

resource "msgraph_update_resource" "test" {
  url = "applications/${msgraph_resource.application.id}"
  body = {
    displayName = "%s"
  }
}
`, displayName)
}

func (r MSGraphTestUpdateResource) withUpdateTimeout(displayName string) string {
	return fmt.Sprintf(`
%s

resource "msgraph_update_resource" "test" {
  url = "applications/${msgraph_resource.application.id}"
  body = {
    displayName = "%s"
  }
  timeouts {
    update = "1ns"
  }
}
`, MSGraphTestUpdateResource{}.applicationOnly(), displayName)
}

func (r MSGraphTestUpdateResource) withReadTimeout(displayName string) string {
	return fmt.Sprintf(`
%s

resource "msgraph_update_resource" "test" {
  url = "applications/${msgraph_resource.application.id}"
  body = {
    displayName = "%s"
  }
  timeouts {
    read = "1ns"
  }
}
`, MSGraphTestUpdateResource{}.applicationOnly(), displayName)
}

func (r MSGraphTestUpdateResource) withCreateTimeout(displayName string) string {
	return fmt.Sprintf(`
%s

resource "msgraph_update_resource" "test" {
  url = "applications/${msgraph_resource.application.id}"
  body = {
    displayName = "%s"
  }
  timeouts {
    create = "1ns"
  }
}
`, MSGraphTestUpdateResource{}.applicationOnly(), displayName)
}

// applicationOnly returns just the application resource to be used for composing
// different update resource configurations.
func (r MSGraphTestUpdateResource) applicationOnly() string {
	return `
resource "msgraph_resource" "application" {
  url = "applications"
  body = {
    displayName = "Demo App"
  }

  lifecycle {
    ignore_changes = [body.displayName]
  }
}
`
}

func (r MSGraphTestUpdateResource) withRetry() string {
	return `
resource "msgraph_resource" "application" {
  url = "applications"
  body = {
    displayName = "Demo App"
  }

  lifecycle {
    ignore_changes = [body.displayName]
  }
}

resource "msgraph_update_resource" "test" {
  url = "applications/${msgraph_resource.application.id}"
  body = {
    displayName = "Demo App Updated With Retry"
  }
  retry = {
    error_message_regex = [
      ".*throttl.*",
      "temporary error",
    ]
  }
}
`
}

func (r MSGraphTestUpdateResource) groupWithOwnerBase() string {
	return `
resource "msgraph_resource" "application" {
  url = "applications"
  body = {
    displayName = "My Application"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "servicePrincipal_application" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application.output.appId
  }
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "My Group Owners Bind"
    mailEnabled     = false
    mailNickname    = "mygroup-owners-bind"
    securityEnabled = true
    "owners@odata.bind" = [
      "https://graph.microsoft.com/v1.0/directoryObjects/${msgraph_resource.servicePrincipal_application.id}"
    ]
  }
  lifecycle {
    ignore_changes = [body.displayName]
  }
}
`
}

func (r MSGraphTestUpdateResource) groupWithOwnerUpdate(displayName string) string {
	return fmt.Sprintf(`
%s

resource "msgraph_update_resource" "test" {
  url = "groups/${msgraph_resource.group.id}"
  body = {
    displayName = "%s"
  }
}
`, r.groupWithOwnerBase(), displayName)
}

func (r MSGraphTestUpdateResource) updateMethod(displayName string) string {
	return fmt.Sprintf(`
resource "msgraph_resource" "group_example" {
  url = "groups"
  body = {
    displayName     = "group-name"
    mailEnabled     = false
    mailNickname    = "group-name"
    securityEnabled = true
  }
}

resource "msgraph_resource" "catalog_example" {
  url = "identityGovernance/entitlementManagement/catalogs"
  body = {
    displayName = "example-catalog"
    description = "Example catalog"
  }
}

resource "msgraph_resource" "access_package_example" {
  url         = "identityGovernance/entitlementManagement/accessPackages"
  api_version = "beta"
  body = {
    catalogId   = msgraph_resource.catalog_example.id
    displayName = "access-package"
    description = "Access Package"
  }
}

resource "msgraph_resource" "access_package_assignment_policy" {
  url           = "identityGovernance/entitlementManagement/accessPackageAssignmentPolicies"
  api_version   = "beta"
  update_method = "PUT"
  body = {
    accessPackageId = msgraph_resource.access_package_example.id
    displayName     = "init displayname"
    description     = "My assignment"
    expiration = {
      type     = "afterDuration"
      duration = "P90D"
    }
    requestorSettings = {
      scopeType = "AllExistingDirectoryMemberUsers"
    }
    requestApprovalSettings = {
      isApprovalRequired = true
      approvalStages = [
        {
          approvalStageTimeOutInDays = 14
          primaryApprovers = [
            {
              "@odata.type" = "#microsoft.graph.groupMembers"
              groupId       = msgraph_resource.group_example.id
              description   = "group-name"
            }
          ]
        }
      ]
    }
    reviewSettings = {
      isEnabled          = true
      expirationBehavior = "keepAccess"
      isSelfReview       = true
      schedule = {
        startDateTime = "2025-12-12T00:00:00Z"
        recurrence = {
          pattern = {
            type     = "weekly"
            interval = 1
          }
          range = {
            type      = "noEnd"
            startDate = "2025-12-12"
          }
        }
      }
    }
    questions = [
      {
        "@odata.type" = "#microsoft.graph.accessPackageTextInputQuestion"
        text = {
          defaultText = "hello, how are you?"
        }
        isRequired = false
      }
    ]
  }

  lifecycle {
    ignore_changes = [body.displayName, body.description]
  }
}

resource "msgraph_update_resource" "test" {
  url           = "identityGovernance/entitlementManagement/accessPackageAssignmentPolicies/${msgraph_resource.access_package_assignment_policy.id}"
  api_version   = "beta"
  update_method = "PUT"
  body = {
    displayName = "%[1]s"
    description = "My assignment %[1]s"
  }
}
`, displayName)
}

type assignmentTransport func(*http.Request) (*http.Response, error)

func (f assignmentTransport) Do(req *http.Request) (*http.Response, error) { return f(req) }

type assignmentCredential struct{}

func (assignmentCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "test", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

func TestUpdateResourceAssignments(t *testing.T) {
	ctx := context.Background()
	r := services.NewMSGraphUpdateResource().(*services.MSGraphUpdateResource)
	var schemaResponse frameworkresource.SchemaResponse
	r.Schema(ctx, frameworkresource.SchemaRequest{}, &schemaResponse)
	s := schemaResponse.Schema
	attrs := map[string]tftypes.Value{}
	for name, typ := range s.Type().TerraformType(ctx).(tftypes.Object).AttributeTypes {
		attrs[name] = tftypes.NewValue(typ, nil)
	}
	state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), attrs)}
	var model services.MSGraphUpdateResourceModel
	if d := state.Get(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	parse := func(input string) types.Dynamic {
		t.Helper()
		v, err := dynamic.FromJSONImplied([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	assertJSON := func(got, want string) {
		t.Helper()
		var a, b interface{}
		if err := json.Unmarshal([]byte(got), &a); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(want), &b); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("got %s, want %s", got, want)
		}
	}
	model.Url = types.StringValue("policies/policy/microsoft.graph.assign")
	model.ReadUrl = types.StringValue("policies/policy/assignments")
	model.ApiVersion = types.StringValue("beta")
	model.UpdateMethod = types.StringValue("POST")
	model.IgnoreMissingProperty = types.BoolValue(true)
	payload := `{"assignments":[{"target":{"@odata.type":"#microsoft.graph.groupAssignmentTarget","groupId":"group"}}]}`
	model.Body = parse(payload)
	model.DeleteBody = parse(`{"assignments":[]}`)
	response := `{"@odata.context":"metadata","value":[{"id":"server","sourceId":"source","target":{"@odata.type":"#microsoft.graph.groupAssignmentTarget","groupId":"group"}}]}`
	status := http.StatusOK
	var calls []string
	client, err := clients.NewMSGraphClient(assignmentCredential{}, cloud.Configuration{}, &policy.ClientOptions{
		Retry: policy.RetryOptions{MaxRetries: -1},
		Transport: assignmentTransport(func(req *http.Request) (*http.Response, error) {
			calls = append(calls, req.Method+" "+req.URL.Path)
			body := response
			if req.Method != http.MethodGet {
				data, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatal(err)
				}
				assertJSON(string(data), payload)
				body = ""
			}
			return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), ContentLength: int64(len(body)), Request: req}, nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	r.Configure(ctx, frameworkresource.ConfigureRequest{ProviderData: &clients.Client{MSGraphClient: client}}, &frameworkresource.ConfigureResponse{})
	if d := state.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	created := frameworkresource.CreateResponse{State: state}
	r.Create(ctx, frameworkresource.CreateRequest{Plan: tfsdk.Plan{Schema: s, Raw: state.Raw}}, &created)
	if created.Diagnostics.HasError() {
		t.Fatal(created.Diagnostics)
	}
	if want := []string{"POST /beta/policies/policy/microsoft.graph.assign", "GET /beta/policies/policy/assignments"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls: %v", calls)
	}
	state = created.State
	if d := state.Get(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	if model.Id.ValueString() != "assignments" {
		t.Fatalf("ID: %s", model.Id)
	}
	for _, tc := range []struct{ name, response, want string }{
		{"metadata stripped", response, payload},
		{"portal edit", strings.ReplaceAll(response, `"group"`, `"changed"`), strings.ReplaceAll(payload, `"group"`, `"changed"`)},
		{"portal unassignment", `{"value":[]}`, `{"assignments":[]}`},
		{"noncollection value", `{"value":null}`, payload},
		{"named key wins", `{"value":[],"assignments":[{"target":{"@odata.type":"#microsoft.graph.groupAssignmentTarget","groupId":"explicit"}}]}`, strings.ReplaceAll(payload, `"group"`, `"explicit"`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response = tc.response
			read := frameworkresource.ReadResponse{State: state}
			r.Read(ctx, frameworkresource.ReadRequest{State: state}, &read)
			if read.Diagnostics.HasError() {
				t.Fatal(read.Diagnostics)
			}
			var refreshed services.MSGraphUpdateResourceModel
			if d := read.State.Get(ctx, &refreshed); d.HasError() {
				t.Fatal(d)
			}
			got, err := dynamic.ToJSON(refreshed.Body)
			if err != nil {
				t.Fatal(err)
			}
			assertJSON(string(got), tc.want)
		})
	}
	payload = `{"assignments":[]}`
	for _, code := range []int{200, 404, 400} {
		status = code
		deleted := frameworkresource.DeleteResponse{State: state}
		r.Delete(ctx, frameworkresource.DeleteRequest{State: state}, &deleted)
		if deleted.Diagnostics.HasError() != (code == 400) {
			t.Fatalf("status %d: %v", code, deleted.Diagnostics)
		}
		if calls[len(calls)-1] != "POST /beta/policies/policy/microsoft.graph.assign" {
			t.Fatal(calls)
		}
	}
	status = http.StatusOK
	for _, v := range []types.Dynamic{types.DynamicNull(), types.DynamicUnknown()} {
		model.DeleteBody = v
		if d := state.Set(ctx, &model); d.HasError() {
			t.Fatal(d)
		}
		before := len(calls)
		deleted := frameworkresource.DeleteResponse{State: state}
		r.Delete(ctx, frameworkresource.DeleteRequest{State: state}, &deleted)
		if deleted.Diagnostics.HasError() || len(calls) != before {
			t.Fatalf("unexpected cleanup: %v", deleted.Diagnostics)
		}
	}
	for _, v := range []types.String{types.StringNull(), types.StringValue("")} {
		model.ReadUrl = v
		if d := state.Set(ctx, &model); d.HasError() {
			t.Fatal(d)
		}
		read := frameworkresource.ReadResponse{State: state}
		r.Read(ctx, frameworkresource.ReadRequest{State: state}, &read)
		if read.Diagnostics.HasError() {
			t.Fatal(read.Diagnostics)
		}
		if calls[len(calls)-1] != "GET /beta/policies/policy/microsoft.graph.assign" {
			t.Fatal(calls)
		}
	}
}
