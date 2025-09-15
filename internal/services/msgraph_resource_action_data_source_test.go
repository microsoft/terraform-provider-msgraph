package services_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
)

type MSGraphResourceActionDataSourceTestResource struct{}

func (MSGraphResourceActionDataSourceTestResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	exists := true
	return &exists, nil
}

func TestAcc_DataSourceResourceActionBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_action", "test")

	r := MSGraphResourceActionDataSourceTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}

func TestAcc_DataSourceResourceActionWithQueryParams(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_action", "test")

	r := MSGraphResourceActionDataSourceTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withQueryParams(),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}

func TestAcc_DataSourceResourceActionWithHeaders(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_action", "test")

	r := MSGraphResourceActionDataSourceTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withHeaders(),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}

func TestAcc_DataSourceResourceActionWithBody(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_action", "test")

	r := MSGraphResourceActionDataSourceTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withBody(),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}

func (r MSGraphResourceActionDataSourceTestResource) basic() string {
	return `
provider "msgraph" {}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Test Group"
    mailEnabled     = false
    mailNickname    = "mygroup"
    securityEnabled = true
  }

  lifecycle {
    ignore_changes = [body.displayName]
  }
}

data "msgraph_resource_action" "test" {
  resource_url = msgraph_resource.group.resource_url
  action       = "members"
  method       = "GET"
}
`
}

func (r MSGraphResourceActionDataSourceTestResource) withQueryParams() string {
	return `
provider "msgraph" {}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Test Group"
    mailEnabled     = false
    mailNickname    = "mygroup"
    securityEnabled = true
  }

  lifecycle {
    ignore_changes = [body.displayName]
  }
}

data "msgraph_resource_action" "test" {
  resource_url = msgraph_resource.group.resource_url
  action       = "owners"
  method       = "GET"

  query_parameters = {
    "$select" = ["id", "displayName"]
    "$top"    = ["5"]
  }
}
`
}

func (r MSGraphResourceActionDataSourceTestResource) withHeaders() string {
	return `
provider "msgraph" {}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Test Group"
    mailEnabled     = false
    mailNickname    = "mygroup"
    securityEnabled = true
  }

  lifecycle {
    ignore_changes = [body.displayName]
  }
}

data "msgraph_resource_action" "test" {
  resource_url = msgraph_resource.group.resource_url
  action       = "owners"
  method       = "GET"

  headers = {
    "X-Custom-Header" = "test-value"
  }
}
`
}

func (r MSGraphResourceActionDataSourceTestResource) withBody() string {
	return `
provider "msgraph" {}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Test Group"
    mailEnabled     = false
    mailNickname    = "mygroup"
    securityEnabled = true
  }

  lifecycle {
    ignore_changes = [body.displayName]
  }
}

data "msgraph_resource_action" "test" {
  resource_url = msgraph_resource.group.resource_url
  action       = "checkMemberObjects"
  method       = "POST"

  body = {
    ids = ["00000000-0000-0000-0000-000000000000"]
  }

  response_export_values = {
    result = "value"
  }
}
`
}
