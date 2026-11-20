---
page_title: "Importing Existing Resources"
subcategory: "Guides"
description: |-
  Guide for importing existing Microsoft Graph resources into Terraform
---

# Importing Existing Resources

This guide explains how to import existing Microsoft Graph resources into Terraform without modifying their properties.

## Overview

When you have existing resources in Microsoft Graph that you want to manage with Terraform, you need to import them into your Terraform state. The `msgraph_resource` provider offers a flexible import mechanism that allows you to control exactly which properties Terraform should manage.

## The Challenge with Dynamic Attributes

The `msgraph_resource` uses a dynamic `body` attribute to support any Microsoft Graph resource type. This flexibility creates a challenge during import: Terraform doesn't automatically know which properties to track.

### Without Proper Import

If you import a resource without specifying properties:

```bash
# ❌ This will cause issues
terraform import msgraph_resource.group /groups/00000000-0000-0000-0000-000000000000
```

The resource is imported with a null `body`. Terraform then sees every property in your configuration as absent from state and plans an in-place update, sending a redundant `PATCH` that rewrites properties which already hold the right values:

```text
  # msgraph_resource.group will be updated in-place
  # (imported from "/groups/00000000-0000-0000-0000-000000000000")
  ~ resource "msgraph_resource" "group" {
      + body = {
          + displayName     = "My Group"
          + mailEnabled     = false
        }
      ~ output = {} -> (known after apply)
    }

Plan: 1 to import, 0 to add, 1 to change, 0 to destroy.
```

### The Solution: `importProperties`

The `importProperties` query parameter solves this by allowing you to specify exactly which properties to track:

```bash
# ✅ This imports only the properties you want to manage
terraform import msgraph_resource.group "/groups/00000000-0000-0000-0000-000000000000?importProperties=displayName,mailEnabled,mailNickname,securityEnabled"
```

## How It Works

When you use `importProperties`:

1. **Seeding**: The specified properties are seeded in the Terraform state with `null` values
2. **Reading**: Terraform performs a Read operation, populating these properties with actual values from Microsoft Graph
3. **Tracking**: Only the specified properties are tracked for drift detection
4. **Ignoring**: Other properties are ignored (when `ignore_missing_property = true`, which is the default)

## Import Methods

### Method 1: Using Import Blocks (Recommended for Terraform 1.5+)

Create an `import.tf` file:

```hcl
import {
  to = msgraph_resource.group
  id = "/groups/00000000-0000-0000-0000-000000000000?importProperties=displayName,mailEnabled,mailNickname,securityEnabled"
}
```

Then define your resource in your main configuration:

```hcl
resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "My Existing Group"
    mailEnabled     = false
    mailNickname    = "mygroup"
    securityEnabled = true
  }
}
```

Run the import:

```bash
terraform plan
terraform apply
```

### Method 2: Using CLI Import

```bash
terraform import msgraph_resource.group "/groups/00000000-0000-0000-0000-000000000000?importProperties=displayName,mailEnabled,mailNickname,securityEnabled"
```

## Step-by-Step Import Workflow

### Step 1: Identify the Resource

Find the resource ID in Microsoft Graph. You can use:

- Azure Portal
- Microsoft Graph Explorer
- Azure CLI: `az ad group show --group "My Group"`
- PowerShell: `Get-MgGroup -Filter "displayName eq 'My Group'"`

### Step 2: Determine Properties to Manage

Decide which properties you want Terraform to manage. Consider:

- **Required properties**: Properties needed for the resource to function
- **Managed properties**: Properties you want to control via Terraform
- **Ignored properties**: Properties managed elsewhere or by other processes

Example for a group:

```
Manage: displayName, mailEnabled, mailNickname, securityEnabled
Ignore: id, createdDateTime, members, owners (managed separately)
```

### Step 3: Create Resource Configuration

Define the resource with the properties you want to manage:

```hcl
resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "My Existing Group"
    mailEnabled     = false
    mailNickname    = "mygroup"
    securityEnabled = true
  }
}
```

**Important**: Set the values to match the current state in Microsoft Graph to avoid unintended changes.

### Step 4: Import the Resource

Using import block:

```hcl
import {
  to = msgraph_resource.group
  id = "/groups/{group-id}?importProperties=displayName,mailEnabled,mailNickname,securityEnabled"
}
```

Or using CLI:

```bash
terraform import msgraph_resource.group "/groups/{group-id}?importProperties=displayName,mailEnabled,mailNickname,securityEnabled"
```

### Step 5: Verify the Import

Run a plan to verify no changes are detected:

```bash
terraform plan
```

If changes are detected, update your configuration to match the actual values in Microsoft Graph.

## Advanced Import Scenarios

### Importing with Nested Properties

For nested properties, use dot notation:

```bash
terraform import msgraph_resource.app "/applications/{app-id}?importProperties=displayName,web.redirectUris,web.homePageUrl,requiredResourceAccess"
```

Configuration:

```hcl
resource "msgraph_resource" "app" {
  url = "applications"
  body = {
    displayName = "My App"
    web = {
      redirectUris = ["https://example.com/callback"]
      homePageUrl  = "https://example.com"
    }
    requiredResourceAccess = [
      # ... resource access configuration
    ]
  }
}
```

#### Overlapping Property Paths

If a property and one of its children are both listed, the more specific path wins, whichever order you write them in. These two are equivalent:

