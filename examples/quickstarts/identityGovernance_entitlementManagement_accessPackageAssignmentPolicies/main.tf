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

resource "msgraph_resource" "accessPackageAssignmentPolicy" {
  url           = "identityGovernance/entitlementManagement/accessPackageAssignmentPolicies"
  api_version   = "beta"
  update_method = "PUT"
  body = {
    accessPackageId = msgraph_resource.accessPackage.id
    displayName     = "My Assignment Policy"
    description     = "My assignment policy description"
    expiration = {
      type     = "afterDuration"
      duration = "P90D"
    }
    requestorSettings = {
      scopeType = "AllExistingDirectoryMemberUsers"
    }
    requestApprovalSettings = {
      isApprovalRequired = true
      approvalStages = [
        {
          approvalStageTimeOutInDays = 14
          primaryApprovers = [
            {
              "@odata.type" = "#microsoft.graph.groupMembers"
              groupId       = msgraph_resource.group.id
              description   = "group-name"
            }
          ]
        }
      ]
    }
    reviewSettings = {
      isEnabled          = true
      expirationBehavior = "keepAccess"
      isSelfReview       = true
      schedule = {
        startDateTime = "2025-12-12T00:00:00Z"
        recurrence = {
          pattern = {
            type     = "weekly"
            interval = 1
          }
          range = {
            type      = "noEnd"
            startDate = "2025-12-12"
          }
        }
      }
    }
    questions = [
      {
        "@odata.type" = "#microsoft.graph.accessPackageTextInputQuestion"
        text = {
          defaultText = "hello, how are you?"
        }
        isRequired = false
      }
    ]
  }
}