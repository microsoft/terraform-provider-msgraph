## 0.2.0 (Unreleased)

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