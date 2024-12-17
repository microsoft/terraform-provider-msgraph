---
layout: "msgraph"
page_title: "MSGraph Provider: Authenticating via the Azure CLI"
description: |-
  This guide will cover how to use the Azure CLI as authentication for the MSGraph Provider.

---

# MSGraph Provider: Authenticating using the Azure CLI

## Important Notes about Authenticating using the Azure CLI

* Terraform only supports authenticating using the `az` CLI (and this must be available on your PATH) - authenticating using the older `azure` CLI or PowerShell Cmdlets are not supported.
* Authenticating via the Azure CLI is only supported when using a User Account. If you're using a Service Principal (for example via `az login --service-principal`) you should instead authenticate via the Service Principal directly (either using a [Client Secret](service_principal_client_secret.md) or a [Client Certificate](service_principal_client_certificate.md)).

---

## Logging into the Azure CLI

~> **Note**: If you're using the **China**, **German** or **Government** Azure Clouds - you'll need to first configure the Azure CLI to work with that Cloud.  You can do this by running:

```shell
$ az cloud set --name AzureChinaCloud|AzureGermanCloud|AzureUSGovernment
```

---

Login to the Azure CLI using:

```shell
az login --allow-no-subscriptions
```

The `--allow-no-subscriptions` argument enables access to tenants that have no linked subscriptions, in addition to tenants that do.


---




Once logged in - it's possible to list the Subscriptions and Tenants associated with the account via:

```shell-session
$ az account list -o table --all --query "[].{TenantID: tenantId, Subscription: name, Default: isDefault}"
```

The output (similar to below) will display one or more Tenants and/or Subscriptions.

```
TenantID                              Subscription                         Default
------------------------------------  -----------------------------------  ---------
00000000-0000-1111-1111-111111111111  N/A(tenant level account)            False
00000000-0000-2222-2222-222222222222  N/A(tenant level account)            False
00000000-0000-1111-1111-111111111111  My Subscription                      True
00000000-0000-1111-1111-111111111111  My Other Subscription                False
```

Each entry shown is referred to as an `Azure CLI account`, which represents either a subscription with its linked tenant, or a tenant without any accessible subscriptions (Azure CLI does not show tenant names or domains). The provider will select the tenant ID from your default Azure CLI account. If you have more than one tenant listed in the output of `az account list`, for example if you are a guest user in other tenants, you can specify the tenant to use.

```shell-session
# sh
export ARM_TENANT_ID=00000000-0000-2222-2222-222222222222
```
```powershell
# PowerShell
$env:ARM_TENANT_ID = 00000000-0000-2222-2222-222222222222
```

You can also configure the tenant ID from within the provider block.

```hcl
provider "msgraph" {
  tenant_id = "00000000-0000-2222-2222-222222222222"
}
```

Alternatively, you can configure the Azure CLI to default to the tenant you are managing with Terraform.

```shell-session
$ az login --allow-no-subscriptions --tenant "TENANT_ID_OR_DOMAIN"
```

<br>

-> **Tenants and Subscriptions** The MSGraph provider operates on tenants and not on subscriptions. We recommend always specifying `az login --allow-no-subscriptions` as it will force the Azure CLI to report tenants with no associated subscriptions, or where your user account does not have any roles assigned for a subscription.

---

## Configuring Azure CLI authentication in Terraform

No specific configuration is required for the provider to use Azure CLI authentication. If you're looking to use Terraform across Tenants - it's possible to do this by configuring the `tenant_id` field in the Provider block, as shown below:

```hcl
provider "msgraph" {
  tenant_id = "00000000-0000-1111-1111-111111111111"
}
```

More information on [the fields supported in the Provider block can be found here](../index.html#argument-reference).

At this point running either `terraform plan` or `terraform apply` should allow Terraform to run using the Azure CLI to authenticate.

## Disabling Azure CLI authentication

For compatibility reasons and to ensure a positive user experience when running Terraform interactively, Azure CLI authentication is enabled by default. It's possible to disable authentication using Azure CLI, which you may wish to do in automated environments such as CI/CD pipelines or when scripting operations with Terraform.

To do so, add the `use_cli` configuration property in the Provider block.

```hcl
provider "msgraph" {
  use_cli = false
}
```

Alternatively, you can set the `ARM_USE_CLI` environment variable.

```shell
# sh
export ARM_USE_CLI=false

# PowerShell
$env:ARM_USE_CLI = false
```
