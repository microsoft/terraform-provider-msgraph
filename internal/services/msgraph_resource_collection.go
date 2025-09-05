package services

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"

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
	"github.com/microsoft/terraform-provider-msgraph/internal/myplanmodifier"
	"github.com/microsoft/terraform-provider-msgraph/internal/myvalidator"
	"github.com/microsoft/terraform-provider-msgraph/internal/retry"
	"github.com/microsoft/terraform-provider-msgraph/internal/utils"
)

var (
	_ resource.Resource               = &MSGraphResourceCollection{}
	_ resource.ResourceWithConfigure  = &MSGraphResourceCollection{}
	_ resource.ResourceWithModifyPlan = &MSGraphResourceCollection{}
)

func NewMSGraphResourceCollection() resource.Resource {
	return &MSGraphResourceCollection{}
}

type MSGraphResourceCollection struct{ client *clients.MSGraphClient }

type MSGraphResourceCollectionModel struct {
	Id                   types.String      `tfsdk:"id"`
	ApiVersion           types.String      `tfsdk:"api_version"`
	Url                  types.String      `tfsdk:"url"`
	ReferenceIds         types.List        `tfsdk:"reference_ids"`
	ReadQueryParameters  types.Map         `tfsdk:"read_query_parameters"`
	Retry                retry.Value       `tfsdk:"retry"`
	ResponseExportValues map[string]string `tfsdk:"response_export_values"`
	Output               types.Dynamic     `tfsdk:"output"`
	Timeouts             timeouts.Value    `tfsdk:"timeouts"`
}

func (r *MSGraphResourceCollection) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_collection"
}

func (r *MSGraphResourceCollection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage the full contents of a child reference collection (such as group members or owners) for an existing Microsoft Graph resource. Missing items are added; extra remote items are removed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of this managed collection. This is the normalized collection URL with the trailing '/$ref' removed (e.g. for 'groups/{group-id}/members/$ref' the id becomes 'groups/{group-id}/members').",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"url": schema.StringAttribute{
				MarkdownDescription: "Full relative path of the target reference collection ending in '/$ref'. For example: `groups/{group-id}/members/$ref`. This must point to a $ref collection; changing this value forces a new resource.",
				Required:            true,
				Validators: []validator.String{
					myvalidator.ResourceCollectionURL(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"api_version": schema.StringAttribute{
				MarkdownDescription: docstrings.ApiVersion(),
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.OneOf("v1.0", "beta")},
				Default:             stringdefault.StaticString("v1.0"),
			},

			"reference_ids": schema.ListAttribute{
				MarkdownDescription: "List of object IDs that MUST exist in this `$ref` collection. Missing IDs are added; extra remote items are removed. Order is ignored. Each value should be the GUID (or string identifier) of an existing directory object (user, group, service principal, etc.).",
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers:       []planmodifier.List{myplanmodifier.OrderInsensitiveStringList()},
			},

			"read_query_parameters": schema.MapAttribute{
				ElementType:         types.ListType{ElemType: types.StringType},
				Optional:            true,
				MarkdownDescription: "A mapping of query parameters to be sent with the read (list) requests.",
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
			"timeouts": timeouts.BlockAll(ctx),
		},
	}
}

func (r *MSGraphResourceCollection) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if v, ok := req.ProviderData.(*clients.Client); ok {
		r.client = v.MSGraphClient
	}
}

func (r *MSGraphResourceCollection) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	var plan, state *MSGraphResourceCollectionModel
	if response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...); response.Diagnostics.HasError() {
		return
	}
	if response.Diagnostics.Append(request.State.Get(ctx, &state)...); response.Diagnostics.HasError() {
		return
	}
	if plan == nil || state == nil {
		return
	}

	plan.Output = state.Output
	if !plan.ReferenceIds.Equal(state.ReferenceIds) || !reflect.DeepEqual(plan.ResponseExportValues, state.ResponseExportValues) {
		plan.Output = types.DynamicUnknown()
	}

	response.Diagnostics.Append(response.Plan.Set(ctx, &plan)...)
}

