package services

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
	"github.com/microsoft/terraform-provider-msgraph/internal/docstrings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MSGraphDataSource{}

func NewMSGraphDataSource() datasource.DataSource {
	return &MSGraphDataSource{}
}

// MSGraphDataSource defines the data source implementation.
type MSGraphDataSource struct {
	client *clients.MSGraphClient
}

// MSGraphDataSourceModel describes the data source data model.
type MSGraphDataSourceModel struct {
	Id                   types.String      `tfsdk:"id"`
	ApiVersion           types.String      `tfsdk:"api_version"`
	Url                  types.String      `tfsdk:"url"`
	ResponseExportValues map[string]string `tfsdk:"response_export_values"`
	Headers              types.Map         `tfsdk:"headers"`
	QueryParameters      types.Map         `tfsdk:"query_parameters"`
	Output               types.Dynamic     `tfsdk:"output"`
	Timeouts             timeouts.Value    `tfsdk:"timeouts"`
}

func (r *MSGraphDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *MSGraphDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "This data source can list resources or read an individual resource from the Microsoft Graph API.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the resource. Normally, it is in the format of UUID if it is a single resource. If it is a collection resource, it will be the URL of the collection.",
				Computed:            true,
			},

			"url": schema.StringAttribute{
				MarkdownDescription: docstrings.Url("data"),
				Required:            true,
			},

			"api_version": schema.StringAttribute{
				MarkdownDescription: docstrings.ApiVersion(),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("v1.0", "beta"),
				},
			},

			"response_export_values": schema.MapAttribute{
				MarkdownDescription: docstrings.ResponseExportValues(),
				Optional:            true,
				ElementType:         types.StringType,
			},

			"headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "A map of headers to include in the request",
			},

			"query_parameters": schema.MapAttribute{
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				Optional:            true,
				MarkdownDescription: "A map of query parameters to include in the request",
			},

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

func (r *MSGraphDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if v, ok := req.ProviderData.(*clients.Client); ok {
		r.client = v.MSGraphClient
	}
}

func (r *MSGraphDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model MSGraphDataSourceModel
	if resp.Diagnostics.Append(req.Config.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := model.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancelRead := context.WithTimeout(ctx, readTimeout)
	defer cancelRead()

	apiVersion := "v1.0"
	if model.ApiVersion.ValueString() != "" {
		apiVersion = model.ApiVersion.ValueString()
	}

	options := clients.NewRequestOptions(AsMapOfString(model.Headers), AsMapOfLists(model.QueryParameters))
	responseBody, err := r.client.Read(ctx, model.Url.ValueString(), apiVersion, options)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read data source", err.Error())
		return
	}

	responseId := model.Url.ValueString()
	if responseBody != nil {
		if responseMap, ok := responseBody.(map[string]interface{}); ok {
			if idValue, ok := responseMap["id"]; ok && idValue != nil {
				if idString, ok := idValue.(string); ok {
					responseId = idString
				}
			}
		}
	}

	model.Id = types.StringValue(responseId)
	model.Output = types.DynamicValue(buildOutputFromBody(responseBody, model.ResponseExportValues))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
