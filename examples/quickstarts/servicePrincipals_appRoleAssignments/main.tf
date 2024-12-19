terraform {
  required_providers {
    msgraph = {
      source = "microsoft/msgraph"
    }
  }
}

provider "msgraph" {
}

locals {
  MicrosoftGraphAppId = "00000003-0000-0000-c000-000000000000"

  # AppRoleAssignment
  userReadAllAppRoleId = one([for role in data.msgraph_resource.servicePrincipal_msgraph.output.all.value[0].appRoles : role.id if role.value == "User.Read.All"])
  userReadWriteRoleId  = one([for role in data.msgraph_resource.servicePrincipal_msgraph.output.all.value[0].oauth2PermissionScopes : role.id if role.value == "User.ReadWrite"])

  # ServicePrincipal
  MSGraphServicePrincipalId         = data.msgraph_resource.servicePrincipal_msgraph.output.all.value[0].id
  TestApplicationServicePrincipalId = msgraph_resource.servicePrincipal_application.output.all.id
}

data "msgraph_resource" "servicePrincipal_msgraph" {
  url = "servicePrincipals"
  query_parameters = {
    "$filter" = ["appId eq '${local.MicrosoftGraphAppId}'"]
  }
  response_export_values = {
    all = "@"
  }
}

resource "msgraph_resource" "application" {
  url = "applications"
  body = {
    displayName = "My Application"
    requiredResourceAccess = [
      {
        resourceAppId = local.MicrosoftGraphAppId
        resourceAccess = [
          {
            id   = local.userReadAllAppRoleId
            type = "Scope"
          },
          {
            id   = local.userReadWriteRoleId
            type = "Scope"
          }
        ]
      }
    ]
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
  response_export_values = {
    all = "@"
  }
}

resource "msgraph_resource" "appRoleAssignment" {
  url = "servicePrincipals/${local.MSGraphServicePrincipalId}/appRoleAssignments"
  body = {
    appRoleId   = local.userReadAllAppRoleId
    principalId = local.TestApplicationServicePrincipalId
    resourceId  = local.MSGraphServicePrincipalId
  }
}
