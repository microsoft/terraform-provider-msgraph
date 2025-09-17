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