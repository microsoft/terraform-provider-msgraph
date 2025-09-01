package docstrings

import "fmt"

func ApiVersion() string {
	return "The API version of the data source. The allowed values are `v1.0` and `beta`. Defaults to `v1.0`."
}

func Url(kind string) string {
	switch kind {
	case "data":
		return "The URL of the data source. It supports both collection URL which is used to list resources, for example `/users`, and item URL which is used to read an individual resource, for example `/users/{id}`."
	case "resource":
		return `The URL which is used to manage the resource. It supports two types of URLs:  
  - Collection URL which is used to make a POST request to create a new resource, for example, "/users", it must support the following operations:
	- Create a new resource: POST "/users"
    - Read an existing resource: GET "/users/{id}"
    - Update an existing resource: PATCH "/users/{id}"
    - Delete an existing resource: DELETE "/users/{id} "
  - URL which has a "$ref" suffix, for example, "/groups/{group-id}/members/$ref", it must support the following operations:
	- Add a reference to a resource: POST "/groups/{group-id}/members/$ref"
	- Remove a reference to a resource: DELETE "/groups/{group-id}/members/{id}/$ref"
  
  More information about the Microsoft Graph API can be found at [Microsoft Graph API](https://docs.microsoft.com/en-us/graph/overview).  
  And there are some [examples](https://github.com/microsoft/terraform-provider-msgraph/tree/main/examples/quickstarts) to help you get started.
`
	case "update_resource":
		return `The item URL of the existing resource to update, for example "/users/{id}".

	This resource performs PATCH requests against the item URL and expects the following operations to be supported by the API endpoint:
		- Read an existing resource: GET "/users/{id}"
		- Update an existing resource: PATCH "/users/{id}"

	More information about the Microsoft Graph API can be found at [Microsoft Graph API](https://docs.microsoft.com/en-us/graph/overview).  
	There are also [examples](https://github.com/microsoft/terraform-provider-msgraph/tree/main/examples/quickstarts) to help you get started.`
	default:
		return ""
	}
}

func Body() string {
	return "A dynamic attribute that contains the request body."
}

func Output() string {
	return fmt.Sprintf(`
The output HCL object containing the properties specified in %[1]sresponse_export_values%[1]s. Here are some examples to use the values.

	%[1]s%[1]s%[1]sterraform
	 output "app_id" {
	   // it will output the value of app_id
	   value = msgraph_resource.application.output.app_id
	 }
	 
	 output "all" {
	   // it will output the whole response
	   value = msgraph_resource.application.output.all
	 }
	%[1]s%[1]s%[1]s`, "`")
}

func ResponseExportValues() string {
	return fmt.Sprintf(`A map where the key is the name for the result and the value is a JMESPath query string to filter the response. Here's an example. If it sets to %[1]s{"all" = "@", "app_id" = "appId"}%[1]s, it will set the following HCL object to the computed property output.

	%[1]s%[1]s%[1]stext
	{
		"all" = {
			"appId" = "00000000-0000-0000-0000-000000000000"
			"displayName" = "example"
			"id" = "00000000-0000-0000-0000-000000000000"
			...
		}
		"app_id" = "00000000-0000-0000-0000-000000000000"
	}
	%[1]s%[1]s%[1]s

To learn more about JMESPath, visit [JMESPath](https://jmespath.org/).
`, "`")
}

func ResourceID() string {
	return "The ID of the resource. Normally, it is in the format of UUID."
}
