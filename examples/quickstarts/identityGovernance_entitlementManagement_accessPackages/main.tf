terraform {
  required_providers {
    msgraph = {
      source = "microsoft/msgraph"
    }
  }
}

provider "msgraph" {
}

resource "msgraph_resource" "catalog" {
  url = "identityGovernance/entitlementManagement/catalogs"
  body = {
    displayName = "example-catalog"
    description = "Example catalog"
  }
}

resource "msgraph_resource" "accessPackage" {
  url         = "identityGovernance/entitlementManagement/accessPackages"
  api_version = "beta"
  body = {
    catalogId   = msgraph_resource.catalog.id
    displayName = "access-package"
    description = "Access Package"
  }
}