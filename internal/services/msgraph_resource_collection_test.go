package services_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance/check"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
	"github.com/microsoft/terraform-provider-msgraph/internal/utils"
)

type MSGraphTestResourceCollection struct{}

func TestAcc_ResourceCollectionBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_collection", "test")
	r := MSGraphTestResourceCollection{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				// id is the base collection URL (without /$ref)
				check.That(data.ResourceName).Key("id").MatchesRegex(regexp.MustCompile(`groups/.+/members$`)),
				resource.TestCheckResourceAttr(data.ResourceName, "reference_ids.#", "1"),
			),
		},
	})
}

func TestAcc_ResourceCollectionUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_collection", "test")
	r := MSGraphTestResourceCollection{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			// start with one member
			Config: r.updateOneMember(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				resource.TestCheckResourceAttr(data.ResourceName, "reference_ids.#", "1"),
			),
		},
		{
			// add second member
			Config: r.updateTwoMembers(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				resource.TestCheckResourceAttr(data.ResourceName, "reference_ids.#", "2"),
			),
		},
		{
			// remove second member again
			Config: r.updateOneMember(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				resource.TestCheckResourceAttr(data.ResourceName, "reference_ids.#", "1"),
			),
		},
	})
}

func TestAcc_ResourceCollectionRetry(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_collection", "test")
	r := MSGraphTestResourceCollection{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withRetry(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				resource.TestCheckResourceAttr(data.ResourceName, "reference_ids.#", "1"),
			),
		},
	})
}

func TestAcc_ResourceCollectionReadQueryParameters(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_collection", "test")
	r := MSGraphTestResourceCollection{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basicWithReadQueryParameters(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				// still expects one member, proving list call executed successfully with query params
				resource.TestCheckResourceAttr(data.ResourceName, "reference_ids.#", "1"),
			),
		},
	})
}

func TestAcc_ResourceCollectionTimeouts_Create(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_collection", "test")
	r := MSGraphTestResourceCollection{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:      r.withCreateTimeout(),
			ExpectError: regexp.MustCompile(`context deadline exceeded`),
		},
	})
}

func TestAcc_ResourceCollectionTimeouts_Update(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_collection", "test")
	r := MSGraphTestResourceCollection{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.updateOneMember(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				resource.TestCheckResourceAttr(data.ResourceName, "reference_ids.#", "1"),
			),
		},
		{
			Config:      r.withUpdateTimeoutTwoMembers(),
			ExpectError: regexp.MustCompile(`context deadline exceeded`),
		},
	})
}

func TestAcc_ResourceCollectionTimeouts_Read(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource_collection", "test")
	r := MSGraphTestResourceCollection{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:      r.withReadTimeoutOneMember(),
			ExpectError: regexp.MustCompile(`context deadline exceeded`),
		},
	})
}

// Exists checks that the underlying collection endpoint exists by listing it.
func (r MSGraphTestResourceCollection) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	apiVersion := state.Attributes["api_version"]
	id := state.Attributes["id"] // base collection URL

	_, err := client.MSGraphClient.List(ctx, id, apiVersion, clients.DefaultRequestOptions())
	if err == nil {
		b := true
		return &b, nil
	}
	if utils.ResponseErrorWasNotFound(err) {
		b := false
		return &b, nil
	}
	return nil, fmt.Errorf("checking for presence of existing collection %s(api_version=%s): %w", id, apiVersion, err)
}

// configuration helpers
func (r MSGraphTestResourceCollection) basic() string {
	return `
resource "msgraph_resource" "application_a" {
  url = "applications"
  body = {
    displayName = "Collection App a"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_a" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_a.output.appId
  }
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Collection Group"
    mailEnabled     = false
    mailNickname    = "collection-group"
    securityEnabled = true
  }
}

resource "msgraph_resource_collection" "test" {
  url           = "groups/${msgraph_resource.group.id}/members/$ref"
  api_version   = "beta"
  reference_ids = [msgraph_resource.sp_a.id]
}
`
}

func (r MSGraphTestResourceCollection) basicWithReadQueryParameters() string {
	return `
resource "msgraph_resource" "application_a" {
  url = "applications"
  body = {
    displayName = "Collection App a"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_a" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_a.output.appId
  }
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Collection Group"
    mailEnabled     = false
    mailNickname    = "collection-group"
    securityEnabled = true
  }
}

resource "msgraph_resource_collection" "test" {
  url           = "groups/${msgraph_resource.group.id}/members/$ref"
  api_version   = "beta"
  reference_ids = [msgraph_resource.sp_a.id]
  read_query_parameters = {
    "$select" = ["id"]
  }
}
`
}