```bash
importProperties=web,web.redirectUris
importProperties=web.redirectUris,web
```

Both seed `body` as `{ web = { redirectUris = ... } }`, so Terraform manages `redirectUris` alone and leaves the other fields of `web` untouched.

To manage the whole `web` object instead, list only the parent:

```bash
importProperties=displayName,web
```

Seeding a property that holds an object pulls in every field the API returns for it, so `body` will contain the complete `web` object after the first read. Bear in mind that your configuration then has to account for all of those fields.

Paths containing empty segments — `web..redirectUris`, `web.`, `.web` — are ignored. The remaining properties are still seeded, so a typo silently narrows what gets imported rather than failing the import; check the resulting `body` if a property you expected is missing.

### Importing with Different API Versions

```bash
terraform import msgraph_resource.example "/groups/{group-id}?api-version=beta&importProperties=displayName,mailEnabled"
```

Configuration:

```hcl
resource "msgraph_resource" "example" {
  url         = "groups"
  api_version = "beta"
  body = {
    displayName = "My Group"
    mailEnabled = false
  }
}
```

### Importing Relationship References

For `$ref` relationships:

```bash
terraform import msgraph_resource.member "/groups/{group-id}/members/$ref/{member-id}"
```

Configuration:

```hcl
resource "msgraph_resource" "member" {
  url = "groups/{group-id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/directoryObjects/{member-id}"
  }
}
```

## Common Import Patterns

### Groups

```bash
terraform import msgraph_resource.group "/groups/{id}?importProperties=displayName,mailEnabled,mailNickname,securityEnabled,description,groupTypes"
```

### Applications

```bash
terraform import msgraph_resource.app "/applications/{id}?importProperties=displayName,signInAudience,web.redirectUris"
```

### Service Principals

```bash
terraform import msgraph_resource.sp "/servicePrincipals/{id}?importProperties=displayName,appId,servicePrincipalType,accountEnabled"
```

### Users

```bash
terraform import msgraph_resource.user "/users/{id}?importProperties=displayName,userPrincipalName,accountEnabled,mailNickname"
```

## Best Practices

### 1. Start with Minimal Properties

Import only the properties you need to manage:

```bash
# ✅ Good: Only essential properties
importProperties=displayName,mailEnabled,mailNickname,securityEnabled

# ❌ Avoid: Too many properties
importProperties=displayName,id,createdDateTime,deletedDateTime,mail,proxyAddresses,...
```

### 2. Use `ignore_missing_property`

Keep the default `ignore_missing_property = true` to prevent drift detection on unmanaged properties:

```hcl
resource "msgraph_resource" "group" {
  url                    = "groups"
  ignore_missing_property = true  # This is the default
  body = {
    displayName = "My Group"
  }
}
```

### 3. Match Configuration to Reality

Before importing, retrieve the current values and set them in your configuration:

```bash
# Get current values
az ad group show --group {id} --query "{displayName:displayName, mailEnabled:mailEnabled}"
```

Then use these values in your configuration to avoid changes on first apply.

### 4. Test in Non-Production First

Always test your import workflow in a non-production environment before applying to production resources.

### 5. Document Your Imports

Keep a record of which properties you're managing for each resource type:

```hcl
# groups.tf
# Managed properties: displayName, mailEnabled, mailNickname, securityEnabled
# Ignored properties: members (managed via msgraph_resource_collection)
resource "msgraph_resource" "group" {
  # ...
}
```

## Troubleshooting

### No Changes After Import but Properties Don't Match

**Symptom**: After import, `terraform plan` shows no changes, but the properties in your config don't match Microsoft Graph.

**Cause**: `ignore_missing_property = true` is ignoring the differences.

**Solution**: Verify that the properties in `importProperties` match those in your `body` configuration.

### Terraform Wants to Change Properties After Import

**Symptom**: After import, `terraform plan` shows changes to properties.

**Cause**: The values in your configuration don't match the actual values in Microsoft Graph.

**Solution**:

1. Retrieve the current values from Microsoft Graph
2. Update your configuration to match
3. Run `terraform plan` again to verify

### Import Fails with "Invalid Import ID"

**Symptom**: Import command fails with an error about invalid ID format.

**Cause**: Incorrect ID format or syntax.

**Solution**: Ensure the ID follows the correct format:

- Regular resources: `/resourceType/{id}`
- With parameters: `/resourceType/{id}?importProperties=prop1,prop2`
- Relationships: `/resourceType/{id}/relationship/$ref/{ref-id}`

### Properties Not Being Tracked

**Symptom**: Changes to properties in Microsoft Graph are not detected by Terraform.

**Cause**: Properties not included in `importProperties` during import.

**Solution**:

1. Remove the resource from state: `terraform state rm msgraph_resource.example`
2. Re-import with the correct properties in `importProperties`

## Examples

See the [examples/imports](https://github.com/microsoft/terraform-provider-msgraph/tree/main/examples/imports) directory for complete working examples of importing various resource types.

## Additional Resources

- [msgraph_resource Documentation](https://registry.terraform.io/providers/Microsoft/msgraph/latest/docs/resources/resource)
- [Terraform Import Documentation](https://developer.hashicorp.com/terraform/language/import)
- [Microsoft Graph API Reference](https://learn.microsoft.com/en-us/graph/api/overview)
