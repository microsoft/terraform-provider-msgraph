# MSGraph resource collection can be imported using the collection $ref URL, e.g.
terraform import msgraph_resource_collection.group_members groups/00000000-0000-0000-0000-000000000000/members/$ref

# To import using the beta API version, append the api-version query parameter:
terraform import msgraph_resource_collection.group_members 'groups/00000000-0000-0000-0000-000000000000/members/$ref?api-version=beta'

# To import using custom collection type, append the collection-type query parameter:
terraform import msgraph_resource_collection.test 'identityGovernance/entitlementManagement/accessPackages/00000000-0000-0000-0000-000000000000/accessPackages/$ref?collection-type=identityGovernance/entitlementManagement/accessPackages'
