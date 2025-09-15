package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance/check"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
)

type MSGraphResourceActionTestResource struct{}

func TestAcc_ResourceActionBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_action", "test")

	r := MSGraphResourceActionTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}

func TestAcc_ResourceActionWithQueryParams(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_action", "test")

	r := MSGraphResourceActionTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withQueryParams(data),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}

func TestAcc_ResourceActionWithHeaders(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_action", "test")

	r := MSGraphResourceActionTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withHeaders(data),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}

func TestAcc_ResourceActionWithExportValues(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_action", "test")

	r := MSGraphResourceActionTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withExportValues(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("output.%").HasValue("2"),
			),
		},
	})
}

func (r MSGraphResourceActionTestResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	exists := false
	return &exists, nil
}

func (r MSGraphResourceActionTestResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
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

resource "msgraph_resource_action" "test" {
  resource_url = msgraph_resource.group.resource_url
  method       = "PATCH"

  body = {
    displayName = "Updated Group Name"
  }
}
`)
}

func (r MSGraphResourceActionTestResource) withQueryParams(data acceptance.TestData) string {
	return fmt.Sprintf(`
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

resource "msgraph_resource_action" "test" {
  resource_url = msgraph_resource.group.resource_url
  method       = "PATCH"

  query_parameters = {
    "$select" = ["id", "displayName"]
  }

  body = {
    displayName = "Updated Group Name with Query Params"
  }
}
`)
}

func (r MSGraphResourceActionTestResource) withHeaders(data acceptance.TestData) string {
	return fmt.Sprintf(`
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

resource "msgraph_resource_action" "test" {
  resource_url = msgraph_resource.group.resource_url
  method       = "PATCH"

  headers = {
    "X-Custom-Header" = "test-value"
    "X-Request-ID"    = "test-123"
  }

  body = {
    displayName = "Updated Group Name with Headers"
  }
}
`)
}

func (r MSGraphResourceActionTestResource) withExportValues(data acceptance.TestData) string {
	return fmt.Sprintf(`
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

resource "msgraph_resource_action" "test" {
  resource_url = msgraph_resource.group.resource_url
  method       = "PATCH"

  body = {
    displayName = "Updated Group Name with Export Values"
  }

  response_export_values = {
    group_id   = "id"
    group_name = "displayName"
  }
}
`)
}
