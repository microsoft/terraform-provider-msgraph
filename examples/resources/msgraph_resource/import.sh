# MSGraph resource can be imported using the resource id, e.g.
terraform import msgraph_resource.servicePrincipal /servicePrincipals/00000000-0000-0000-0000-000000000000
terraform import msgraph_resource.member /groups/group-id/members/$ref/00000000-0000-0000-0000-000000000000

# For resources that use the beta API, append ?api-version=beta to the import ID:
terraform import msgraph_resource.settings '/settings/00000000-0000-0000-0000-000000000000?api-version=beta'
