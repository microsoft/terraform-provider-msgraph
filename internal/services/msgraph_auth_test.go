package services_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance"
)

// TestAccAuth_clientSecret tests authentication using client secret
func TestAccAuth_clientSecret(t *testing.T) {
	if ok := os.Getenv("ARM_CLIENT_SECRET"); ok == "" {
		t.Skip("Skipping as `ARM_CLIENT_SECRET` is not specified")
	}

	data := acceptance.BuildTestData(t, "data.msgraph_resource", "test")
	r := MSGraphTestDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.basic(data),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}

// TestAccAuth_azureCLI tests authentication using Azure CLI
func TestAccAuth_azureCLI(t *testing.T) {
	if ok := os.Getenv("ARM_USE_CLI"); ok == "" {
		t.Skip("Skipping as `ARM_USE_CLI` is not specified")
	}

	data := acceptance.BuildTestData(t, "data.msgraph_resource", "test")
	r := MSGraphTestDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.basic(data),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}

// TestAccAuth_azurePowerShell tests authentication using Azure PowerShell
func TestAccAuth_azurePowerShell(t *testing.T) {
	if ok := os.Getenv("ARM_USE_POWERSHELL"); ok == "" {
		t.Skip("Skipping as `ARM_USE_POWERSHELL` is not specified")
	}

	data := acceptance.BuildTestData(t, "data.msgraph_resource", "test")
	r := MSGraphTestDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.basic(data),
			Check:  resource.ComposeTestCheckFunc(),
		},
	})
}
