## 0.2.0 (Unreleased)

ENHANCEMENTS:
- `msgraph` resources and data sources now support `retry` configuration to handle transient failures.

BUG FIXES:
- Fixed an issue where `msgraph_resource` resource did not wait for the resource to be fully provisioned before completing.

## 0.1.0

FEATURES:
- **New Data Source**: msgraph_resource
- **New Resource**: msgraph_resource