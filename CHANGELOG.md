## 0.2.0 (Unreleased)

FEATURES:
- **New Resource**: msgraph_update_resource

ENHANCEMENTS:
- `msgraph` resources and data sources now support `retry` configuration to handle transient failures.
- `msgraph` resource and data source: support for `timeouts` configuration block

BUG FIXES:
- Fixed an issue where `msgraph_resource` resource did not wait for the resource to be fully provisioned before completing.
- Fixed an issue with the `msgraph_resource` resource could not detect resource drift.

## 0.1.0

FEATURES:
- **New Data Source**: msgraph_resource
- **New Resource**: msgraph_resource