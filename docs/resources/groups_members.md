---
subcategory: "Reference"
page_title: "groups/members - group member"
description: |-
  Manages a group member.
---

# groups/members - group member

This article demonstrates how to use `msgraph` provider to manage the group member resource in MSGraph.

## Example Usage

### individual_references

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

# Create a group
resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Finance Team"
    mailEnabled     = false
    mailNickname    = "finance-team"
    securityEnabled = true
  }
}

# Create users to be added as members
resource "msgraph_resource" "user1" {
  url = "users"
  body = {
    userPrincipalName = "david@example.com"
    displayName       = "David Wilson"
    mailNickname      = "david"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

resource "msgraph_resource" "user2" {
  url = "users"
  body = {
    userPrincipalName = "emma@example.com"
    displayName       = "Emma Brown"
    mailNickname      = "emma"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

# Variable to conditionally add a third member
variable "include_third_member" {
  description = "Whether to include the third member in the group"
  type        = bool
  default     = false
}

resource "msgraph_resource" "user3" {
  count = var.include_third_member ? 1 : 0
  url   = "users"
  body = {
    userPrincipalName = "frank@example.com"
    displayName       = "Frank Garcia"
    mailNickname      = "frank"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

# Add individual members using separate resources
# This approach provides fine-grained control over each relationship
resource "msgraph_resource" "member1" {
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/users/${msgraph_resource.user1.id}"
  }
}

resource "msgraph_resource" "member2" {
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/users/${msgraph_resource.user2.id}"
  }
}

# Conditional member - only added if variable is true
resource "msgraph_resource" "member3" {
  count = var.include_third_member ? 1 : 0
  url   = "groups/${msgraph_resource.group.id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/users/${msgraph_resource.user3[0].id}"
  }
}
```

### inline_creation

```hcl
terraform {
  required_providers {
    msgraph = {
      source = "microsoft/msgraph"
    }
    azurerm = {
      source = "hashicorp/azurerm"
    }
  }
}

provider "msgraph" {
}

provider "azurerm" {
  features {}
}

# Get current client configuration for setting initial owner
data "azurerm_client_config" "current" {}

# Create users first (they need to exist before being referenced)
resource "msgraph_resource" "user1" {
  url = "users"
  body = {
    userPrincipalName = "sarah@example.com"
    displayName       = "Sarah Johnson"
    mailNickname      = "sarah"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

resource "msgraph_resource" "user2" {
  url = "users"
  body = {
    userPrincipalName = "mike@example.com"
    displayName       = "Mike Anderson"
    mailNickname      = "mike"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

resource "msgraph_resource" "manager_user" {
  url = "users"
  body = {
    userPrincipalName = "manager@example.com"
    displayName       = "Team Manager"
    mailNickname      = "manager"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

# Create group with initial members and owners using OData bind syntax
resource "msgraph_resource" "group_with_relationships" {
  url = "groups"
  body = {
    displayName     = "DevOps Team"
    mailEnabled     = false
    mailNickname    = "devops-team"
    securityEnabled = true
    
    # Set initial owners using OData bind syntax
    "owners@odata.bind" = [
      "https://graph.microsoft.com/v1.0/users/${data.azurerm_client_config.current.object_id}",
      "https://graph.microsoft.com/v1.0/users/${msgraph_resource.manager_user.id}"
    ]
    
    # Set initial members using OData bind syntax
    "members@odata.bind" = [
      "https://graph.microsoft.com/v1.0/users/${msgraph_resource.user1.id}",
      "https://graph.microsoft.com/v1.0/users/${msgraph_resource.user2.id}",
      "https://graph.microsoft.com/v1.0/users/${msgraph_resource.manager_user.id}"
    ]
  }
}
```

### resource_collection

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

# Create a group
resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Development Team"
    mailEnabled     = false
    mailNickname    = "dev-team"
    securityEnabled = true
  }
}

# Create users to be added as members
resource "msgraph_resource" "user1" {
  url = "users"
  body = {
    userPrincipalName = "alice@example.com"
    displayName       = "Alice Smith"
    mailNickname      = "alice"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

resource "msgraph_resource" "user2" {
  url = "users"
  body = {
    userPrincipalName = "bob@example.com"
    displayName       = "Bob Johnson"
    mailNickname      = "bob"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

resource "msgraph_resource" "user3" {
  url = "users"
  body = {
    userPrincipalName = "charlie@example.com"
    displayName       = "Charlie Davis"
    mailNickname      = "charlie"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

# Manage all group members as a collection (RECOMMENDED APPROACH)
# This provides declarative management where you specify the complete desired state
resource "msgraph_resource_collection" "group_members" {
  url           = "groups/${msgraph_resource.group.id}/members/$ref"
  reference_ids = [
    msgraph_resource.user1.id,
    msgraph_resource.user2.id,
    msgraph_resource.user3.id
  ]
}
```



## Arguments Reference

The following arguments are supported:

* `url` - (Required) The URL which is used to manage the resource. This should be set to `groups/{group-id}/members/$ref`.

* `body` - (Required) Specifies the configuration of the resource. More information about the arguments in `body` can be found in the [Microsoft documentation](https://learn.microsoft.com/en-us/azure/templates/groups/members?pivots=deployment-language-terraform).

* `api_version` - (Optional) The API version used to manage the resource. The default value is `v1.0`. The allowed values are `v1.0` and `beta`.

For other arguments, please refer to the [msgraph_resource](https://registry.terraform.io/providers/Microsoft/msgraph/latest/docs/resources/resource) documentation.

### Read-Only

- `id` (String) The ID of the resource. Normally, it is in the format of UUID.

## Import

 ```shell
 # MSGraph resource can be imported using the resource id, e.g.
 terraform import msgraph_resource.example /groups/{group-id}/members/{members-id}/$ref
 
 # It also supports specifying API version by using the resource id with api-version as a query parameter, e.g.
 terraform import msgraph_resource.example /groups/{group-id}/members/{members-id}/$ref?api-version=v1.0
 ```
