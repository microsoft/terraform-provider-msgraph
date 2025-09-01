package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
	"github.com/microsoft/terraform-provider-msgraph/internal/docstrings"
	"github.com/microsoft/terraform-provider-msgraph/internal/dynamic"
	"github.com/microsoft/terraform-provider-msgraph/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                     = &MSGraphUpdateResource{}
	_ resource.ResourceWithConfigValidators = &MSGraphUpdateResource{}
	_ resource.ResourceWithModifyPlan       = &MSGraphUpdateResource{}
)

func NewMSGraphUpdateResource() resource.Resource {
	return &MSGraphUpdateResource{}
}

// MSGraphUpdateResource defines the resource implementation.
type MSGraphUpdateResource struct {
	client *clients.MSGraphClient
}

func (r *MSGraphUpdateResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

// MSGraphUpdateResourceModel describes the resource data model.
type MSGraphUpdateResourceModel struct {
	Id                    types.String      `tfsdk:"id"`
	ApiVersion            types.String      `tfsdk:"api_version"`
	Url                   types.String      `tfsdk:"url"`
	Body                  types.Dynamic     `tfsdk:"body"`
	UpdateQueryParameters types.Map         `tfsdk:"update_query_parameters"`
	ReadQueryParameters   types.Map         `tfsdk:"read_query_parameters"`
	ResponseExportValues  map[string]string `tfsdk:"response_export_values"`
	Output                types.Dynamic     `tfsdk:"output"`
}

func (r *MSGraphUpdateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_update_resource"
}

func (r *MSGraphUpdateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource can manage a subset of any existing Microsoft Graph resource's properties.\n\n" +
			"-> **Note** This resource is used to add or modify properties on an existing resource. When `msgraph_update_resource` is deleted, no operation will be performed, and these properties will stay unchanged. If you want to restore the modified properties to some values, you must apply the restored properties before deleting.",
		Description: "This resource can manage a subset of any existing Microsoft Graph resource's properties.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: docstrings.ResourceID(),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"url": schema.StringAttribute{
				MarkdownDescription: docstrings.Url("update_resource"),
				Required:            true,
			},

			"api_version": schema.StringAttribute{
				MarkdownDescription: docstrings.ApiVersion(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("v1.0", "beta"),
				},
				Default: stringdefault.StaticString("v1.0"),
			},

			"body": schema.DynamicAttribute{
				MarkdownDescription: docstrings.Body(),
				Optional:            true,
			},

			"update_query_parameters": schema.MapAttribute{
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				Optional:            true,
				MarkdownDescription: "A mapping of query parameters to be sent with the update request.",
			},

			"read_query_parameters": schema.MapAttribute{
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				Optional:            true,
				MarkdownDescription: "A mapping of query parameters to be sent with the read request.",
			},

			"response_export_values": schema.MapAttribute{
				MarkdownDescription: docstrings.ResponseExportValues(),
				Optional:            true,
				ElementType:         types.StringType,
			},

			"output": schema.DynamicAttribute{
				MarkdownDescription: docstrings.Output(),
				Computed:            true,
			},
		},
	}
}

func (r *MSGraphUpdateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if v, ok := req.ProviderData.(*clients.Client); ok {
		r.client = v.MSGraphClient
	}
}

func (r *MSGraphUpdateResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	var plan *MSGraphUpdateResourceModel
	if response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...); response.Diagnostics.HasError() {
		return
	}

	var state *MSGraphUpdateResourceModel
	if response.Diagnostics.Append(request.State.Get(ctx, &state)...); response.Diagnostics.HasError() {
		return
	}
}

func (r *MSGraphUpdateResource) CreateUpdate(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	var model MSGraphUpdateResourceModel
	var stateModel *MSGraphUpdateResourceModel
	diagnostics.Append(plan.Get(ctx, &model)...)
	diagnostics.Append(state.Get(ctx, &stateModel)...)
	if diagnostics.HasError() {
		return
	}

	data, err := dynamic.ToJSON(model.Body)
	if err != nil {
		diagnostics.AddError("Failed to marshal body", err.Error())
		return
	}
	var requestBody interface{}
	if err = json.Unmarshal(data, &requestBody); err != nil {
		diagnostics.AddError("Failed to unmarshal body", err.Error())
		return
	}

	options := clients.NewRequestOptions(nil, AsMapOfLists(model.UpdateQueryParameters))
	_, err = r.client.Update(ctx, model.Url.ValueString(), model.ApiVersion.ValueString(), requestBody, options)
	if err != nil {
		diagnostics.AddError("Failed to create resource", err.Error())
		return
	}

	options = clients.NewRequestOptions(nil, AsMapOfLists(model.ReadQueryParameters))
	responseBody, err := r.client.Read(ctx, model.Url.ValueString(), model.ApiVersion.ValueString(), options)
	if err != nil {
		diagnostics.AddError("Failed to read data source", err.Error())
		return
	}
	model.Output = types.DynamicValue(buildOutputFromBody(responseBody, model.ResponseExportValues))
	model.Id = types.StringValue(utils.LastSegment(model.Url.ValueString()))
	diagnostics.Append(state.Set(ctx, &model)...)
}

func (r *MSGraphUpdateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	r.CreateUpdate(ctx, request.Plan, &response.State, &response.Diagnostics)
}

func (r *MSGraphUpdateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.CreateUpdate(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *MSGraphUpdateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *MSGraphUpdateResourceModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	if model.ApiVersion.ValueString() == "" {
		model.ApiVersion = types.StringValue("v1.0")
	}

	options := clients.NewRequestOptions(nil, AsMapOfLists(model.ReadQueryParameters))
	responseBody, err := r.client.Read(ctx, model.Url.ValueString(), model.ApiVersion.ValueString(), options)
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			tflog.Info(ctx, fmt.Sprintf("Error reading %q - removing from state", model.Id.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read data source", err.Error())
		return
	}

	state := model
	state.Output = types.DynamicValue(buildOutputFromBody(responseBody, model.ResponseExportValues))

	if !model.Body.IsNull() {
		requestBody := make(map[string]interface{})
		if err := unmarshalBody(model.Body, &requestBody); err != nil {
			resp.Diagnostics.AddError("Invalid body", fmt.Sprintf(`The argument "body" is invalid: %s`, err.Error()))
			return
		}

		option := utils.UpdateJsonOption{
			IgnoreCasing:          false,
			IgnoreMissingProperty: false,
			IgnoreNullProperty:    false,
		}
		body := utils.UpdateObject(requestBody, responseBody, option)

		data, err := json.Marshal(body)
		if err != nil {
			resp.Diagnostics.AddError("Invalid body", err.Error())
			return
		}
		payload, err := dynamic.FromJSON(data, model.Body.UnderlyingValue().Type(ctx))
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to parse payload: %s", err.Error()))
			payload, err = dynamic.FromJSONImplied(data)
			if err != nil {
				resp.Diagnostics.AddError("Invalid payload", err.Error())
				return
			}
		}
		state.Body = payload
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MSGraphUpdateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
