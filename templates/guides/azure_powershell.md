---
layout: "msgraph"
page_title: "MSGraph Provider: Authenticating via Azure PowerShell"
subcategory: "Authentication"
description: |-
  This guide will cover how to use Azure PowerShell as authentication for the MSGraph Provider.

---

# Authenticating using Azure PowerShell

Terraform supports a number of different methods for authenticating to Azure:

* [Authenticating to Azure using the Azure CLI](azure_cli.html)
* Authenticating to Azure using Azure PowerShell (covered in this guide)
* [Authenticating to Azure using Managed Identity](managed_service_identity.html)
* [Authenticating to Azure using a Service Principal and a Client Certificate](service_principal_client_certificate.html)
* [Authenticating to Azure using a Service Principal and a Client Secret](service_principal_client_secret.html)
* [Authenticating to Azure using a Service Principal and OpenID Connect](service_principal_oidc.html)

---

We recommend using either a Service Principal or Managed Identity when running Terraform non-interactively (such as when running Terraform in a CI server) - and authenticating using Azure PowerShell or the Azure CLI when running Terraform locally.

## Important Notes about Authenticating using Azure PowerShell

* Terraform requires the `Az.Accounts` PowerShell module to be installed (version 2.2.0 or later).
* PowerShell Core (`pwsh`) is supported on all platforms (Windows, macOS, Linux), while Windows PowerShell is supported on Windows only.

---

## Installing Azure PowerShell

### Prerequisites

First, ensure you have PowerShell installed:

