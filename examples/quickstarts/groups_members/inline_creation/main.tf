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