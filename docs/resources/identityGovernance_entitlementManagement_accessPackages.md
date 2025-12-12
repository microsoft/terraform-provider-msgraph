---
subcategory: "Reference"
page_title: "identityGovernance/entitlementManagement/accessPackages - access package"
description: |-
  Manages a access package.
---

# identityGovernance/entitlementManagement/accessPackages - access package

This article demonstrates how to use `msgraph` provider to manage the access package resource in MSGraph.

## Example Usage

### default

```hcl
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
```



## Arguments Reference

The following arguments are supported:

* `url` - (Required) The URL which is used to manage the resource. This should be set to `identityGovernance/entitlementManagement/accessPackages`.

* `body` - (Required) Specifies the configuration of the resource. More information about the arguments in `body` can be found in the [Microsoft documentation](https://learn.microsoft.com/en-us/graph/templates/terraform/reference/v1.0/identityGovernance/entitlementManagement/accessPackages).

* `api_version` - (Optional) The API version used to manage the resource. The default value is `v1.0`. The allowed values are `v1.0` and `beta`.

For other arguments, please refer to the [msgraph_resource](https://registry.terraform.io/providers/Microsoft/msgraph/latest/docs/resources/resource) documentation.

### Read-Only

- `id` (String) The ID of the resource. Normally, it is in the format of UUID.

## Import

 ```shell
 # MSGraph resource can be imported using the resource id, e.g.
 terraform import msgraph_resource.example /identityGovernance/entitlementManagement/accessPackages/{accessPackages-id}
 
 # It also supports specifying API version by using the resource id with api-version as a query parameter, e.g.
 terraform import msgraph_resource.example /identityGovernance/entitlementManagement/accessPackages/{accessPackages-id}?api-version=v1.0
 ```
