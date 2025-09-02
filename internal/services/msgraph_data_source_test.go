package services_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance/check"
)

type MSGraphTestDataSource struct{}

func TestAcc_DataSourceBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.msgraph_resource", "test")
	r := MSGraphTestDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("output.%").Exists(),
				check.That(data.ResourceName).Key("id").IsUUID(),
			),
		},
	})
}

func TestAcc_DataSourceQueryParameters(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.msgraph_resource", "test")
	r := MSGraphTestDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.query(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("output.%").Exists(),
			),
		},
	})
}

func TestAcc_DataSourceList(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.msgraph_resource", "test")
	r := MSGraphTestDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.list(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("output.%").Exists(),
			),
		},
	})
}

func (r MSGraphTestDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "msgraph_resource" "test" {
  url = "applications/${msgraph_resource.test.id}"
}
`, MSGraphTestResource{}.basic(data))
}

func (r MSGraphTestDataSource) query(data acceptance.TestData) string {
	return `
locals {
  MicrosoftGraphAppId = "00000003-0000-0000-c000-000000000000"
}

data "msgraph_resource" "test" {
  url = "servicePrincipals"
  query_parameters = {
    "$filter" = ["appId eq '${local.MicrosoftGraphAppId}'"]
  }
  response_export_values = {
    all = "@"
  }
}`
}

func (r MSGraphTestDataSource) list(data acceptance.TestData) string {
	return `
data "msgraph_resource" "test" {
  url = "groups"
  response_export_values = {
    all = "@"
  }
}`
}
