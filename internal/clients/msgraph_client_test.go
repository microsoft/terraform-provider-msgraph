package clients

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type testCredential struct {
	scopes []string
}

func (c *testCredential) GetToken(_ context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	c.scopes = append([]string{}, opts.Scopes...)
	return azcore.AccessToken{Token: "token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type testTransport struct {
	url string
}

func (t *testTransport) Do(req *http.Request) (*http.Response, error) {
	t.url = req.URL.String()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"id":"1"}`)),
		Request:    req,
	}, nil
}

func TestNewMSGraphClientUsesConfiguredGraphEndpointAndScope(t *testing.T) {
	// Verifies that for each sovereign cloud, requests are sent to the correct
	// Graph endpoint and tokens are requested with the matching `.default` scope,
	// without requiring real credentials in those clouds.
	tests := []struct {
		name      string
		graphHost string
		wantURL   string
		wantScope string
	}{
		{
			name:      "china",
			graphHost: "https://microsoftgraph.chinacloudapi.cn",
			wantURL:   "https://microsoftgraph.chinacloudapi.cn/v1.0/users",
			wantScope: "https://microsoftgraph.chinacloudapi.cn/.default",
		},
		{
			name:      "us government l4",
			graphHost: "https://graph.microsoft.us",
			wantURL:   "https://graph.microsoft.us/v1.0/users",
			wantScope: "https://graph.microsoft.us/.default",
		},
		{
			name:      "us government l5",
			graphHost: "https://dod-graph.microsoft.us",
			wantURL:   "https://dod-graph.microsoft.us/v1.0/users",
			wantScope: "https://dod-graph.microsoft.us/.default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credential := &testCredential{}
			transport := &testTransport{}
			cloudCfg := cloud.Configuration{
				Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
					MicrosoftGraph: {
						Endpoint: tt.graphHost,
						Audience: tt.graphHost,
					},
				},
			}

			client, err := NewMSGraphClient(credential, cloudCfg, &policy.ClientOptions{Transport: transport})
			if err != nil {
				t.Fatalf("NewMSGraphClient() returned error: %v", err)
			}

			if _, err := client.Read(context.Background(), "users", "v1.0", RequestOptions{}); err != nil {
				t.Fatalf("Read() returned error: %v", err)
			}

			if transport.url != tt.wantURL {
				t.Fatalf("request URL = %q, want %q", transport.url, tt.wantURL)
			}
			if len(credential.scopes) != 1 || credential.scopes[0] != tt.wantScope {
				t.Fatalf("scopes = %#v, want [%q]", credential.scopes, tt.wantScope)
			}
		})
	}
}

func TestNewMSGraphClientDefaultsToPublicGraph(t *testing.T) {
	client, err := NewMSGraphClient(&testCredential{}, cloud.Configuration{}, nil)
	if err != nil {
		t.Fatalf("NewMSGraphClient() returned error: %v", err)
	}

	if client.GraphBaseUrl() != defaultGraphHost {
		t.Fatalf("GraphBaseUrl() = %q, want %q", client.GraphBaseUrl(), defaultGraphHost)
	}
}
