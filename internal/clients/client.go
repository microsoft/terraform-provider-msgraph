package clients

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"log"
	"time"
)

type Client struct {
	// StopContext is used for propagating control from Terraform Core (e.g. Ctrl/Cmd+C)
	StopContext context.Context

	MSGraphClient *MSGraphClient

	Option *Option
}

type Option struct {
	Cred                        azcore.TokenCredential
	ApplicationUserAgent        string
	DisableCorrelationRequestID bool
	CloudCfg                    cloud.Configuration
	CustomCorrelationRequestID  string
	SubscriptionId              string
	TenantId                    string
}

func (client *Client) Build(ctx context.Context, o *Option) error {
	client.StopContext = ctx
	client.Option = o

	azlog.SetListener(func(cls azlog.Event, msg string) {
		log.Printf("[DEBUG] %s %s: %s\n", time.Now().Format(time.StampMicro), cls, msg)
	})

	perCallPolicies := make([]policy.Policy, 0)
	perCallPolicies = append(perCallPolicies, withUserAgent(o.ApplicationUserAgent))
	if !o.DisableCorrelationRequestID {
		id := o.CustomCorrelationRequestID
		if id == "" {
			id = correlationRequestID()
		}
		perCallPolicies = append(perCallPolicies, withCorrelationRequestID(id))
	}
	perRetryPolicies := make([]policy.Policy, 0)
	perRetryPolicies = append(perRetryPolicies, NewLiveTrafficLogPolicy())

	allowedHeaders := []string{
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Origin",
		"Elapsed-Time",
		"Location",
		"Metadata",
		"Ocp-Automation-Accountid",
		"P3p",
		"Strict-Transport-Security",
		"Vary",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-Ms-Correlation-Request-Id",
		"X-Ms-Ests-Server",
		"X-Ms-Failure-Cause",
		"X-Ms-Ratelimit-Remaining-Subscription-Reads",
		"X-Ms-Ratelimit-Remaining-Subscription-Writes",
		"X-Ms-Ratelimit-Remaining-Tenant-Reads",
		"X-Ms-Ratelimit-Remaining-Tenant-Writes",
		"X-Ms-Request-Id",
		"X-Ms-Routing-Request-Id",
		"X-Xss-Protection",
	}
	allowedQueryParams := []string{
		"api-version",
		"$skipToken",
	}

	msgraphClient, err := NewMSGraphClient(o.Cred, &policy.ClientOptions{
		Logging: policy.LogOptions{
			IncludeBody:        false,
			AllowedHeaders:     allowedHeaders,
			AllowedQueryParams: allowedQueryParams,
		},
		PerCallPolicies:  perCallPolicies,
		PerRetryPolicies: perRetryPolicies,
	})
	if err != nil {
		return err
	}

	client.MSGraphClient = msgraphClient

	return nil
}
