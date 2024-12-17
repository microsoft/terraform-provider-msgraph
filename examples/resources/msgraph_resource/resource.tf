terraform {
  required_providers {
    msgraph = {
      source = "Microsoft/msgraph"
    }
  }
}

provider "msgraph" {
}

resource "msgraph_resource" "application" {
  url = "applications"
  body = {
    displayName = "My Application"
  }
  response_export_values = {
    all    = "@"
    app_id = "appId"
  }
}

output "app_id" {
  value = msgraph_resource.application.output.app_id
}

output "all" {
  // it will output the whole response
  value = msgraph_resource.application.output.all
}
