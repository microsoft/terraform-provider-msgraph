terraform {
  required_providers {
    msgraph = {
      source = "microsoft/msgraph"
    }
  }
}

provider "msgraph" {
}

resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "My Group"
    mailEnabled     = false
    mailNickname    = "mygroup"
    securityEnabled = true
  }
}

resource "msgraph_resource" "application" {
  url = "applications"
  body = {
    displayName = "My Application"
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
}

resource "msgraph_resource" "member" {
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/directoryObjects/${msgraph_resource.servicePrincipal_application.id}"
  }
}
