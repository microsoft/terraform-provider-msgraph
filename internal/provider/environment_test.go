package provider

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
)

func TestResolveCloudEnvironment(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantAuthority string
		wantGraphHost string
	}{
		{
			name:          "empty defaults to public",
			input:         "",
			wantAuthority: cloud.AzurePublic.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostPublic,
		},
		{
			name:          "public",
			input:         "public",
			wantAuthority: cloud.AzurePublic.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostPublic,
		},
		{
			name:          "global alias maps to public",
			input:         "global",
			wantAuthority: cloud.AzurePublic.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostPublic,
		},
		{
			name:          "china case insensitive",
			input:         "China",
			wantAuthority: cloud.AzureChina.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostChina,
		},
		{
			name:          "us government",
			input:         "usgovernment",
			wantAuthority: cloud.AzureGovernment.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostUSGovernmentL4,
		},
		{
			name:          "us government l4 alias",
			input:         "usgovernmentl4",
			wantAuthority: cloud.AzureGovernment.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostUSGovernmentL4,
		},
		{
			name:          "us government l5",
			input:         "usgovernmentl5",
			wantAuthority: cloud.AzureGovernment.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostUSGovernmentL5,
		},
		{
			name:          "dod alias maps to l5",
			input:         "dod",
			wantAuthority: cloud.AzureGovernment.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostUSGovernmentL5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveCloudEnvironment(tt.input)
			if err != nil {
				t.Fatalf("ResolveCloudEnvironment() returned error: %v", err)
			}
			if got.ActiveDirectoryAuthorityHost != tt.wantAuthority {
				t.Fatalf("ActiveDirectoryAuthorityHost = %q, want %q", got.ActiveDirectoryAuthorityHost, tt.wantAuthority)
			}
			graphCfg, ok := got.Services[clients.MicrosoftGraph]
			if !ok {
				t.Fatal("missing Microsoft Graph service configuration")
			}
			if graphCfg.Endpoint != tt.wantGraphHost {
				t.Fatalf("Graph endpoint = %q, want %q", graphCfg.Endpoint, tt.wantGraphHost)
			}
			if graphCfg.Audience != tt.wantGraphHost {
				t.Fatalf("Graph audience = %q, want %q", graphCfg.Audience, tt.wantGraphHost)
			}
		})
	}
}

func TestResolveCloudEnvironmentInvalid(t *testing.T) {
	if _, err := ResolveCloudEnvironment("moon"); err == nil {
		t.Fatal("expected error for unsupported environment")
	}
}
