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
