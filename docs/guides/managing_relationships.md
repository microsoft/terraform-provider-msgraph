---
layout: "msgraph"
page_title: "MSGraph Provider: Managing Relationships"
subcategory: "Configuration"
---

# Managing Relationships with the MSGraph Provider

Microsoft Graph APIs often involve managing relationships between resources, such as group members, group owners, application roles, and other reference collections. The MSGraph provider offers three different approaches to manage these relationships, each with its own benefits and use cases.

## Overview of Approaches

1. **[Resource Collection](#resource-collection-recommended)** - Manage entire relationship collections declaratively
2. **[Individual Resource References](#individual-resource-references)** - Manage single relationships 
3. **[Inline During Resource Creation](#inline-during-resource-creation)** - Set relationships as part of the initial resource creation

## Resource Collection (Recommended)

The `msgraph_resource_collection` resource is the recommended approach for managing relationship collections. It provides declarative management where you specify the complete desired state of a collection, and Terraform ensures the actual state matches by adding missing items and removing extra ones.

### Benefits

- **Declarative**: Define the complete desired state of the collection
- **Atomic**: All changes are applied together
- **Clean lifecycle**: Properly handles resource destruction without orphaned references
- **Efficient**: Minimizes API calls by managing the entire collection

### Example: Managing Group Members

```terraform
resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Development Team"
    mailEnabled     = false
    mailNickname    = "dev-team"
    securityEnabled = true
  }
}

resource "msgraph_resource" "user1" {
  url = "users"
  body = {
    userPrincipalName = "alice@example.com"
    displayName       = "Alice Smith"
    mailNickname      = "alice"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

resource "msgraph_resource" "user2" {
  url = "users"
  body = {
    userPrincipalName = "bob@example.com"
    displayName       = "Bob Johnson"
    mailNickname      = "bob"
    passwordProfile = {
      forceChangePasswordNextSignIn = true
      password                      = "TempPassword123!"
    }
  }
}

# Manage all group members as a collection
resource "msgraph_resource_collection" "group_members" {
  url           = "groups/${msgraph_resource.group.id}/members/$ref"
  reference_ids = [
    msgraph_resource.user1.id,
    msgraph_resource.user2.id
  ]
}
```

### Example: Managing Group Owners

```terraform
resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Project Alpha"
    mailEnabled     = false
    mailNickname    = "project-alpha"
    securityEnabled = true
  }
}

data "azurerm_client_config" "current" {}

# Manage group owners using resource collection
resource "msgraph_resource_collection" "group_owners" {
  url           = "groups/${msgraph_resource.group.id}/owners/$ref"
  reference_ids = [
    data.azurerm_client_config.current.object_id,
    msgraph_resource.manager_user.id
  ]
}
```

### Working with Service Principals

When managing service principals as group members, use the `beta` API version due to a known Microsoft Graph issue:

```terraform
resource "msgraph_resource_collection" "group_members" {
  url         = "groups/${msgraph_resource.group.id}/members/$ref"
  api_version = "beta"  # Required for service principals
  reference_ids = [
    msgraph_resource.sp_a.id,
    msgraph_resource.sp_b.id
  ]
}
```

## Individual Resource References

Use `msgraph_resource` with `$ref` URLs to manage individual relationships. This approach is useful when you need fine-grained control over each relationship or when relationships are managed conditionally.

### Benefits

- **Fine-grained control**: Manage each relationship independently
- **Conditional management**: Use Terraform conditionals for specific relationships
- **Incremental changes**: Add or remove relationships without affecting others

### Limitations

- **Deletion constraints**: Some Microsoft Graph resources have constraints (e.g., groups must have at least one owner)
- **More complex state**: Multiple resources to track instead of one collection

### Example: Adding Individual Group Members

```terraform
resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    displayName     = "Finance Team"
    mailEnabled     = false
    mailNickname    = "finance-team"
    securityEnabled = true
  }
}

# Add individual members
resource "msgraph_resource" "member1" {
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/users/${msgraph_resource.user1.id}"
  }
}

resource "msgraph_resource" "member2" {
  url = "groups/${msgraph_resource.group.id}/members/$ref"
  body = {
    "@odata.id" = "https://graph.microsoft.com/v1.0/users/${msgraph_resource.user2.id}"
  }
}
```

## Inline During Resource Creation

Set relationships directly in the resource body during creation using OData bind syntax. This approach is most useful for initial setup and avoiding circular dependencies.

### Benefits

- **Atomic creation**: Resource and relationships created together
- **Avoids constraints**: No issues with minimum ownership requirements
- **Simple initial setup**: Good for bootstrapping resources with required relationships

### Limitations

- **Creation-time only**: Relationships set this way cannot be easily updated later
- **Less flexible**: Harder to manage dynamic relationship changes
- **Limited to creation**: Cannot add/remove relationships after resource creation using this method

### Example: Group with Initial Owners and Members

```terraform
data "azurerm_client_config" "current" {}

resource "msgraph_resource" "group_with_relationships" {
  url = "groups"
  body = {
    displayName     = "DevOps Team"
    mailEnabled     = false
    mailNickname    = "devops-team"
    securityEnabled = true
    
    # Set initial owners
    "owners@odata.bind" = [
      "https://graph.microsoft.com/v1.0/users/${data.azurerm_client_config.current.object_id}",
      "https://graph.microsoft.com/v1.0/users/${local.team_lead_id}"
    ]
    
    # Set initial members
    "members@odata.bind" = [
      "https://graph.microsoft.com/v1.0/users/${local.developer1_id}",
      "https://graph.microsoft.com/v1.0/users/${local.developer2_id}"
    ]
  }
}
```

### Example: Application with Initial Owners

```terraform
resource "msgraph_resource" "application_with_owners" {
  url = "applications"
  body = {
    displayName = "API Gateway Application"
    
    # Set application owners during creation
    "owners@odata.bind" = [
      "https://graph.microsoft.com/v1.0/servicePrincipals/${data.azurerm_client_config.current.object_id}",
      "https://graph.microsoft.com/v1.0/users/${local.app_admin_id}"
    ]
  }
}
```

## Best Practices and Recommendations

### When to Use Each Approach

| Scenario | Recommended Approach | Reason |
|----------|---------------------|---------|
| Managing complete group memberships | Resource Collection | Declarative, handles all members atomically |
| Dynamic membership based on conditions | Resource Collection | Easy to update the complete list |
| Initial group setup with required owners | Inline During Creation | Avoids deletion constraints |
| Adding single relationships conditionally | Individual Resources | Fine-grained control |
| Managing application role assignments | Resource Collection | Clean lifecycle management |
| Bootstrapping resources with dependencies | Inline During Creation | Avoids circular dependencies |

### General Guidelines

1. **Prefer Resource Collection**: Use `msgraph_resource_collection` for most relationship management scenarios
2. **Use Inline for Initial Setup**: Set critical relationships (like required owners) during resource creation
3. **Handle Constraints**: Be aware of Microsoft Graph constraints like minimum owner requirements
4. **Use Beta API When Needed**: Service principal memberships require the beta API version
5. **Plan for Destruction**: Consider the order of resource destruction to avoid constraint violations

### Error Handling

Common errors and solutions:

**"The group must have at least one owner"**
- Solution: Use inline creation to set initial owners, or ensure resource destruction order

**"Service principals not visible in group members"**
- Solution: Use `api_version = "beta"` for service principal memberships

**"Circular dependency"**
- Solution: Use inline creation for one direction of the relationship

## Conclusion

The MSGraph provider offers flexible approaches for managing relationships. Choose the method that best fits your use case, with `msgraph_resource_collection` being the recommended approach for most scenarios due to its declarative nature and clean lifecycle management.