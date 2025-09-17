terraform {
  required_providers {
    msgraph = {
      source = "Microsoft/msgraph"
    }
    azuread = {
      source = "hashicorp/azuread"
    }
  }
}

provider "msgraph" {}
provider "azuread" {}

# This example creates an application using azuread provider first (so we have something to update),
# then uses msgraph_update_resource to PATCH its displayName.

resource "azuread_application" "application" {
  display_name = "Demo App"

  lifecycle {
    ignore_changes = [display_name]
  }
}

resource "msgraph_update_resource" "application_update" {
  # Point directly at the item URL you want to PATCH
  url = "applications/${azuread_application.application.object_id}"

  body = {
    displayName = "Demo App Updated"
  }
}