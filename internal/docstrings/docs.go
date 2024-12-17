package docstrings

import "fmt"

func ApiVersion() string {
	return "The API version of the data source. The allowed values are `v1.0` and `beta`. Defaults to `v1.0`."
}

func Url(kind string) string {
	if kind == "data" {
		return "The URL of the data source. It supports both collection URL which is used to list resources, for example `/users`, and item URL which is used to read an individual resource, for example `/users/{id}`."
	}
	return "The collection URL of the resource. For example, `/users`, `/groups`, `/applications`."
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
