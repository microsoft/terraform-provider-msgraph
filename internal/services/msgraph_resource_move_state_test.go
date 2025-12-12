package services_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance"
	"github.com/microsoft/terraform-provider-msgraph/internal/acceptance/check"
)

func externalProvidersAzureAD() map[string]resource.ExternalProvider {
	return map[string]resource.ExternalProvider{
		"azuread": {
			Source:            "hashicorp/azuread",
			VersionConstraint: "3.0.2",
		},
	}
}

func TestAcc_ResourceMoveState_Application(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateApplicationSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("azuread_application.test", "display_name", fmt.Sprintf("acctest%s", data.RandomString)),
			),
		},
		{
			Config:            r.moveStateApplicationMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("body.displayName").HasValue(fmt.Sprintf("acctest%s", data.RandomString)),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionUpdate),
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_Group(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateGroupSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("azuread_group.test", "display_name", fmt.Sprintf("acctest%s", data.RandomString)),
			),
		},
		{
			Config:            r.moveStateGroupMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("body.displayName").HasValue(fmt.Sprintf("acctest%s", data.RandomString)),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionUpdate),
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_ServicePrincipal(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateServicePrincipalSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("azuread_application.test", "display_name", fmt.Sprintf("acctest%s", data.RandomString)),
			),
		},
		{
			Config:            r.moveStateServicePrincipalMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("body.appId").IsUUID(),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionUpdate),
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_User(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateUserSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("azuread_user.test", "display_name", fmt.Sprintf("acctest%s", data.RandomString)),
			),
		},
		{
			Config:            r.moveStateUserMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("body.displayName").HasValue(fmt.Sprintf("acctest%s", data.RandomString)),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionNoop),
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_AdministrativeUnit(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateAdministrativeUnitSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("azuread_administrative_unit.test", "display_name", fmt.Sprintf("acctest%s", data.RandomString)),
			),
		},
		{
			Config:            r.moveStateAdministrativeUnitMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("body.displayName").HasValue(fmt.Sprintf("acctest%s", data.RandomString)),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionUpdate),
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_ApplicationRegistration(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateApplicationRegistrationSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("azuread_application_registration.test", "display_name", fmt.Sprintf("acctest%s", data.RandomString)),
			),
		},
		{
			Config:            r.moveStateApplicationRegistrationMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
				check.That(data.ResourceName).Key("body.displayName").HasValue(fmt.Sprintf("acctest%s", data.RandomString)),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionUpdate),
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_GroupMember(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateGroupMemberSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("azuread_group_member.test", "id"),
			),
		},
		{
			Config:            r.moveStateGroupMemberMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionNoop),
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_AdministrativeUnitMember(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateAdministrativeUnitMemberSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("azuread_administrative_unit_member.test", "id"),
			),
		},
		{
			Config:            r.moveStateAdministrativeUnitMemberMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionNoop),
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_ApplicationOwner(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateApplicationOwnerSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("azuread_application_owner.test", "id"),
			),
		},
		{
			Config:            r.moveStateApplicationOwnerMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionNoop),
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_DirectoryRoleMember(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.RunAcceptanceTest(t, resource.TestCase{
		PreCheck: func() { acceptance.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:            r.moveStateDirectoryRoleMemberSetup(data),
				ExternalProviders: externalProvidersAzureAD(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("azuread_directory_role_member.test", "id"),
				),
			},
			{
				Config:            r.moveStateDirectoryRoleMemberMoved(data),
				ExternalProviders: externalProvidersAzureAD(),
				Check: resource.ComposeTestCheckFunc(
					check.That(data.ResourceName).Exists(r),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAcc_ResourceMoveState_ServicePrincipalClaimsMappingPolicyAssignment(t *testing.T) {
	data := acceptance.BuildTestData(t, "msgraph_resource", "test")
	r := MSGraphTestResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:            r.moveStateServicePrincipalClaimsMappingPolicyAssignmentSetup(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("azuread_service_principal_claims_mapping_policy_assignment.test", "id"),
			),
		},
		{
			Config:            r.moveStateServicePrincipalClaimsMappingPolicyAssignmentMoved(data),
			ExternalProviders: externalProvidersAzureAD(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Exists(r),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(data.ResourceName, plancheck.ResourceActionNoop),
				},
			},
		},
	})
}

func (r MSGraphTestResource) moveStateGroupMemberSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

resource "azuread_group" "test" {
  display_name     = "acctest%[1]s"
  security_enabled = true
  mail_enabled     = false
  mail_nickname    = "acctest%[1]s"
}

resource "azuread_user" "test" {
  user_principal_name = "acctest%[1]s@${local.domain}"
  display_name        = "acctest%[1]s"
  mail_nickname       = "acctest%[1]s"
  password            = "SecretP@sswd%[1]s"
}

resource "azuread_group_member" "test" {
  group_object_id  = azuread_group.test.object_id
  member_object_id = azuread_user.test.object_id
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateGroupMemberMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

resource "azuread_group" "test" {
  display_name     = "acctest%[1]s"
  security_enabled = true
  mail_enabled     = false
  mail_nickname    = "acctest%[1]s"
}

resource "azuread_user" "test" {
  user_principal_name = "acctest%[1]s@${local.domain}"
  display_name        = "acctest%[1]s"
  mail_nickname       = "acctest%[1]s"
  password            = "SecretP@sswd%[1]s"
}

# resource "azuread_group_member" "test" {
#  group_object_id  = azuread_group.test.object_id
#  member_object_id = azuread_user.test.object_id
# }

moved {
  from = azuread_group_member.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/groups/${azuread_group.test.object_id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/directoryObjects/${azuread_user.test.object_id}"
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateApplicationSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

resource "azuread_application" "test" {
  display_name = "acctest%[1]s"
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateApplicationMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

# resource "azuread_application" "test" {
#   display_name = "acctest%[1]s"
# }

moved {
  from = azuread_application.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/applications"
  body = {
    displayName = "acctest%[1]s"
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateGroupSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

resource "azuread_group" "test" {
  display_name     = "acctest%[1]s"
  mail_enabled     = false
  mail_nickname    = "acctest%[1]s"
  security_enabled = true
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateGroupMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

# resource "azuread_group" "test" {
#   display_name     = "acctest%[1]s"
#   mail_enabled     = false
#   mail_nickname    = "acctest%[1]s"
#   security_enabled = true
# }

moved {
  from = azuread_group.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/groups"
  body = {
    displayName     = "acctest%[1]s"
    mailEnabled     = false
    mailNickname    = "acctest%[1]s"
    securityEnabled = true
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateServicePrincipalSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

resource "azuread_application" "test" {
  display_name = "acctest%[1]s"
}

resource "azuread_service_principal" "test" {
  client_id = azuread_application.test.client_id
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateServicePrincipalMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

resource "azuread_application" "test" {
  display_name = "acctest%[1]s"
}

# resource "azuread_service_principal" "test" {
#   client_id = azuread_application.test.client_id
# }

moved {
  from = azuread_service_principal.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/servicePrincipals"
  body = {
    appId = azuread_application.test.client_id
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateUserSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

resource "azuread_user" "test" {
  user_principal_name = "acctest%[1]s@${local.domain}"
  display_name        = "acctest%[1]s"
  mail_nickname       = "acctest%[1]s"
  password            = "SecretP@sswd%[1]s"
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateUserMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

//resource "azuread_user" "test" {
//  user_principal_name = "acctest%[1]s@${local.domain}"
//  display_name        = "acctest%[1]s"
//  mail_nickname       = "acctest%[1]s"
//  password            = "SecretP@sswd%[1]s"
//}

moved {
  from = azuread_user.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/users"
  body = {
    userPrincipalName = "acctest%[1]s@${local.domain}"
    displayName       = "acctest%[1]s"
    mailNickname      = "acctest%[1]s"
    passwordProfile = {
      password = "SecretP@sswd%[1]s"
    }
    accountEnabled = true
  }

  lifecycle {
    ignore_changes = [body]
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateAdministrativeUnitSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

resource "azuread_administrative_unit" "test" {
  display_name = "acctest%[1]s"
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateAdministrativeUnitMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

# resource "azuread_administrative_unit" "test" {
#   display_name = "acctest%[1]s"
# }

moved {
  from = azuread_administrative_unit.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/directory/administrativeUnits"
  body = {
    displayName = "acctest%[1]s"
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateApplicationRegistrationSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

resource "azuread_application_registration" "test" {
  display_name = "acctest%[1]s"
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateApplicationRegistrationMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

# resource "azuread_application_registration" "test" {
#   display_name = "acctest%[1]s"
# }

moved {
  from = azuread_application_registration.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/applications"
  body = {
    displayName = "acctest%[1]s"
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateAdministrativeUnitMemberSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

resource "azuread_administrative_unit" "test" {
  display_name = "acctest%[1]s"
}

resource "azuread_user" "test" {
  user_principal_name = "acctest%[1]s@${local.domain}"
  display_name        = "acctest%[1]s"
  mail_nickname       = "acctest%[1]s"
  password            = "SecretP@sswd%[1]s"
}

resource "azuread_administrative_unit_member" "test" {
  administrative_unit_object_id = azuread_administrative_unit.test.object_id
  member_object_id              = azuread_user.test.object_id
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateAdministrativeUnitMemberMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

resource "azuread_administrative_unit" "test" {
  display_name = "acctest%[1]s"
}

resource "azuread_user" "test" {
  user_principal_name = "acctest%[1]s@${local.domain}"
  display_name        = "acctest%[1]s"
  mail_nickname       = "acctest%[1]s"
  password            = "SecretP@sswd%[1]s"
}

moved {
  from = azuread_administrative_unit_member.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/directory/administrativeUnits/${azuread_administrative_unit.test.object_id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/directoryObjects/${azuread_user.test.object_id}"
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateApplicationOwnerSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

resource "azuread_application" "test" {
  display_name = "acctest%[1]s"

  lifecycle {
    ignore_changes = [owners]
  }
}

resource "azuread_user" "test" {
  user_principal_name = "acctest%[1]s@${local.domain}"
  display_name        = "acctest%[1]s"
  mail_nickname       = "acctest%[1]s"
  password            = "SecretP@sswd%[1]s"
}

resource "azuread_application_owner" "test" {
  application_id  = azuread_application.test.id
  owner_object_id = azuread_user.test.object_id
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateApplicationOwnerMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

resource "azuread_application" "test" {
  display_name = "acctest%[1]s"
  lifecycle {
    ignore_changes = [owners]
  }
}

resource "azuread_user" "test" {
  user_principal_name = "acctest%[1]s@${local.domain}"
  display_name        = "acctest%[1]s"
  mail_nickname       = "acctest%[1]s"
  password            = "SecretP@sswd%[1]s"
}

moved {
  from = azuread_application_owner.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/applications/${azuread_application.test.object_id}/owners/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/directoryObjects/${azuread_user.test.object_id}"
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateDirectoryRoleMemberSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

resource "azuread_directory_role" "test" {
  display_name = "Application Administrator"
}

resource "azuread_user" "test" {
  user_principal_name = "acctest%[1]s@${local.domain}"
  display_name        = "acctest%[1]s"
  mail_nickname       = "acctest%[1]s"
  password            = "SecretP@sswd%[1]s"
}

resource "azuread_directory_role_member" "test" {
  role_object_id   = azuread_directory_role.test.object_id
  member_object_id = azuread_user.test.object_id
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateDirectoryRoleMemberMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "msgraph_resource" "domains" {
  url = "domains"
  response_export_values = {
    all = "@"
  }
}

locals {
  domain = one([for domain in data.msgraph_resource.domains.output.all.value : domain.id if domain.isInitial])
}

resource "azuread_directory_role" "test" {
  display_name = "Application Administrator"
}

resource "azuread_user" "test" {
  user_principal_name = "acctest%[1]s@${local.domain}"
  display_name        = "acctest%[1]s"
  mail_nickname       = "acctest%[1]s"
  password            = "SecretP@sswd%[1]s"
}

moved {
  from = azuread_directory_role_member.test
  to   = msgraph_resource.test
}

resource "msgraph_resource" "test" {
  url = "/directoryRoles/${azuread_directory_role.test.object_id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/directoryObjects/${azuread_user.test.object_id}"
  }
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateServicePrincipalClaimsMappingPolicyAssignmentSetup(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

resource "azuread_application" "test" {
  display_name = "acctest%[1]s"
}

resource "azuread_service_principal" "test" {
  client_id = azuread_application.test.client_id
}

resource "azuread_claims_mapping_policy" "test" {
  definition = [
    jsonencode(
      {
        ClaimsMappingPolicy = {
          Version              = 1
          IncludeBasicClaimSet = "true"
          ClaimsSchema = [
            {
              Source       = "user"
              ID           = "employeeid"
              JwtClaimType = "employeeid"
            },
          ]
        }
      }
    ),
  ]
  display_name = "acctest%[1]s"
}

resource "azuread_service_principal_claims_mapping_policy_assignment" "test" {
  claims_mapping_policy_id = azuread_claims_mapping_policy.test.id
  service_principal_id     = azuread_service_principal.test.id
}
`, data.RandomString)
}

func (r MSGraphTestResource) moveStateServicePrincipalClaimsMappingPolicyAssignmentMoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

resource "azuread_application" "test" {
  display_name = "acctest%[1]s"
}

resource "azuread_service_principal" "test" {
  client_id = azuread_application.test.client_id
}

resource "azuread_claims_mapping_policy" "test" {
  definition = [
    jsonencode(
      {
        ClaimsMappingPolicy = {
          Version              = 1
          IncludeBasicClaimSet = "true"
          ClaimsSchema = [
            {
              Source       = "user"
              ID           = "employeeid"
              JwtClaimType = "employeeid"
            },
          ]
        }
      }
    ),
  ]
  display_name = "acctest%[1]s"
}

moved {
  from = azuread_service_principal_claims_mapping_policy_assignment.test
  to   = msgraph_resource.test
}

locals {
  policyObjectId = trimprefix(azuread_claims_mapping_policy.test.id, "/policies/claimsMappingPolicies/")
}

resource "msgraph_resource" "test" {
  url = "${azuread_service_principal.test.id}/claimsMappingPolicies/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/directoryObjects/${local.policyObjectId}"
  }
}
`, data.RandomString)
}
