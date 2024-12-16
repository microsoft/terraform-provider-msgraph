// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ExampleResource{}
var _ resource.ResourceWithImportState = &ExampleResource{}
var _ resource.ResourceWithConfigValidators = &ExampleResource{}

func NewExampleResource() resource.Resource {
	return &ExampleResource{}
}

// ExampleResource defines the resource implementation.
type ExampleResource struct {
	client *http.Client
}

func (r *ExampleResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

// ExampleResourceModel describes the resource data model.
type ExampleResourceModel struct {
	Id        types.String    `tfsdk:"id"`
	Name      types.String    `tfsdk:"name"`
	ParentId  types.String    `tfsdk:"parent_id"`
	V20230101 *V20230101Model `tfsdk:"v2023_01_01"`
	V20240101 *V20240101Model `tfsdk:"v2024_01_01"`
}

type V20230101Model struct {
	Properties V20230101PropertiesModel `tfsdk:"properties"`
}

type V20230101PropertiesModel struct {
	P0 types.String `tfsdk:"p0"`
	P1 types.String `tfsdk:"p1"`
	P2 types.String `tfsdk:"p2"`
}

type V20240101Model struct {
	Properties V20240101PropertiesModel `tfsdk:"properties"`
}

type V20240101PropertiesModel struct {
	P0 types.String `tfsdk:"p0"`
	P1 types.String `tfsdk:"p1"`
	P2 types.String `tfsdk:"p2"`
	P3 types.String `tfsdk:"p3"`
}

func (r *ExampleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_example"
}

func (r *ExampleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"parent_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"v2023_01_01": schema.SingleNestedBlock{
				Blocks: map[string]schema.Block{
					"properties": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"p0": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},

							"p1": schema.StringAttribute{
								Optional: true,
							},

							"p2": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},

			"v2024_01_01": schema.SingleNestedBlock{
				Blocks: map[string]schema.Block{
					"properties": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"p0": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},

							"p1": schema.StringAttribute{
								Optional: true,
							},

							"p2": schema.StringAttribute{
								Optional: true,
							},

							"p3": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *ExampleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ExampleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ExampleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue("example-id")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExampleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ExampleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExampleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ExampleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExampleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ExampleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *ExampleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
