// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/azure/terraform-provider-msgraph/internal/acceptance"
	"github.com/azure/terraform-provider-msgraph/internal/acceptance/check"
	"github.com/azure/terraform-provider-msgraph/internal/clients"
	"github.com/azure/terraform-provider-msgraph/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func defaultIgnores() []string {
	return []string{"body", "output"}
}

type MSGraphTestResource struct {
}

func TestAcc_ResourceBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")

	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
		data.ImportStep(defaultIgnores()...),
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
		data.ImportStep(defaultIgnores()...),
		{
			Config: r.basicUpdate(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
		},
		data.ImportStep(defaultIgnores()...),
	})
}

func (M MSGraphTestResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	apiVersion := state.Attributes["api_version"]
	_, err := client.MSGraphClient.Read(ctx, state.ID, apiVersion)
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
