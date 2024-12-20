package clients

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"net/http"
)

const (
	moduleName    = "resource"
	moduleVersion = "v0.1.0"
)

type MSGraphClient struct {
	host string
	pl   runtime.Pipeline
}

func NewMSGraphClient(credential azcore.TokenCredential, opt *policy.ClientOptions) (*MSGraphClient, error) {
	pl := runtime.NewPipeline(moduleName, moduleVersion, runtime.PipelineOptions{
		AllowedHeaders:         nil,
		AllowedQueryParameters: nil,
		APIVersion:             runtime.APIVersionOptions{},
		PerCall:                nil,
		PerRetry: []policy.Policy{
			runtime.NewBearerTokenPolicy(credential, []string{"https://graph.microsoft.com/.default"}, nil),
		},
		Tracing: runtime.TracingOptions{},
	}, opt)
	return &MSGraphClient{
		host: "https://graph.microsoft.com",
		pl:   pl,
	}, nil
}

func (client *MSGraphClient) Read(ctx context.Context, url string, apiVersion string, options RequestOptions) (interface{}, error) {
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(client.host, apiVersion, url))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	for key, value := range options.QueryParameters {
		reqQP.Set(key, value)
	}
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	for key, value := range options.Headers {
		req.Raw().Header.Set(key, value)
	}
	resp, err := client.pl.Do(req)
	if err != nil {
		return nil, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK) {
		return nil, runtime.NewResponseError(resp)
	}

	var responseBody interface{}
	if err := runtime.UnmarshalAsJSON(resp, &responseBody); err != nil {
		return nil, err
	}
	return responseBody, nil
}

func (client *MSGraphClient) Create(ctx context.Context, url string, apiVersion string, body interface{}, options RequestOptions) (interface{}, error) {
	req, err := runtime.NewRequest(ctx, http.MethodPost, runtime.JoinPaths(client.host, apiVersion, url))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	for key, value := range options.QueryParameters {
		reqQP.Set(key, value)
	}
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	for key, value := range options.Headers {
		req.Raw().Header.Set(key, value)
	}
	if err := runtime.MarshalAsJSON(req, body); err != nil {
		return nil, err
	}
	resp, err := client.pl.Do(req)
	if err != nil {
		return nil, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent) {
		return nil, runtime.NewResponseError(resp)
	}

	//pt, err := runtime.NewPoller[interface{}](resp, client.pl, nil)
	//if err == nil {
	//	resp, err := pt.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
	//		Frequency: 10 * time.Second,
	//	})
	//	if err == nil {
	//		return resp, nil
	//	}
	//	return nil, err
	//}

	var responseBody interface{}
	if err := runtime.UnmarshalAsJSON(resp, &responseBody); err != nil {
		return nil, err
	}
	return responseBody, nil
}

func (client *MSGraphClient) Update(ctx context.Context, url string, apiVersion string, body interface{}, options RequestOptions) (interface{}, error) {
	req, err := runtime.NewRequest(ctx, http.MethodPatch, runtime.JoinPaths(client.host, apiVersion, url))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	for key, value := range options.QueryParameters {
		reqQP.Set(key, value)
	}
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	for key, value := range options.Headers {
		req.Raw().Header.Set(key, value)
	}
	if err := runtime.MarshalAsJSON(req, body); err != nil {
		return nil, err
	}
	resp, err := client.pl.Do(req)
	if err != nil {
		return nil, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusAccepted, http.StatusNoContent) {
		return nil, runtime.NewResponseError(resp)
	}

	//pt, err := runtime.NewPoller[interface{}](resp, client.pl, nil)
	//if err == nil {
	//	resp, err := pt.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
	//		Frequency: 10 * time.Second,
	//	})
	//	if err == nil {
	//		return resp, nil
	//	}
	//	return nil, err
	//}

	var responseBody interface{}
	if err := runtime.UnmarshalAsJSON(resp, &responseBody); err != nil {
		return nil, err
	}
	return responseBody, nil
}

func (client *MSGraphClient) Delete(ctx context.Context, url string, apiVersion string, options RequestOptions) error {
	req, err := runtime.NewRequest(ctx, http.MethodDelete, runtime.JoinPaths(client.host, apiVersion, url))
	if err != nil {
		return err
	}
	reqQP := req.Raw().URL.Query()
	for key, value := range options.QueryParameters {
		reqQP.Set(key, value)
	}
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	for key, value := range options.Headers {
		req.Raw().Header.Set(key, value)
	}
	resp, err := client.pl.Do(req)
	if err != nil {
		return err
	}

	//pt, err := runtime.NewPoller[interface{}](resp, client.pl, nil)
	//if err == nil {
	//	_, err := pt.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
	//		Frequency: 10 * time.Second,
	//	})
	//	return err
	//}

	if !runtime.HasStatusCode(resp, http.StatusNoContent) {
		return runtime.NewResponseError(resp)
	}
	return nil
}
