terraform {
  required_providers {
    msgraph = {
      source = "Microsoft/msgraph"
    }
  }
}

provider "msgraph" {
}


data "msgraph_resource" "me" {
  url = "me"
  response_export_values = {
    all          = "@"
    display_name = "displayName"
  }
}

output "display_name" {
  // it will output "John Doe"
  value = data.msgraph_resource.me.output.display_name
}

output "all" {
  // it will output the whole response
  value = msgraph_resource.application.output.all
}
