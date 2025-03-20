// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package services

import (
	"context"

	"github.com/azure/terraform-provider-msgraph/internal/clients"
	"github.com/azure/terraform-provider-msgraph/internal/docstrings"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Id                   types.String        `tfsdk:"id"`
	ApiVersion           types.String        `tfsdk:"api_version"`
	Url                  types.String        `tfsdk:"url"`
	ResponseExportValues map[string]string   `tfsdk:"response_export_values"`
	Headers              map[string]string   `tfsdk:"headers"`
	QueryParameters      map[string][]string `tfsdk:"query_parameters"`
	Output               types.Dynamic       `tfsdk:"output"`
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
				MarkdownDescription: "The ID of the MSGraph resource",
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

	// support pagination
	apiVersion := "v1.0"
	if model.ApiVersion.ValueString() != "" {
		apiVersion = model.ApiVersion.ValueString()
	}
	responseBody, err := r.client.Read(ctx, model.Url.ValueString(), apiVersion, clients.NewRequestOptions(model.Headers, model.QueryParameters))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read data source", err.Error())
		return
	}

	model.Id = model.Url
	model.Output = types.DynamicValue(buildOutputFromBody(responseBody, model.ResponseExportValues))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
