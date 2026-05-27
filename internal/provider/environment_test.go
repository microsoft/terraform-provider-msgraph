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
		wantName      string
		wantAuthority string
		wantGraphHost string
		wantScope     string
	}{
		{
			name:          "empty defaults to public",
			input:         "",
			wantName:      EnvironmentPublic,
			wantAuthority: cloud.AzurePublic.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostPublic,
			wantScope:     GraphHostPublic + "/.default",
		},
		{
			name:          "public",
			input:         "public",
			wantName:      EnvironmentPublic,
			wantAuthority: cloud.AzurePublic.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostPublic,
			wantScope:     GraphHostPublic + "/.default",
		},
		{
			name:          "china case insensitive",
			input:         "China",
			wantName:      EnvironmentChina,
			wantAuthority: cloud.AzureChina.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostChina,
			wantScope:     GraphHostChina + "/.default",
		},
		{
			name:          "us government",
			input:         "usgovernment",
			wantName:      EnvironmentUSGovernment,
			wantAuthority: cloud.AzureGovernment.ActiveDirectoryAuthorityHost,
			wantGraphHost: GraphHostUSGovernment,
			wantScope:     GraphHostUSGovernment + "/.default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveCloudEnvironment(tt.input)
			if err != nil {
				t.Fatalf("ResolveCloudEnvironment() returned error: %v", err)
			}
			if got.Name != tt.wantName {
				t.Fatalf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Cloud.ActiveDirectoryAuthorityHost != tt.wantAuthority {
				t.Fatalf("ActiveDirectoryAuthorityHost = %q, want %q", got.Cloud.ActiveDirectoryAuthorityHost, tt.wantAuthority)
			}
			graphCfg, ok := got.Cloud.Services[clients.MicrosoftGraph]
			if !ok {
				t.Fatal("missing Microsoft Graph service configuration")
			}
			if graphCfg.Endpoint != tt.wantGraphHost {
				t.Fatalf("Graph endpoint = %q, want %q", graphCfg.Endpoint, tt.wantGraphHost)
			}
			if graphCfg.Audience+"/.default" != tt.wantScope {
				t.Fatalf("Graph scope = %q, want %q", graphCfg.Audience+"/.default", tt.wantScope)
			}
		})
	}
}

func TestResolveCloudEnvironmentInvalid(t *testing.T) {
	_, err := ResolveCloudEnvironment("moon")
	if err == nil {
		t.Fatal("expected error for unsupported environment")
	}
}
