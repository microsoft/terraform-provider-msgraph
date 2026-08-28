## 0.5.0

ENHANCEMENTS:
- Added sovereign cloud support via optional `environment` (`ARM_ENVIRONMENT`) for `public/global`, `usgovernment/usgovernmentl4`, `usgovernmentl5/dod`, and `china`, including cloud-based Graph endpoint/token-scope resolution and `@odata.id` relationship URLs. ([#147](https://github.com/microsoft/terraform-provider-msgraph/pull/147))
- `msgraph_resource`: Added a `create_method` attribute (`POST` default, or `PUT`) to support Graph resources that are created via `PUT` on the resource URL.

DEPENDENCIES:
- Updated `github.com/Azure/azure-sdk-for-go/sdk/azcore` from v1.21.1 to v1.22.0
- Updated `github.com/Azure/azure-sdk-for-go/sdk/azidentity` from v1.13.1 to v1.14.0
- Updated `github.com/hashicorp/terraform-plugin-log` from v0.10.0 to v0.11.0
- Updated Go from 1.25.8 to 1.26.5.

BUG FIXES:
- `msgraph_resource_collection`: Fixed a "Provider produced inconsistent final plan" error on `reference_ids` that occurred when only some referenced resources were updated in the same apply (a mix of known and known-after-apply values). Because a `/$ref` collection is unordered, reads now preserve the ordering already recorded in state, keeping the stored order aligned with the configured order so pinned positions no longer shift between plan and apply. ([#135](https://github.com/microsoft/terraform-provider-msgraph/issues/135))
- `msgraph_resource`: Fixed an issue where updating a nested object (complex type) sent only the changed sub-fields, causing Microsoft Graph to reset omitted siblings to their defaults. The provider now sends the full nested object when any of its fields change (e.g. `federatedIdentityCredential.claimsMatchingExpression`). ([#137](https://github.com/microsoft/terraform-provider-msgraph/issues/137))
- `msgraph_resource`: Fixed an issue where importing a resource that represents a `$ref` relationship sent an empty request body. ([#148](https://github.com/microsoft/terraform-provider-msgraph/pull/148))
- Capped `MaxRetryDelay` at 600s so that Microsoft Graph `Retry-After` values up to 315s (seen on Identity Governance endpoints) are honoured instead of being truncated. ([#119](https://github.com/microsoft/terraform-provider-msgraph/pull/119))

## 0.4.0

ENHANCEMENTS:
- `msgraph_resource_collection`: Support resource import for collection resources (e.g. groups/{id}/members/$ref) to allow importing existing relationships into Terraform state.
- `msgraph_resource_collection`: Added support for `skip_destroy` attribute to remove the resource from state without removing the references on destroy. This allows deleting a parent resource (such as a group) that would otherwise fail with constraints like "The group must have at least one owner, hence this owner cannot be removed." ([#78](https://github.com/microsoft/terraform-provider-msgraph/issues/78))
- Retry of transient failures (`429`, `408`, `500`, `502`, `503` and `504`) is now applied by default regardless of whether a `retry` block is configured.

DEPENDENCIES:
- Updated `github.com/Azure/azure-sdk-for-go/sdk/azcore` from v1.19.1 to v1.21.1
- Updated `github.com/Azure/azure-sdk-for-go/sdk/azidentity` from v1.13.0 to v1.13.1
- Updated `github.com/hashicorp/terraform-plugin-framework` from v1.13.0 to v1.19.0
- Updated `github.com/hashicorp/terraform-plugin-framework-validators` from v0.15.0 to v0.19.0
- Updated `github.com/hashicorp/terraform-plugin-framework-timeouts` from v0.5.0 to v0.7.0
- Updated `github.com/hashicorp/terraform-plugin-go` from v0.25.0 to v0.31.0
- Updated `github.com/hashicorp/terraform-plugin-sdk/v2` from v2.35.0 to v2.40.1
- Updated `github.com/hashicorp/terraform-plugin-testing` from v1.11.0 to v1.16.0
- Updated `github.com/hashicorp/terraform-plugin-docs` from v0.20.1 to v0.25.0

BUG FIXES:
- Fixed an issue where transient errors such as `429 Too Many Requests` were not retried during create, delete, and the create/delete consistency checks, causing operations to fail under throttling. ([#129](https://github.com/microsoft/terraform-provider-msgraph/issues/129))
- `msgraph_resource`: Fixed an issue where creating a resource failed to determine its ID when the create response contained no top-level `id` field. The provider now falls back to the `Location` header returned by the API. ([#107](https://github.com/microsoft/terraform-provider-msgraph/issues/107))

## 0.3.0

FEATURES:
- **New Authentication Method**: Azure PowerShell authentication support via `use_powershell` provider attribute

ENHANCEMENTS:
- `msgraph_resource`: Added support for `update_method` attribute to allow choosing between `PATCH` (default) and `PUT` for update operations.
- `msgraph_update_resource`: Added support for `update_method` attribute to allow choosing between `PATCH` (default) and `PUT` for update operations.
- provider: Added support for authenticating with Azure PowerShell via the `use_powershell` attribute and `ARM_USE_POWERSHELL` environment variable. This provides an alternative to Azure CLI authentication without the client ID permission limitations ([#67](https://github.com/microsoft/terraform-provider-msgraph/issues/67))
- `msgraph_resource`: Support `moved` block to move resources from `azuread` provider to `msgraph` provider.
- `msgraph_resource`: Added support for waiting for creation/deletion consistency.

DEPENDENCIES:
- Updated `github.com/Azure/azure-sdk-for-go/sdk/azidentity` from v1.8.0 to v1.13.0 to enable Azure PowerShell authentication support
- Updated `github.com/Azure/azure-sdk-for-go/sdk/azcore` from v1.16.0 to v1.19.1

BUG FIXES:
- Fixed an issue where `msgraph_resource` failed to update resources when external changes are detected, specifically when clearing array fields ([#58](https://github.com/microsoft/terraform-provider-msgraph/issues/58))
- Fixed an issue where `msgraph_resource` failed to track state for `$ref` resources (relationships), causing drift detection failures ([#68](https://github.com/microsoft/terraform-provider-msgraph/issues/68))
- Fixed an issue where `@odata.type` property was missing in PATCH requests for resources that require it (e.g. Named Locations) ([#59](https://github.com/microsoft/terraform-provider-msgraph/issues/59))

## 0.2.0

FEATURES:
- **New Resource**: msgraph_update_resource
- **New Resource**: msgraph_resource_collection
- **New Resource**: msgraph_resource_action
- **New Data Source**: msgraph_resource_action

ENHANCEMENTS:
- `msgraph` resources and data sources now support `retry` configuration to handle transient failures.
- `msgraph` resource and data source: support for `timeouts` configuration block.
- `msgraph_resource` and `msgraph_update_resource` resources: support for `ignore_missing_property` field.
- `msgraph` resource and data source: support for `timeouts` configuration block
- `msgraph_resource`: Update operations now send only changed fields in the request body to Microsoft Graph (minimal PATCH payloads) reducing unnecessary updates.
- `msgraph_update_resource`: Create operations send the full body, while subsequent updates send only changed fields computed from prior state.
- `msgraph_resource`: Added `resource_url` computed attribute that provides the full URL path to the resource instance.

BUG FIXES:
- Fixed an issue where `msgraph_resource` resource did not wait for the resource to be fully provisioned before completing.
- Fixed an issue with the `msgraph_resource` resource could not detect resource drift.
- Fixed an issue that 200 OK responses were not being handled correctly when deleting resources.

## 0.1.0

FEATURES:
- **New Data Source**: msgraph_resource
- **New Resource**: msgraph_resource