package services_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance/check"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
	"github.com/microsoft/terraform-provider-msgraph/internal/utils"
)

func defaultIgnores() []string {
	return []string{"body", "output", "retry"}
}

type MSGraphTestResource struct{}

func TestAcc_ResourceBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")

	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("id").IsUUID(),
			),
		},
		data.ImportStepWithImportStateIdFunc(r.ImportIdFunc, defaultIgnores()...),
	})
}

func TestAcc_ResourceUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")

	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
		data.ImportStepWithImportStateIdFunc(r.ImportIdFunc, defaultIgnores()...),
		{
			Config: r.basicUpdate(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
		data.ImportStepWithImportStateIdFunc(r.ImportIdFunc, defaultIgnores()...),
	})
}

func TestAcc_ResourceGroupMember(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")

	r := MSGraphTestResource{}

	importStep := data.ImportStepWithImportStateIdFunc(r.ImportIdFunc, defaultIgnores()...)
	importStep.ImportStateVerify = false
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.groupMember(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("id").IsUUID(),
				check.That(data.ResourceName).Key("id").MatchesOtherKey(check.That("msgraph_resource.servicePrincipal_application").Key("id")),
			),
		},
		importStep,
	})
}

func TestAcc_ResourceRetry(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")

	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withRetry(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("id").IsUUID(),
			),
		},
		data.ImportStepWithImportStateIdFunc(r.ImportIdFunc, defaultIgnores()...),
	})
}

func TestAcc_ResourceTimeouts_Create(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")

	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withCreateTimeout(),
			// Creating with 1ns should fail quickly with a deadline exceeded error
			ExpectError: regexp.MustCompile(`context deadline exceeded`),
		},
	})
}

func TestAcc_ResourceTimeouts_Update(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")

	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
		{
			Config:      r.withUpdateTimeout(),
			ExpectError: regexp.MustCompile(`context deadline exceeded`),
		},
	})
}

func (r MSGraphTestResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	apiVersion := state.Attributes["api_version"]
	url := state.Attributes["url"]

	var checkUrl string
	if !strings.Contains(url, "/$ref") {
		checkUrl = fmt.Sprintf("%s/%s", url, state.ID)
	} else {
		checkUrl = url
	}

	_, err := client.MSGraphClient.Read(ctx, checkUrl, apiVersion, clients.DefaultRequestOptions())
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

func (r MSGraphTestResource) ImportIdFunc(tfState *terraform.State) (string, error) {
	state := tfState.RootModule().Resources["msgraph_resource.test"].Primary
	url := state.Attributes["url"]
	if !strings.Contains(url, "/$ref") {
		return fmt.Sprintf("%s/%s", url, state.ID), nil
	}
	return strings.ReplaceAll(url, "/$ref", fmt.Sprintf("/%s/$ref", state.ID)), nil
}

func (r MSGraphTestResource) basic(data acceptance.TestData) string {
	return `
resource "msgraph_resource" "test" {
  url = "applications"
  body = {
    displayName = "Demo App"
  }
}
`
}

func (r MSGraphTestResource) basicUpdate(data acceptance.TestData) string {
	return `
resource "msgraph_resource" "test" {
  url = "applications"
  body = {
    displayName = "Demo App Updated"
  }
}
`
}

func (r MSGraphTestResource) groupMember(data acceptance.TestData) string {
	return `
resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "My Group"
    mailEnabled     = false
    mailNickname    = "mygroup"
    securityEnabled = true
  }
}

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

resource "msgraph_resource" "test" {
  # url = "groups/{group-id}/members/$ref"
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/directoryObjects/${msgraph_resource.servicePrincipal_application.id}"
  }
}
`
}

func (r MSGraphTestResource) withRetry(data acceptance.TestData) string {
	return `
resource "msgraph_resource" "test" {
  url = "applications"
  body = {
    displayName = "Demo App Retry"
  }
  retry = {
    error_message_regex = [
      "temporary error",
      ".*throttl.*",
    ]
  }
}`
}

func (r MSGraphTestResource) withCreateTimeout() string {
	return `
resource "msgraph_resource" "test" {
  url = "applications"
  timeouts {
    create = "1ns"
  }
  body = {
    displayName = "Demo App Timeout Create"
  }
}
`
}

func (r MSGraphTestResource) withUpdateTimeout() string {
	return `
resource "msgraph_resource" "test" {
  url = "applications"
  timeouts {
    update = "1ns"
  }
  body = {
    displayName = "Demo App Updated Timeout Update"
  }
}
`
}
