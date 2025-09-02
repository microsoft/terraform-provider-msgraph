terraform {
  required_providers {
    msgraph = {
      source = "Microsoft/msgraph"
    }
  }
}

provider "msgraph" {}

# This example creates an application first (so we have something to update),
# then uses msgraph_update_resource to PATCH its displayName.

resource "msgraph_resource" "application" {
  url = "applications"
  body = {
    displayName = "Demo App"
  }

  # We ignore the displayName change here because the update is handled by
  # the separate msgraph_update_resource below.
  lifecycle {
    ignore_changes = [body.displayName]
  }
}

resource "msgraph_update_resource" "application_update" {
  # Point directly at the item URL you want to PATCH
  url = "applications/${msgraph_resource.application.id}"

  body = {
    displayName = "Demo App Updated"
  }
}