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