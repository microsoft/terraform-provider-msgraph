package utils

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func TestResponseErrorWasNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "non-ResponseError",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "HTTP 404 status code",
			err: &azcore.ResponseError{
				StatusCode: http.StatusNotFound,
				ErrorCode:  "",
				RawResponse: &http.Response{
					StatusCode: http.StatusNotFound,
					Status:     "404 Not Found",
					Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
					Request: &http.Request{
						Method: "GET",
						URL:    &url.URL{Scheme: "https", Host: "graph.microsoft.com", Path: "/v1.0/users/test"},
					},
				},
			},
			expected: true,
		},
		{
			name: "HTTP 400 with ResourceNotFound error code",
			err: &azcore.ResponseError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "ResourceNotFound",
				RawResponse: &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":{"code":"ResourceNotFound","message":"Resource not found"}}`))),
					Request: &http.Request{
						Method: "GET",
						URL:    &url.URL{Scheme: "https", Host: "graph.microsoft.com", Path: "/beta/deviceManagement/configurationPolicies/test-id"},
					},
				},
			},
			expected: true,
		},
		{
			name: "HTTP 500 with different error code",
			err: &azcore.ResponseError{
				StatusCode: http.StatusInternalServerError,
				ErrorCode:  "InternalServerError",
				RawResponse: &http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
					Request: &http.Request{
						Method: "GET",
						URL:    &url.URL{Scheme: "https", Host: "graph.microsoft.com", Path: "/v1.0/users/test"},
					},
				},
			},
			expected: false,
		},
		{
			name: "HTTP 400 with different error code",
			err: &azcore.ResponseError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "BadRequest",
				RawResponse: &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
					Request: &http.Request{
						Method: "GET",
						URL:    &url.URL{Scheme: "https", Host: "graph.microsoft.com", Path: "/v1.0/users/test"},
					},
				},
			},
			expected: false,
		},
		{
			name: "HTTP 404 with ResourceNotFound error code (both conditions)",
			err: &azcore.ResponseError{
				StatusCode: http.StatusNotFound,
				ErrorCode:  "ResourceNotFound",
				RawResponse: &http.Response{
					StatusCode: http.StatusNotFound,
					Status:     "404 Not Found",
					Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
					Request: &http.Request{
						Method: "GET",
						URL:    &url.URL{Scheme: "https", Host: "graph.microsoft.com", Path: "/v1.0/users/test"},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResponseErrorWasNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("ResponseErrorWasNotFound() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestResponseErrorWasStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		expected   bool
	}{
		{
			name:       "nil error",
			err:        nil,
			statusCode: http.StatusNotFound,
			expected:   false,
		},
		{
			name:       "non-ResponseError",
			err:        errors.New("some error"),
			statusCode: http.StatusNotFound,
			expected:   false,
		},
		{
			name: "matching status code",
			err: &azcore.ResponseError{
				StatusCode: http.StatusBadRequest,
				RawResponse: &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
					Request: &http.Request{
						Method: "GET",
						URL:    &url.URL{Scheme: "https", Host: "graph.microsoft.com", Path: "/v1.0/users/test"},
					},
				},
			},
			statusCode: http.StatusBadRequest,
			expected:   true,
		},
		{
			name: "non-matching status code",
			err: &azcore.ResponseError{
				StatusCode: http.StatusBadRequest,
				RawResponse: &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
					Request: &http.Request{
						Method: "GET",
						URL:    &url.URL{Scheme: "https", Host: "graph.microsoft.com", Path: "/v1.0/users/test"},
					},
				},
			},
			statusCode: http.StatusNotFound,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResponseErrorWasStatusCode(tt.err, tt.statusCode)
			if result != tt.expected {
				t.Errorf("ResponseErrorWasStatusCode() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
