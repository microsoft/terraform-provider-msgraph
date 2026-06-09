package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

const (
	moduleName    = "resource"
	moduleVersion = "v0.1.0"
	nextLinkKey   = "@odata.nextLink"
	listCacheTTL  = 10 * time.Second
)

type listCacheEntry struct {
	body      interface{}
	fetchedAt time.Time
}

type listInflight struct {
	done chan struct{}
	body interface{}
	err  error
}

type MSGraphClient struct {
	host      string
	pl        runtime.Pipeline
	cacheMu   sync.Mutex
	listCache map[string]listCacheEntry
	inFlight  map[string]*listInflight
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
		host:      "https://graph.microsoft.com",
		pl:        pl,
		listCache: make(map[string]listCacheEntry),
		inFlight:  make(map[string]*listInflight),
	}, nil
}

func (client *MSGraphClient) Read(ctx context.Context, url string, apiVersion string, options RequestOptions) (interface{}, error) {
	if options.RetryOptions != nil {
		ctx = policy.WithRetryOptions(ctx, *options.RetryOptions)
	}
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

	if responseBodyMap, ok := responseBody.(map[string]interface{}); ok {
		if nextLink := responseBodyMap["@odata.nextLink"]; nextLink != nil {
			return client.List(ctx, url, apiVersion, options)
		}
	}

	return responseBody, nil
}

func (client *MSGraphClient) cachedList(ctx context.Context, url, apiVersion string, options RequestOptions) (interface{}, error) {
	key := listCacheKey(url, apiVersion, options.QueryParameters)

	client.cacheMu.Lock()
	if entry, ok := client.listCache[key]; ok && time.Since(entry.fetchedAt) < listCacheTTL {
		client.cacheMu.Unlock()
		return entry.body, nil
	}
	if f, ok := client.inFlight[key]; ok {
		client.cacheMu.Unlock()
		select {
		case <-f.done:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return f.body, f.err
	}
	f := &listInflight{done: make(chan struct{})}
	client.inFlight[key] = f
	client.cacheMu.Unlock()

	body, err := client.List(ctx, url, apiVersion, options)

	client.cacheMu.Lock()
	delete(client.inFlight, key)
	if err == nil {
		client.listCache[key] = listCacheEntry{body: body, fetchedAt: time.Now()}
	}
	client.cacheMu.Unlock()

	f.body = body
	f.err = err
	close(f.done)

	return body, err
}

func listCacheKey(url, apiVersion string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	key := fmt.Sprintf("%s|%s", apiVersion, url)
	for _, k := range keys {
		key += fmt.Sprintf("|%s=%s", k, params[k])
	}
	return key
}

func (client *MSGraphClient) ListRefIDs(ctx context.Context, url string, apiVersion string, options RequestOptions) ([]string, error) {
	responseBody, err := client.cachedList(ctx, url, apiVersion, options)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(responseBody)
	if err != nil {
		return nil, err
	}
	type ListResponse struct {
		Values []struct {
			ID string `json:"id"`
		} `json:"value"`
	}
	var listResp ListResponse
	if err := json.Unmarshal(data, &listResp); err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, v := range listResp.Values {
		result = append(result, v.ID)
	}
	return result, nil
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
			if options.RetryOptions != nil {
				ctx = policy.WithRetryOptions(ctx, *options.RetryOptions)
			}
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
			for key, val := range pageMap {
				if key != nextLinkKey && key != "value" {
					out[key] = val
				}
			}
		}

		return page, nil
	}

	out["value"] = value

	return out, nil
}

func (client *MSGraphClient) invalidateListCache(collectionUrl, apiVersion string) {
	prefix := fmt.Sprintf("%s|%s", apiVersion, collectionUrl)
	client.cacheMu.Lock()
	defer client.cacheMu.Unlock()
	for key := range client.listCache {
		if strings.HasPrefix(key, prefix) {
			delete(client.listCache, key)
		}
	}
}

func parentCollectionUrl(itemUrl string) string {
	u := strings.TrimSuffix(itemUrl, "/$ref")
	if idx := strings.LastIndex(u, "/"); idx >= 0 {
		return u[:idx]
	}
	return u
}