func (r *MSGraphResourceCollection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *MSGraphResourceCollectionModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := model.Timeouts.Create(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	newItems := AsListOfString(model.ReferenceIds)
	if err := r.syncCollection(ctx, model, nil, newItems); err != nil {
		resp.Diagnostics.AddError("Failed to sync collection", err.Error())
		return
	}

	model.Id = types.StringValue(baseCollectionUrl(model.Url.ValueString()))

	base := baseCollectionUrl(model.Url.ValueString())
	opts := clients.RequestOptions{
		QueryParameters: clients.NewQueryParameters(AsMapOfLists(model.ReadQueryParameters)),
		RetryOptions:    clients.NewRetryOptions(model.Retry),
	}
	body, err := r.client.List(ctx, base, model.ApiVersion.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read collection", err.Error())
		return
	}

	model.Output = types.DynamicValue(buildOutputFromBody(body, model.ResponseExportValues))
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *MSGraphResourceCollection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model, state *MSGraphResourceCollectionModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}
	if resp.Diagnostics.Append(req.State.Get(ctx, &state)...); resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := model.Timeouts.Update(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	newItems := AsListOfString(model.ReferenceIds)
	oldItems := AsListOfString(state.ReferenceIds)
	if err := r.syncCollection(ctx, model, oldItems, newItems); err != nil {
		resp.Diagnostics.AddError("Failed to sync collection", err.Error())
		return
	}

	base := baseCollectionUrl(model.Url.ValueString())
	opts := clients.RequestOptions{
		QueryParameters: clients.NewQueryParameters(AsMapOfLists(model.ReadQueryParameters)),
		RetryOptions:    clients.NewRetryOptions(model.Retry),
	}
	body, err := r.client.List(ctx, base, model.ApiVersion.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read collection", err.Error())
		return
	}

	model.Output = types.DynamicValue(buildOutputFromBody(body, model.ResponseExportValues))
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *MSGraphResourceCollection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *MSGraphResourceCollectionModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := model.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	base := baseCollectionUrl(model.Url.ValueString())
	opts := clients.RequestOptions{
		QueryParameters: clients.NewQueryParameters(AsMapOfLists(model.ReadQueryParameters)),
		RetryOptions:    clients.NewRetryOptions(model.Retry),
	}
	body, err := r.client.List(ctx, base, model.ApiVersion.ValueString(), opts)
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			tflog.Info(ctx, "Collection not found - removing from state")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read collection", err.Error())
		return
	}

	referenceIds, err := flattenReferenceIds(body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse collection", err.Error())
		return
	}
	model.ReferenceIds = referenceIds
	model.Output = types.DynamicValue(buildOutputFromBody(body, model.ResponseExportValues))
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *MSGraphResourceCollection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *MSGraphResourceCollectionModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := model.Timeouts.Delete(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	oldItems := AsListOfString(model.ReferenceIds)
	if err := r.syncCollection(ctx, model, oldItems, nil); err != nil {
		resp.Diagnostics.AddError("Failed to sync collection", err.Error())
		return
	}
}

func (r *MSGraphResourceCollection) syncCollection(ctx context.Context, model *MSGraphResourceCollectionModel, oldItems []string, newItems []string) error {
	toRemove := make([]string, 0)
	toAdd := make([]string, 0)
	oldSet := make(map[string]bool)
	for _, item := range oldItems {
		oldSet[item] = true
	}
	newSet := make(map[string]bool)
	for _, item := range newItems {
		newSet[item] = true
	}
	for _, item := range oldItems {
		if !newSet[item] {
			toRemove = append(toRemove, item)
		}
	}
	for _, item := range newItems {
		if !oldSet[item] {
			toAdd = append(toAdd, item)
		}
	}
	return r.applyCollection(ctx, model, toRemove, toAdd)
}

func (r *MSGraphResourceCollection) applyCollection(ctx context.Context, model *MSGraphResourceCollectionModel, toRemove []string, toAdd []string) error {
	errs := make([]error, 0)
	for _, item := range toAdd {
		body := map[string]string{}
		body["@odata.id"] = fmt.Sprintf("%s/%s/directoryObjects/%s", r.client.GraphBaseUrl(), model.ApiVersion.ValueString(), item)
		_, err := r.client.Create(ctx, model.Url.ValueString(), model.ApiVersion.ValueString(), body, clients.RequestOptions{RetryOptions: clients.NewRetryOptions(model.Retry)})
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, item := range toRemove {
		delUrl := fmt.Sprintf("%s/%s/$ref", baseCollectionUrl(model.Url.ValueString()), item)
		err := r.client.Delete(ctx, delUrl, model.ApiVersion.ValueString(), clients.RequestOptions{RetryOptions: clients.NewRetryOptions(model.Retry)})
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors during sync: %v", errs)
	}
	return nil
}

func flattenReferenceIds(body interface{}) (types.List, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return types.ListNull(types.StringType), err
	}
	type ListResponse struct {
		Values []struct {
			ID string `json:"id"`
		} `json:"value"`
	}
	var listResp ListResponse
	if err := json.Unmarshal(data, &listResp); err != nil {
		return types.ListNull(types.StringType), err
	}

	result := make([]attr.Value, 0, len(listResp.Values))
	for _, v := range listResp.Values {
		result = append(result, types.StringValue(v.ID))
	}
	return types.ListValueMust(types.StringType, result), nil
}

func baseCollectionUrl(url string) string { return strings.TrimSuffix(url, "/$ref") }
