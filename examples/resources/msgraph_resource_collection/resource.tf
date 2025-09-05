terraform {
  required_providers {
    msgraph = {
      source = "Microsoft/msgraph"
    }
  }
}

provider "msgraph" {}

resource "msgraph_resource" "application_a" {
  url = "applications"
  body = {
    displayName = "Collection Example App A"
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
    displayName = "Collection Example App B"
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
    displayName     = "Collection Example Group"
    mailEnabled     = false
    mailNickname    = "collection-example-group"
    securityEnabled = true
  }
}

resource "msgraph_resource_collection" "group_members" {
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  // This API has a known issue where service principals are not listed as group members in v1.0. As a workaround, 
  // use this API on the beta endpoint or use the /groups/{id}?$expand=members API. For more information, 
  // see the related known issue: https://developer.microsoft.com/en-us/graph/known-issues/?search=25984
  api_version   = "beta"
  reference_ids = [msgraph_resource.sp_a.id, msgraph_resource.sp_b.id]
}
