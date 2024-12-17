// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package services_test

import (
	"fmt"
	"testing"

	"github.com/azure/terraform-provider-msgraph/internal/acceptance"
	"github.com/azure/terraform-provider-msgraph/internal/acceptance/check"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
			),
		},
	})
}

func (r MSGraphTestDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "msgraph_resource" "test" {
  url = msgraph_resource.test.id
}
`, MSGraphTestResource{}.basic(data))
}
