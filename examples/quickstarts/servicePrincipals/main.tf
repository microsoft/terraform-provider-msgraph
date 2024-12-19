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