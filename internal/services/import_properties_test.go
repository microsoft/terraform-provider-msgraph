package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAddNullProperty(t *testing.T) {
	testCases := []struct {
		name       string
		properties []string
		expected   string
	}{
		{
			name:       "single property",
			properties: []string{"displayName"},
			expected:   `{"displayName":null}`,
		},
		{
			name:       "multiple properties",
			properties: []string{"displayName", "mailEnabled"},
			expected:   `{"displayName":null,"mailEnabled":null}`,
		},
		{
			name:       "nested property",
			properties: []string{"web.redirectUris"},
			expected:   `{"web":{"redirectUris":null}}`,
		},
		{
			name:       "deeply nested property",
			properties: []string{"a.b.c.d"},
			expected:   `{"a":{"b":{"c":{"d":null}}}}`,
		},
		{
			name:       "siblings under a shared parent",
			properties: []string{"web.redirectUris", "web.homePageUrl"},
			expected:   `{"web":{"homePageUrl":null,"redirectUris":null}}`,
		},
		{
			name:       "duplicate property is idempotent",
			properties: []string{"displayName", "displayName"},
			expected:   `{"displayName":null}`,
		},
		// The parent/child collision cases must agree: the more specific path
		// wins whichever order the properties were given in.
		{
			name:       "parent before child",
			properties: []string{"web", "web.redirectUris"},
			expected:   `{"web":{"redirectUris":null}}`,
		},
		{
			name:       "child before parent",
			properties: []string{"web.redirectUris", "web"},
			expected:   `{"web":{"redirectUris":null}}`,
		},
		{
			name:       "parent between two children",
			properties: []string{"web.redirectUris", "web", "web.homePageUrl"},
			expected:   `{"web":{"homePageUrl":null,"redirectUris":null}}`,
		},
		// Malformed paths are skipped without leaving partial objects behind.
		{
			name:       "empty segment",
			properties: []string{"a..b"},
			expected:   `{}`,
		},
		{
			name:       "trailing dot",
			properties: []string{"a."},
			expected:   `{}`,
		},
		{
			name:       "leading dot",
			properties: []string{".a"},
			expected:   `{}`,
		},
		{
			name:       "malformed path does not discard valid ones",
			properties: []string{"displayName", "a..b"},
			expected:   `{"displayName":null}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			seed := make(map[string]interface{})
			for _, property := range testCase.properties {
				addNullProperty(seed, property)
			}

			actual, err := json.Marshal(seed)
			if err != nil {
				t.Fatalf("failed to marshal seed: %s", err)
			}
			if string(actual) != testCase.expected {
				t.Errorf("expected %s, got %s", testCase.expected, actual)
			}
		})
	}
}

func TestImportStateBody(t *testing.T) {
	testCases := []struct {
		name         string
		importID     string
		expectedBody string
		expectedURL  string
		expectedID   string
		expectedAPI  string
	}{
		{
			name:         "without importProperties the body stays null",
			importID:     "/groups/00000000-0000-0000-0000-000000000000",
			expectedBody: "<null>",
			expectedURL:  "groups",
			expectedID:   "00000000-0000-0000-0000-000000000000",
			expectedAPI:  "v1.0",
		},
		{
			name:         "single property",
			importID:     "/groups/00000000-0000-0000-0000-000000000000?importProperties=displayName",
			expectedBody: `{"displayName":<null>}`,
			expectedURL:  "groups",
			expectedID:   "00000000-0000-0000-0000-000000000000",
			expectedAPI:  "v1.0",
		},
		{
			name:         "several properties including a nested one",
			importID:     "/applications/app-id?importProperties=displayName,web.redirectUris",
			expectedBody: `{"displayName":<null>,"web":{"redirectUris":<null>}}`,
			expectedURL:  "applications",
			expectedID:   "app-id",
			expectedAPI:  "v1.0",
		},
		{
			name:         "surrounding whitespace is trimmed",
			importID:     "/groups/group-id?importProperties=displayName,%20mailEnabled",
			expectedBody: `{"displayName":<null>,"mailEnabled":<null>}`,
			expectedURL:  "groups",
			expectedID:   "group-id",
			expectedAPI:  "v1.0",
		},
		{
			name:         "empty importProperties leaves the body null",
			importID:     "/groups/group-id?importProperties=",
			expectedBody: "<null>",
			expectedURL:  "groups",
			expectedID:   "group-id",
			expectedAPI:  "v1.0",
		},
		{
			name:         "only separators leaves the body null",
			importID:     "/groups/group-id?importProperties=,,",
			expectedBody: "<null>",
			expectedURL:  "groups",
			expectedID:   "group-id",
			expectedAPI:  "v1.0",
		},
		{
			name:         "combined with api-version",
			importID:     "/groups/group-id?api-version=beta&importProperties=displayName",
			expectedBody: `{"displayName":<null>}`,
			expectedURL:  "groups",
			expectedID:   "group-id",
			expectedAPI:  "beta",
		},
		{
			name:         "$ref import",
			importID:     "/groups/group-id/members/member-id/$ref?importProperties=displayName",
			expectedBody: `{"displayName":<null>}`,
			expectedURL:  "groups/group-id/members/$ref",
			expectedID:   "member-id",
			expectedAPI:  "v1.0",
		},
	}

	ctx := context.Background()
	r := &MSGraphResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", schemaResp.Diagnostics)
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resp := &resource.ImportStateResponse{
				State: tfsdk.State{
					Schema: schemaResp.Schema,
					Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
				},
			}

			r.ImportState(ctx, resource.ImportStateRequest{ID: testCase.importID}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}

			var model MSGraphResourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read imported state: %v", diags)
			}

			if actual := model.Body.String(); actual != testCase.expectedBody {
				t.Errorf("body: expected %s, got %s", testCase.expectedBody, actual)
			}
			if actual := model.Url.ValueString(); actual != testCase.expectedURL {
				t.Errorf("url: expected %s, got %s", testCase.expectedURL, actual)
			}
			if actual := model.Id.ValueString(); actual != testCase.expectedID {
				t.Errorf("id: expected %s, got %s", testCase.expectedID, actual)
			}
			if actual := model.ApiVersion.ValueString(); actual != testCase.expectedAPI {
				t.Errorf("api_version: expected %s, got %s", testCase.expectedAPI, actual)
			}
		})
	}
}
