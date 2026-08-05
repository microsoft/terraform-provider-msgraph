package myplanmodifier

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OrderInsensitiveStringList returns a plan modifier that ignores order-only changes
// between state and planned values for a list of primitive values (strings, numbers).
// If the multiset of items is identical, the plan value is reverted to the state value
// to avoid a diff.
type orderInsensitiveStringList struct{}

// OrderInsensitiveStringList creates a new order-insensitive list plan modifier for string lists.
func OrderInsensitiveStringList() planmodifier.List { return orderInsensitiveStringList{} }

func (m orderInsensitiveStringList) Description(ctx context.Context) string {
	return "Ignore order-only changes"
}

func (m orderInsensitiveStringList) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m orderInsensitiveStringList) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() || req.StateValue.Equal(req.PlanValue) {
		return
	}

	// Only handle simple element types we can compare deterministically.
	// only for strings
	if req.PlanValue.ElementType(ctx) != types.StringType {
		return
	}

	var stateItems, planItems []string
	if diags := req.StateValue.ElementsAs(ctx, &stateItems, false); diags.HasError() {
		return
	}
	if diags := req.PlanValue.ElementsAs(ctx, &planItems, false); diags.HasError() {
		return
	}
	if len(stateItems) != len(planItems) {
		return
	}

	sa := append([]string{}, stateItems...)
	sb := append([]string{}, planItems...)
	sort.Strings(sa)
	sort.Strings(sb)
	for i := range sa {
		if sa[i] != sb[i] {
			return
		}
	}
	resp.PlanValue = req.StateValue
}
