package clients

import (
	"context"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

const (
	moduleName    = "resource"
	moduleVersion = "v0.1.0"
	nextLinkKey   = "@odata.nextLink"
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

	// if response has nextLink, follow the link and return the final response
	if responseBodyMap, ok := responseBody.(map[string]interface{}); ok {
		if nextLink := responseBodyMap["@odata.nextLink"]; nextLink != nil {
			return client.List(ctx, url, apiVersion, options)
		}
	}

	return responseBody, nil
}

func (client *MSGraphClient) List(ctx context.Context, url string, apiVersion string, options RequestOptions) (interface{}, error) {
	pager := runtime.NewPager(runtime.PagingHandler[interface{}]{
		More: func(current interface{}) bool {
			if current == nil {
				return false
			}
			currentMap, ok := current.(map[string]interface{})
			if !ok {
				return false
			}
			if currentMap[nextLinkKey] == nil {
				return false
			}
			if nextLink := currentMap[nextLinkKey].(string); nextLink == "" {
				return false
			}
			return true
		},
		Fetcher: func(ctx context.Context, current *interface{}) (interface{}, error) {
			var request *policy.Request
			if current == nil {
				req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(client.host, apiVersion, url))
				if err != nil {
					return nil, err
				}
				reqQP := req.Raw().URL.Query()
				for key, value := range options.QueryParameters {
					reqQP.Set(key, value)
				}
				req.Raw().URL.RawQuery = reqQP.Encode()
				for key, value := range options.Headers {
					req.Raw().Header.Set(key, value)
				}
				request = req
			} else {
				nextLink := ""
				if currentMap, ok := (*current).(map[string]interface{}); ok && currentMap[nextLinkKey] != nil {
					nextLink = currentMap[nextLinkKey].(string)
				}
				req, err := runtime.NewRequest(ctx, http.MethodGet, nextLink)
				if err != nil {
					return nil, err
				}
				request = req
			}
			request.Raw().Header.Set("Accept", "application/json")
			resp, err := client.pl.Do(request)
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
		},
	})

	out := make(map[string]interface{})
	value := make([]interface{}, 0)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		if pageMap, ok := page.(map[string]interface{}); ok {
			if pageMap["value"] != nil {
				if pageValue, ok := pageMap["value"].([]interface{}); ok {
					value = append(value, pageValue...)
					continue
				}
			}
			// copy all fields except for nextLinkKey and value
			for key, val := range pageMap {
				if key != nextLinkKey && key != "value" {
					out[key] = val
				}
			}
		}

		// if response doesn't follow the paging guideline, return the response as is
		return page, nil
	}

	out["value"] = value

	return out, nil
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

	// TODO: Handle long-running operations if needed

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

	// TODO: Handle long-running operations if needed

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

	// TODO: Handle long-running operations if needed

	if !runtime.HasStatusCode(resp, http.StatusNoContent) {
		return runtime.NewResponseError(resp)
	}
	return nil
}
