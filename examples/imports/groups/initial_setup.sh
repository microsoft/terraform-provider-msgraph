terraform init
terraform apply -auto-approve
terraform state rm msgraph_resource.group
terraform import msgraph_resource.group "/groups/<GUID>"