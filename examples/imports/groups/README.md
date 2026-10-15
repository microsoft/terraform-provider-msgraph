# Importing Existing Groups Example

This example demonstrates how to import an existing Microsoft Graph group into Terraform without modifying its properties.

## Overview

When you have existing groups in Microsoft Graph that you want to manage with Terraform, you need to import them carefully to avoid unintended changes. The `importProperties` query parameter allows you to specify exactly which properties Terraform should track and manage.

## Why Use `importProperties`?

The `msgraph_resource` uses a dynamic `body` attribute, which means Terraform doesn't automatically know which properties to track during import. Without specifying properties:

- Importing with an empty body would cause Terraform to want to remove all properties on the next plan
- Importing all properties would force you to manage everything, even properties you don't care about

The `importProperties` parameter solves this by:

1. Seeding specific properties in the Terraform state as `null` values
2. Allowing the Read operation to populate them with actual values from Microsoft Graph
3. Enabling you to manage only the properties you specify while ignoring others

## Workflow

### Step 1: Initial Setup (Optional)

If you want to test the import workflow, you can create a test group first:

```bash
bash initial_setup.sh
```

This script:

1. Initializes Terraform
2. Creates a group using `main.tf`
3. Removes it from Terraform state (simulating an existing resource)
4. Shows you the group ID for import

### Step 2: Configure Your Import

Edit `import.tf` and specify:

- The resource to import to (e.g., `msgraph_resource.group`)
- The group ID from Microsoft Graph
- The properties you want to manage using `importProperties`

Example:

```hcl
import {
    to = msgraph_resource.group
    id = "/groups/e1c41ddc-7947-4591-ae98-ed63e6739e42?importProperties=displayName,mailEnabled,mailNickname,securityEnabled"
}
```

### Step 3: Configure Your Resource

In `main.tf`, define the resource with the properties you want to manage:

```hcl
resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "My Group"
    mailEnabled     = false
    mailNickname    = "mygroup"
    securityEnabled = true
  }
}
```

**Important:** The properties in your `body` should match the properties specified in `importProperties`.

### Step 4: Generate Import Plan

```bash
terraform plan -generate-config-out=generated.tf
```

This will show you what Terraform will import and generate a configuration file with the current values.

### Step 5: Apply the Import

```bash
terraform apply
```

Terraform will import the group and update the state. If your configuration matches the existing group, there should be no changes.

## Alternative: Using CLI Import

You can also import without using an `import` block:

```bash
terraform import msgraph_resource.group "/groups/00000000-0000-0000-0000-000000000000?importProperties=displayName,mailEnabled,mailNickname,securityEnabled"
```

## Best Practices

1. **Specify Only Managed Properties**: Only include properties in `importProperties` that you want Terraform to manage and track for drift detection.

2. **Use `ignore_missing_property`**: Keep the default `ignore_missing_property = true` to prevent Terraform from detecting drift on properties not in your configuration.

3. **Match Configuration to Reality**: Ensure your `body` configuration matches the actual values in Microsoft Graph to avoid unintended changes on first apply.

4. **Test First**: If possible, test the import workflow in a non-production environment first.

5. **Nested Properties**: For nested properties, use dot notation:
   ```
   importProperties=displayName,web.redirectUris,web.homePageUrl
   ```

## Common Properties by Resource Type

### Groups

```
displayName,mailEnabled,mailNickname,securityEnabled,description,groupTypes
```

### Applications

```
displayName,signInAudience,web.redirectUris,requiredResourceAccess
```

### Service Principals

```
displayName,appId,servicePrincipalType,accountEnabled
```

### Users

```
displayName,userPrincipalName,accountEnabled,mailNickname
```

## Troubleshooting

### "No changes" after import but properties don't match

- Verify that `ignore_missing_property` is `true` (default)
- Check that the properties in `importProperties` match those in your `body`

### Terraform wants to change properties after import

- Ensure the values in your `body` match the actual values in Microsoft Graph
- Verify you included all managed properties in `importProperties`

### Import fails with "Invalid Import ID"

- Ensure the ID format is correct: `/groups/{id}` or `/groups/{id}?importProperties=...`
- Check that query parameters are properly URL-encoded if needed

## Additional Resources

- [Microsoft Graph Groups API](https://learn.microsoft.com/en-us/graph/api/resources/group)
- [Terraform Import Documentation](https://developer.hashicorp.com/terraform/language/import)
- [msgraph_resource Documentation](https://registry.terraform.io/providers/Microsoft/msgraph/latest/docs/resources/resource)
