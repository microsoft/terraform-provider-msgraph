package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
	"github.com/microsoft/terraform-provider-msgraph/internal/docstrings"
	"github.com/microsoft/terraform-provider-msgraph/internal/retry"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MSGraphResourceActionDataSource{}

func NewMSGraphResourceActionDataSource() datasource.DataSource {
	return &MSGraphResourceActionDataSource{}
}

// MSGraphResourceActionDataSource defines the data source implementation.
type MSGraphResourceActionDataSource struct {
	client *clients.MSGraphClient
}

// MSGraphResourceActionDataSourceModel describes the data source data model.
type MSGraphResourceActionDataSourceModel struct {
	Id                   types.String      `tfsdk:"id"`
	ApiVersion           types.String      `tfsdk:"api_version"`
	ResourceUrl          types.String      `tfsdk:"resource_url"`
	Action               types.String      `tfsdk:"action"`
	Method               types.String      `tfsdk:"method"`
	Body                 types.Dynamic     `tfsdk:"body"`
	QueryParameters      types.Map         `tfsdk:"query_parameters"`
	Headers              types.Map         `tfsdk:"headers"`
	ResponseExportValues map[string]string `tfsdk:"response_export_values"`
	Retry                retry.Value       `tfsdk:"retry"`
	Output               types.Dynamic     `tfsdk:"output"`
	Timeouts             timeouts.Value    `tfsdk:"timeouts"`
}

func (r *MSGraphResourceActionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_action"
}

func (r *MSGraphResourceActionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "This data source can perform any Microsoft Graph API action and return the result. Use this for read-only operations like retrieving calculated values, checking status, or performing queries.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: docstrings.ResourceID(),
				Computed:            true,
			},

			"resource_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the resource to perform the action on. This should be the full resource path, for example `applications/12345678-1234-1234-1234-123456789abc` or `users/user@example.com`. You can use the `resource_url` output from `msgraph_resource`.",
				Required:            true,
			},

			"action": schema.StringAttribute{
				MarkdownDescription: "The action to perform on the resource. This is the action path that will be appended to the resource URL, for example `getMemberGroups`, `checkMemberGroups`, `calculateDisplayNames`, or `members`. Leave empty for actions directly on the resource.",
				Optional:            true,
			},

			"method": schema.StringAttribute{
				MarkdownDescription: "The HTTP method to use for the action. For data sources, this is typically `GET` or `POST` for actions that require a request body.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(http.MethodGet, http.MethodPost),
				},
			},

			"api_version": schema.StringAttribute{
				MarkdownDescription: docstrings.ApiVersion(),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("v1.0", "beta"),
				},
			},

			"body": schema.DynamicAttribute{
				MarkdownDescription: docstrings.Body(),
				Optional:            true,
			},

			"query_parameters": schema.MapAttribute{
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				Optional:            true,
				MarkdownDescription: "A mapping of query parameters to be sent with the action request.",
			},

			"headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "A mapping of HTTP headers to be sent with the action request. Note that authentication headers are automatically handled.",
			},

			"response_export_values": schema.MapAttribute{
				MarkdownDescription: docstrings.ResponseExportValues(),
				Optional:            true,
				ElementType:         types.StringType,
			},

			"retry": retry.Schema(ctx),

			"output": schema.DynamicAttribute{
				MarkdownDescription: docstrings.Output(),
				Computed:            true,
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Read: true,
			}),
		},
	}
}

func (r *MSGraphResourceActionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if v, ok := req.ProviderData.(*clients.Client); ok {
		r.client = v.MSGraphClient
	}
}

func (r *MSGraphResourceActionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model *MSGraphResourceActionDataSourceModel
	if resp.Diagnostics.Append(req.Config.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := model.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	// Prepare request body
	var requestBody interface{}
	if !model.Body.IsNull() && !model.Body.IsUnknown() {
		if err := unmarshalBody(model.Body, &requestBody); err != nil {
			resp.Diagnostics.AddError("Failed to unmarshal body", err.Error())
			return
		}
	}

	// Prepare request options
	options := clients.RequestOptions{
		Headers:         AsMapOfString(model.Headers),
		QueryParameters: clients.NewQueryParameters(AsMapOfLists(model.QueryParameters)),
		RetryOptions:    clients.NewRetryOptions(model.Retry),
	}

	// Construct the full URL from resource_url and action
	fullUrl := model.ResourceUrl.ValueString()
	if !model.Action.IsNull() && model.Action.ValueString() != "" {
		fullUrl = fmt.Sprintf("%s/%s", fullUrl, model.Action.ValueString())
	}

	// Default to GET method if not specified
	method := model.Method.ValueString()
	if method == "" {
		method = http.MethodGet
	}

	// Default to v1.0 API version if not specified
	apiVersion := model.ApiVersion.ValueString()
	if apiVersion == "" {
		apiVersion = "v1.0"
	}

	// Log the action
	tflog.Info(ctx, fmt.Sprintf("Executing %s action on %s", method, fullUrl))

	// Execute the action
	responseBody, err := r.client.Action(ctx, method, fullUrl, apiVersion, requestBody, options)
	if err != nil {
		resp.Diagnostics.AddError("API call failed", err.Error())
		return
	}

	// Use the full URL as the ID for this action data source
	model.Id = types.StringValue(fullUrl)

	// Build output from response
	model.Output = types.DynamicValue(buildOutputFromBody(responseBody, model.ResponseExportValues))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