func (client *MSGraphClient) ReadFromList(ctx context.Context, collectionUrl string, id string, apiVersion string, options RequestOptions) (interface{}, error) {
	responseBody, err := client.cachedList(ctx, collectionUrl, apiVersion, options)
	if err != nil {
		return nil, err
	}
	return findItemInList(responseBody, id)
}

func (client *MSGraphClient) ReadFromListWithWait(ctx context.Context, collectionUrl, id, apiVersion string, options RequestOptions) (interface{}, error) {
	for {
		body, err := client.cachedList(ctx, collectionUrl, apiVersion, options)
		if err != nil {
			return nil, err
		}
		if item, err := findItemInList(body, id); err == nil {
			return item, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for resource %q to appear in collection %s", id, collectionUrl)
		case <-time.After(listCacheTTL):
		}
	}
}

func findItemInList(body interface{}, id string) (interface{}, error) {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		return nil, &azcore.ResponseError{StatusCode: http.StatusNotFound}
	}
	values, ok := bodyMap["value"].([]interface{})
	if !ok {
		return nil, &azcore.ResponseError{StatusCode: http.StatusNotFound}
	}
	for _, item := range values {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if itemId, ok := itemMap["id"].(string); ok && itemId == id {
			return item, nil
		}
	}
	return nil, &azcore.ResponseError{StatusCode: http.StatusNotFound}
}

func (client *MSGraphClient) Create(ctx context.Context, url string, apiVersion string, body interface{}, options RequestOptions) (interface{}, error) {
	if options.RetryOptions != nil {
		ctx = policy.WithRetryOptions(ctx, *options.RetryOptions)
	}
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

	var responseBody interface{}
	if err := runtime.UnmarshalAsJSON(resp, &responseBody); err != nil {
		return nil, err
	}
	client.invalidateListCache(url, apiVersion)
	return responseBody, nil
}

func (client *MSGraphClient) Update(ctx context.Context, url string, apiVersion string, body interface{}, options RequestOptions) (interface{}, error) {
	if options.RetryOptions != nil {
		ctx = policy.WithRetryOptions(ctx, *options.RetryOptions)
	}
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

	var responseBody interface{}
	if err := runtime.UnmarshalAsJSON(resp, &responseBody); err != nil {
		return nil, err
	}
	client.invalidateListCache(parentCollectionUrl(url), apiVersion)
	return responseBody, nil
}

func (client *MSGraphClient) Delete(ctx context.Context, url string, apiVersion string, options RequestOptions) error {
	if options.RetryOptions != nil {
		ctx = policy.WithRetryOptions(ctx, *options.RetryOptions)
	}
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

	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusAccepted, http.StatusNoContent) {
		return runtime.NewResponseError(resp)
	}
	client.invalidateListCache(parentCollectionUrl(url), apiVersion)
	return nil
}

func (client *MSGraphClient) Action(ctx context.Context, method string, url string, apiVersion string, body interface{}, options RequestOptions) (interface{}, error) {
	if options.RetryOptions != nil {
		ctx = policy.WithRetryOptions(ctx, *options.RetryOptions)
	}

	req, err := runtime.NewRequest(ctx, method, runtime.JoinPaths(client.host, apiVersion, url))
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

	if body != nil {
		if err := runtime.MarshalAsJSON(req, body); err != nil {
			return nil, err
		}
		req.Raw().Header.Set("Content-Type", "application/json")
	}

	resp, err := client.pl.Do(req)
	if err != nil {
		return nil, err
	}

	// Check for successful status codes (2xx range)
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent) {
		return nil, runtime.NewResponseError(resp)
	}

	// Invalidate the list cache for the parent collection after any mutating action.
	client.invalidateListCache(parentCollectionUrl(url), apiVersion)

	// For methods that typically don't return a body (like DELETE), or if response is empty
	if resp.StatusCode == http.StatusNoContent || resp.ContentLength == 0 {
		return nil, nil
	}

	var responseBody interface{}
	if err := runtime.UnmarshalAsJSON(resp, &responseBody); err != nil {
		return nil, err
	}

	return responseBody, nil
}

func (client *MSGraphClient) GraphBaseUrl() string {
	return client.host
}