**Windows**: PowerShell 5.1 or later is included, or install [PowerShell Core](https://learn.microsoft.com/powershell/scripting/install/installing-powershell-on-windows)

**macOS**:
```shell
brew install powershell
```

**Linux**: See [Installing PowerShell on Linux](https://learn.microsoft.com/powershell/scripting/install/installing-powershell-on-linux)

### Install the Az.Accounts Module

Once PowerShell is installed, install the Az.Accounts module:

```powershell
Install-Module -Name Az.Accounts -Repository PSGallery -Force
```

Verify the installation:

```powershell
Get-Module -Name Az.Accounts -ListAvailable
```

---

## Logging into Azure PowerShell

-> **Using other clouds** If you're using the **China**, **German** or **Government** Azure Clouds - you'll need to first configure Azure PowerShell to work with that Cloud, so that the correct authentication service is used. You can do this by running: <br><br>`Connect-AzAccount -Environment AzureChinaCloud|AzureGermanCloud|AzureUSGovernment`

---

Firstly, login to Azure PowerShell using a User, Service Principal or Managed Identity.

User Account:

```powershell
Connect-AzAccount
```

User Account with Specific Tenant:

```powershell
Connect-AzAccount -Tenant "00000000-0000-1111-1111-111111111111"
```

Service Principal with a Secret:

```powershell
$clientId = "00000000-0000-0000-0000-000000000000"
$clientSecret = "MyCl1eNtSeCr3t"
$tenantId = "10000000-2000-3000-4000-500000000000"

$securePassword = ConvertTo-SecureString $clientSecret -AsPlainText -Force
$credential = New-Object System.Management.Automation.PSCredential($clientId, $securePassword)

Connect-AzAccount -ServicePrincipal -Credential $credential -Tenant $tenantId
```

Service Principal with a Certificate:

```powershell
$clientId = "00000000-0000-0000-0000-000000000000"
$tenantId = "10000000-2000-3000-4000-500000000000"
$thumbprint = "CERTIFICATE_THUMBPRINT"

Connect-AzAccount -ServicePrincipal -ApplicationId $clientId -CertificateThumbprint $thumbprint -Tenant $tenantId
```

Managed Identity:

```powershell
Connect-AzAccount -Identity
```

Or with a specific User-Assigned Managed Identity:

```powershell
Connect-AzAccount -Identity -AccountId "00000000-0000-0000-0000-000000000000"
```

---

Once logged in - it's possible to list the contexts associated with the account via:

```powershell
Get-AzContext -ListAvailable | Format-Table -Property Name, Account, Tenant
```

The output (similar to below) will display one or more contexts with their associated tenants.

```
Name                                   Account                      Tenant
----                                   -------                      ------
Tenant 1 (00000000-0000-1111-1111...) user@example.com             00000000-0000-1111-1111-111111111111
Tenant 2 (00000000-0000-2222-2222...) user@example.com             00000000-0000-2222-2222-222222222222
```

If you have more than one tenant, you can specify the tenant to use.

```powershell
# PowerShell
$env:ARM_TENANT_ID = "00000000-0000-2222-2222-222222222222"
```
```shell
# sh
export ARM_TENANT_ID=00000000-0000-2222-2222-222222222222
```

You can also configure the tenant ID from within the provider block.

```hcl
provider "msgraph" {
  tenant_id = "00000000-0000-2222-2222-222222222222"
}
```

Alternatively, you can configure Azure PowerShell to use a specific tenant by setting the context.

```powershell
Set-AzContext -Tenant "00000000-0000-2222-2222-222222222222"
```

<br>

-> **Tenants and Subscriptions** The MSGraph provider operates on tenants and not on subscriptions. Azure PowerShell authentication works directly with tenant-level access.

---

## Configuring Azure PowerShell authentication in Terraform

To enable Azure PowerShell authentication, configure the `use_powershell` field in the Provider block:

```hcl
provider "msgraph" {
  use_powershell = true
  tenant_id      = "00000000-0000-1111-1111-111111111111"
}
```

Alternatively, you can enable it via the `ARM_USE_POWERSHELL` environment variable:

```powershell
# PowerShell
$env:ARM_USE_POWERSHELL = "true"
```
```shell
# sh
export ARM_USE_POWERSHELL=true
```

More information on [the fields supported in the Provider block can be found here](../index.html#argument-reference).

At this point running either `terraform plan` or `terraform apply` should allow Terraform to run using Azure PowerShell to authenticate.

## Disabling Azure PowerShell authentication

For compatibility reasons, Azure PowerShell authentication is disabled by default. If you've enabled it and wish to disable it again (for example in automated environments), you can do so with the `use_powershell` configuration property.

```hcl
provider "msgraph" {
  use_powershell = false
}
```

Alternatively, you can set the `ARM_USE_POWERSHELL` environment variable.

```powershell
# PowerShell
$env:ARM_USE_POWERSHELL = "false"
```
```shell
# sh
export ARM_USE_POWERSHELL=false
```

## Troubleshooting

### PowerShell not found

**Error**: `executable not found on path`

**Solution**: Ensure PowerShell is installed and available in your PATH. Verify with:

```shell
# Check if pwsh (PowerShell Core) is available
pwsh -version

# On Windows, you can also check for Windows PowerShell
powershell -version
```

### Az.Accounts module not found

**Error**: `Az.Accounts module not found`

**Solution**: Install the Az.Accounts module:

```powershell
Install-Module -Name Az.Accounts -Repository PSGallery -Force
```

Verify installation:

```powershell
Get-Module -Name Az.Accounts -ListAvailable
```

### Not logged in to Azure PowerShell

**Error**: `Please run "Connect-AzAccount" to set up account`

**Solution**: Log in to Azure PowerShell:

```powershell
Connect-AzAccount
```

Verify you're logged in:

```powershell
Get-AzContext
```

### Insufficient permissions

**Error**: `required scopes are missing in the token`

**Solution**: Ensure your user account has the necessary Microsoft Graph permissions assigned. Contact your Azure AD administrator to grant the required permissions or use a Service Principal with the appropriate API permissions configured.

### Multiple tenants

If you have access to multiple tenants, explicitly specify the tenant:

```hcl
provider "msgraph" {
  use_powershell = true
  tenant_id      = "00000000-0000-1111-1111-111111111111"
}
```

Or set the context in PowerShell before running Terraform:

```powershell
Set-AzContext -Tenant "00000000-0000-1111-1111-111111111111"
```
