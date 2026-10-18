package clients

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type paginationTransport func(*http.Request) (*http.Response, error)

func (f paginationTransport) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestMSGraphClientPagination(t *testing.T) {
	for _, method := range []string{"Read", "List"} {
		for _, emptyFirst := range []bool{false, true} {
			t.Run(method+"/emptyFirst="+map[bool]string{false: "false", true: "true"}[emptyFirst], func(t *testing.T) {
				calls := 0
				client, err := NewMSGraphClient(&testCredential{}, cloud.Configuration{}, &policy.ClientOptions{Transport: paginationTransport(func(req *http.Request) (*http.Response, error) {
					calls++
					if got := req.Header.Get("ConsistencyLevel"); got != "eventual" {
						t.Errorf("request %d ConsistencyLevel = %q", calls, got)
					}
					if got := req.Header.Get("Prefer"); got != "odata.maxpagesize=1" {
						t.Errorf("request %d Prefer = %q", calls, got)
					}
					if got := req.Header.Get("Accept"); got != "application/json" {
						t.Errorf("request %d Accept = %q", calls, got)
					}
					body := `{"value":[{"id":"second"}],"@odata.deltaLink":"https://graph.microsoft.com/v1.0/groups/delta?$deltatoken=opaque"}`
					if req.URL.Query().Get("$skiptoken") == "" {
						if req.URL.Query().Get("$count") != "true" {
							t.Error("initial query parameters missing")
						}
						values := `[{"id":"first"}]`
						if emptyFirst {
							values = `[]`
						}
						body = `{"@odata.context":"collection-context","@odata.count":2,"value":` + values + `,"@odata.nextLink":"https://graph.microsoft.com/v1.0/groups?$skiptoken=opaque%2Btoken"}`
					} else if req.URL.RawQuery != "$skiptoken=opaque%2Btoken" {
						t.Errorf("continuation query was modified: %s", req.URL.RawQuery)
					}
					return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
				})})
				if err != nil {
					t.Fatal(err)
				}
				opts := RequestOptions{Headers: map[string]string{"ConsistencyLevel": "eventual", "Prefer": "odata.maxpagesize=1"}, QueryParameters: map[string]string{"$count": "true"}}
				var result interface{}
				if method == "Read" {
					result, err = client.Read(context.Background(), "groups", "v1.0", opts)
				} else {
					result, err = client.List(context.Background(), "groups", "v1.0", opts)
				}
				if err != nil {
					t.Fatal(err)
				}
				wantValues := []interface{}{map[string]interface{}{"id": "second"}}
				if !emptyFirst {
					wantValues = append([]interface{}{map[string]interface{}{"id": "first"}}, wantValues...)
				}
				want := map[string]interface{}{"value": wantValues, "@odata.context": "collection-context", "@odata.count": float64(2), "@odata.deltaLink": "https://graph.microsoft.com/v1.0/groups/delta?$deltatoken=opaque"}
				if !reflect.DeepEqual(result, want) {
					t.Errorf("result = %#v, want %#v", result, want)
				}
				if calls != 2 {
					t.Errorf("requests = %d, want 2", calls)
				}
			})
		}
	}
}

func TestMSGraphClientPaginationRetryAndCancellation(t *testing.T) {
	for _, cancelAfterFirst := range []bool{false, true} {
		t.Run(map[bool]string{false: "retry", true: "cancel"}[cancelAfterFirst], func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				calls := 0
				client, err := NewMSGraphClient(&testCredential{}, cloud.Configuration{}, &policy.ClientOptions{Transport: paginationTransport(func(req *http.Request) (*http.Response, error) {
					calls++
					status, body := http.StatusOK, `{"value":[],"@odata.nextLink":"https://graph.microsoft.com/v1.0/groups?$skiptoken=next"}`
					switch {
					case cancelAfterFirst:
						cancel()
					case calls == 1:
						status = http.StatusServiceUnavailable
					case calls == 3:
						body = `{"value":[{"id":"last"}]}`
					}
					return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
				})})
				if err != nil {
					t.Fatal(err)
				}
				_, err = client.List(ctx, "groups", "v1.0", RequestOptions{RetryOptions: &policy.RetryOptions{MaxRetries: 1, RetryDelay: time.Millisecond, MaxRetryDelay: time.Millisecond}})
				if cancelAfterFirst {
					if err == nil || calls != 1 {
						t.Fatalf("cancel: calls=%d, error=%v", calls, err)
					}
				} else if err != nil || calls != 3 {
					t.Fatalf("retry: calls=%d, error=%v", calls, err)
				}
			})
		})
	}
}

func TestMSGraphClientReadSingleObject(t *testing.T) {
	client, err := NewMSGraphClient(&testCredential{}, cloud.Configuration{}, &policy.ClientOptions{Transport: &testTransport{}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.Read(context.Background(), "groups/1", "v1.0", RequestOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, map[string]interface{}{"id": "1"}) {
		t.Fatalf("object read = %#v", got)
	}
}
