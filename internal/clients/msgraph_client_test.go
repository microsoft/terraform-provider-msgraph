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
	credential := &testCredential{}
	transport := &testTransport{}
	cloudCfg := cloud.Configuration{
		Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
			MicrosoftGraph: {
				Endpoint: "https://microsoftgraph.chinacloudapi.cn",
				Audience: "https://microsoftgraph.chinacloudapi.cn",
			},
		},
	}

	client, err := NewMSGraphClient(credential, cloudCfg, &policy.ClientOptions{Transport: transport})
	if err != nil {
		t.Fatalf("NewMSGraphClient() returned error: %v", err)
	}

	_, err = client.Read(context.Background(), "users", "v1.0", RequestOptions{})
	if err != nil {
		t.Fatalf("Read() returned error: %v", err)
	}

	if transport.url != "https://microsoftgraph.chinacloudapi.cn/v1.0/users" {
		t.Fatalf("request URL = %q", transport.url)
	}
	if len(credential.scopes) != 1 || credential.scopes[0] != "https://microsoftgraph.chinacloudapi.cn/.default" {
		t.Fatalf("scopes = %#v", credential.scopes)
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