func (r MSGraphTestResourceCollection) updateOneMember() string {
	return `
resource "msgraph_resource" "application_a" {
  url = "applications"
  body = {
    displayName = "Collection App a"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_a" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_a.output.appId
  }
}

resource "msgraph_resource" "application_b" {
  url = "applications"
  body = {
    displayName = "Collection App b"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_b" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_b.output.appId
  }
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Collection Group"
    mailEnabled     = false
    mailNickname    = "collection-group"
    securityEnabled = true
  }
}

resource "msgraph_resource_collection" "test" {
  url           = "groups/${msgraph_resource.group.id}/members/$ref"
  api_version   = "beta"
  reference_ids = [msgraph_resource.sp_a.id]
}
`
}

func (r MSGraphTestResourceCollection) updateTwoMembers() string {
	return `
resource "msgraph_resource" "application_a" {
  url = "applications"
  body = {
    displayName = "Collection App a"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_a" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_a.output.appId
  }
}

resource "msgraph_resource" "application_b" {
  url = "applications"
  body = {
    displayName = "Collection App b"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_b" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_b.output.appId
  }
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Collection Group"
    mailEnabled     = false
    mailNickname    = "collection-group"
    securityEnabled = true
  }
}

resource "msgraph_resource_collection" "test" {
  url         = "groups/${msgraph_resource.group.id}/members/$ref"
  api_version = "beta"
  reference_ids = [
    msgraph_resource.sp_a.id,
    msgraph_resource.sp_b.id,
  ]
}
`
}

func (r MSGraphTestResourceCollection) withRetry() string {
	return `
resource "msgraph_resource" "application_a" {
  url = "applications"
  body = {
    displayName = "Collection App a"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_a" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_a.output.appId
  }
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Collection Group"
    mailEnabled     = false
    mailNickname    = "collection-group"
    securityEnabled = true
  }
}

resource "msgraph_resource_collection" "test" {
  url           = "groups/${msgraph_resource.group.id}/members/$ref"
  api_version   = "beta"
  reference_ids = [msgraph_resource.sp_a.id]
  retry = {
    error_message_regex = [
      ".*throttl.*",
      "temporary error",
    ]
  }
}
`
}

func (r MSGraphTestResourceCollection) withCreateTimeout() string {
	return `
resource "msgraph_resource" "application_a" {
  url = "applications"
  body = {
    displayName = "Collection App a"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_a" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_a.output.appId
  }
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Collection Group"
    mailEnabled     = false
    mailNickname    = "collection-group"
    securityEnabled = true
  }
}

resource "msgraph_resource_collection" "test" {
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  timeouts {
    create = "1ns"
  }
  api_version   = "beta"
  reference_ids = [msgraph_resource.sp_a.id]
}
`
}

func (r MSGraphTestResourceCollection) withUpdateTimeoutTwoMembers() string {
	return `
resource "msgraph_resource" "application_a" {
  url = "applications"
  body = {
    displayName = "Collection App a"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_a" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_a.output.appId
  }
}

resource "msgraph_resource" "application_b" {
  url = "applications"
  body = {
    displayName = "Collection App b"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_b" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_b.output.appId
  }
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Collection Group"
    mailEnabled     = false
    mailNickname    = "collection-group"
    securityEnabled = true
  }
}

resource "msgraph_resource_collection" "test" {
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  timeouts {
    update = "1ns"
  }
  api_version = "beta"
  reference_ids = [
    msgraph_resource.sp_a.id,
    msgraph_resource.sp_b.id,
  ]
}
`
}

func (r MSGraphTestResourceCollection) withReadTimeoutOneMember() string {
	return `
resource "msgraph_resource" "application_a" {
  url = "applications"
  body = {
    displayName = "Collection App a"
  }
  response_export_values = {
    appId = "appId"
  }
}

resource "msgraph_resource" "sp_a" {
  url = "servicePrincipals"
  body = {
    appId = msgraph_resource.application_a.output.appId
  }
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Collection Group"
    mailEnabled     = false
    mailNickname    = "collection-group"
    securityEnabled = true
  }
}

resource "msgraph_resource_collection" "test" {
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  timeouts {
    read = "1ns"
  }
  api_version   = "beta"
  reference_ids = [msgraph_resource.sp_a.id]
}
`
}
