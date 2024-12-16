// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"github.com/azure/terraform-provider-msgraph/internal/clients"
	"github.com/azure/terraform-provider-msgraph/internal/dynamic"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	Url    types.String  `tfsdk:"url"`
	Id     types.String  `tfsdk:"id"`
	Output types.Dynamic `tfsdk:"output"`
}

func (r *MSGraphDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *MSGraphDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Example identifier",
				Computed:            true,
			},

			"type": schema.StringAttribute{
				Description: "The type of the data source, for example: `users/todo/lists/tasks@v1.0`",
				Required:    true,
			},

			"name": schema.StringAttribute{
				Description: "The name of the data source",
				Optional:    true,
			},

			"parent_id": schema.StringAttribute{
				Description: "The parent identifier",
				Optional:    true,
			},

			"resource_id": schema.StringAttribute{
				Description: "The resource identifier",
				Optional:    true,
			},

			"output": schema.DynamicAttribute{
				Computed: true,
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

	out, err := r.client.Read(ctx, model.Url.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read data source", err.Error())
		return
	}

	model.Id = model.Url

	data, _ := json.Marshal(out)
	model.Output, _ = dynamic.FromJSONImplied(data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
