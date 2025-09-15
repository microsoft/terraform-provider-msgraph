package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
	"github.com/microsoft/terraform-provider-msgraph/internal/docstrings"
	"github.com/microsoft/terraform-provider-msgraph/internal/retry"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource               = &MSGraphResourceAction{}
	_ resource.ResourceWithModifyPlan = &MSGraphResourceAction{}
)

func NewMSGraphResourceAction() resource.Resource {
	return &MSGraphResourceAction{}
}

// MSGraphResourceAction defines the resource implementation.
type MSGraphResourceAction struct {
	client *clients.MSGraphClient
}

// MSGraphResourceActionModel describes the resource data model.
type MSGraphResourceActionModel struct {
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

func (r *MSGraphResourceAction) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_action"
}

func (r *MSGraphResourceAction) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "This resource can perform any Microsoft Graph API action. Use this for operations like password resets, sending emails, or other one-time actions.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: docstrings.ResourceID(),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"resource_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the resource to perform the action on. This should be the full resource path, for example `applications/12345678-1234-1234-1234-123456789abc` or `users/user@example.com`. You can use the `resource_url` output from `msgraph_resource`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"action": schema.StringAttribute{
				MarkdownDescription: "The action to perform on the resource. This is the action path that will be appended to the resource URL, for example `addPassword`, `sendMail`, `changePassword`, or `members/$ref`. Leave empty for actions directly on the resource.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"method": schema.StringAttribute{
				MarkdownDescription: "The HTTP method to use for the action. Common methods include `GET`, `POST`, `PATCH`, `DELETE`, and `PUT`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodPut),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"api_version": schema.StringAttribute{
				MarkdownDescription: docstrings.ApiVersion(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("v1.0", "beta"),
				},
				Default: stringdefault.StaticString("v1.0"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				Create: true,
			}),
		},
	}
}

func (r *MSGraphResourceAction) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if v, ok := req.ProviderData.(*clients.Client); ok {
		r.client = v.MSGraphClient
	}
}

func (r *MSGraphResourceAction) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	var plan *MSGraphResourceActionModel
	if response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...); response.Diagnostics.HasError() {
		return
	}

	var state *MSGraphResourceActionModel
	if response.Diagnostics.Append(request.State.Get(ctx, &state)...); response.Diagnostics.HasError() {
		return
	}
}

func (r *MSGraphResourceAction) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *MSGraphResourceActionModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := model.Timeouts.Create(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Construct the full URL from resource_url and action
	fullUrl := model.ResourceUrl.ValueString()
	if !model.Action.IsNull() && model.Action.ValueString() != "" {
		fullUrl = fmt.Sprintf("%s/%s", fullUrl, model.Action.ValueString())
	}

	// Use the full URL as the ID for this action resource
	model.Id = types.StringValue(fullUrl)

	// Execute the action
	if err := r.executeAction(ctx, model); err != nil {
		resp.Diagnostics.AddError("Failed to execute action", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *MSGraphResourceAction) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since this is an action resource, update should re-execute the action
	var model *MSGraphResourceActionModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := model.Timeouts.Create(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Re-execute the action
	if err := r.executeAction(ctx, model); err != nil {
		resp.Diagnostics.AddError("Failed to execute action", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// executeAction is a helper function that performs the actual API call
func (r *MSGraphResourceAction) executeAction(ctx context.Context, model *MSGraphResourceActionModel) error {
	// Prepare request body
	var requestBody interface{}
	if !model.Body.IsNull() && !model.Body.IsUnknown() {
		if err := unmarshalBody(model.Body, &requestBody); err != nil {
			return fmt.Errorf("failed to unmarshal body: %w", err)
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

	// Log the action
	tflog.Info(ctx, fmt.Sprintf("Executing %s action on %s", model.Method.ValueString(), fullUrl))

	// Execute the action
	responseBody, err := r.client.Action(ctx, model.Method.ValueString(), fullUrl, model.ApiVersion.ValueString(), requestBody, options)
	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}

	// Build output from response
	model.Output = types.DynamicValue(buildOutputFromBody(responseBody, model.ResponseExportValues))

	return nil
}

func (r *MSGraphResourceAction) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *MSGraphResourceActionModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	// For action resources, read is essentially a no-op since actions are one-time operations
	// We'll just maintain the current state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *MSGraphResourceAction) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// For action resources, delete is typically a no-op since actions are one-time operations
	var model *MSGraphResourceActionModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	// Log the deletion (no actual action needed for most cases)
	tflog.Info(ctx, fmt.Sprintf("Deleting action resource %s", model.Id.ValueString()))
}
