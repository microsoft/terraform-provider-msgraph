## 0.3.0 (Unreleased)

FEATURES:
- **New Authentication Method**: Azure PowerShell authentication support via `use_powershell` provider attribute

ENHANCEMENTS:
- provider: Added support for authenticating with Azure PowerShell via the `use_powershell` attribute and `ARM_USE_POWERSHELL` environment variable. This provides an alternative to Azure CLI authentication without the client ID permission limitations ([#67](https://github.com/microsoft/terraform-provider-msgraph/issues/67))
- `msgraph_resource`: Support `moved` block to move resources from `azuread` provider to `msgraph` provider.

DEPENDENCIES:
- Updated `github.com/Azure/azure-sdk-for-go/sdk/azidentity` from v1.8.0 to v1.13.0 to enable Azure PowerShell authentication support
- Updated `github.com/Azure/azure-sdk-for-go/sdk/azcore` from v1.16.0 to v1.19.1

BUG FIXES:
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