terraform {
  required_providers {
    msgraph = {
      source = "Microsoft/msgraph"
    }
  }
}

provider "msgraph" {}

# Example 1: Get user's member groups
data "msgraph_resource_action" "user_member_groups" {
  resource_url = "users/john@example.com"
  action       = "getMemberGroups"
  method       = "POST"

  body = {
    securityEnabledOnly = false
  }

  response_export_values = {
    groups = "value"
  }
}

# Example 2: Check group membership
data "msgraph_resource_action" "check_membership" {
  resource_url = "users/john@example.com"
  action       = "checkMemberGroups"
  method       = "POST"

  body = {
    groupIds = [
      "{group-id-1}",
      "{group-id-2}"
    ]
  }

  response_export_values = {
    matched_groups = "value"
  }
}

# Example 3: Get group members with query parameters
data "msgraph_resource_action" "group_members" {
  resource_url = "groups/{group-id}"
  action       = "members"
  method       = "GET"

  query_parameters = {
    "$select" = ["id", "displayName", "userPrincipalName"]
    "$top"    = ["100"]
  }

  headers = {
    "ConsistencyLevel" = "eventual"
  }

  response_export_values = {
    members    = "value"
    member_ids = "value[].id"
    next_link  = "@odata.nextLink"
  }
}

# Example 4: Get application service principal
data "msgraph_resource_action" "app_service_principal" {
  resource_url = "applications/{application-id}"
  action       = "servicePrincipals"
  method       = "GET"

  query_parameters = {
    "$select" = ["id", "appId", "displayName"]
  }

  response_export_values = {
    service_principals = "value"
    sp_id              = "value[0].id"
  }
}

# Output the results
output "user_groups" {
  value = data.msgraph_resource_action.user_member_groups.output.groups
}

output "matched_groups" {
  value = data.msgraph_resource_action.check_membership.output.matched_groups
}

output "group_members" {
  value = data.msgraph_resource_action.group_members.output.members
}

output "service_principal_id" {
  value = data.msgraph_resource_action.app_service_principal.output.sp_id
}
