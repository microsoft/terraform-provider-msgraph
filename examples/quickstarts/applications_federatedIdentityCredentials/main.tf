terraform {
  required_providers {
    msgraph = {
      source = "microsoft/msgraph"
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
}

resource "msgraph_resource" "federatedIdentityCredential" {
  # url = "applications/{id}/federatedIdentityCredentials"
  url = "applications/${msgraph_resource.application.id}/federatedIdentityCredentials"
  body = {
    name        = "myFederatedIdentityCredentials"
    description = "My test federated identity credentials"
    audiences   = ["https://myapp.com"]
    issuer      = "https://sts.windows.net/00000000-0000-0000-0000-000000000000/"
    subject     = "00000000-0000-0000-0000-000000000000"
  }
}
