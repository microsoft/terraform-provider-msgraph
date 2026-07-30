package clients

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func responseWithBody(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestNewRetryOptionsForReadAfterCreate_ShouldRetry(t *testing.T) {
	const replicationBody = `{
	  "error": {
	    "code": "Request_ResourceNotFound",
	    "message": "Resource 'fc64da37-6c49-4363-9af1-98aa6472fcd1' does not exist or one of its queried reference-property objects are not present."
	  }
	}`

	testCases := []struct {
		name       string
		statusCode int
		body       string
		expected   bool
	}{
		{
			name:       "404 is retried while the resource replicates",
			statusCode: http.StatusNotFound,
			body:       replicationBody,
			expected:   true,
		},
		{
			name:       "403 is retried while the resource replicates",
			statusCode: http.StatusForbidden,
			body:       `{"error":{"code":"Authorization_RequestDenied"}}`,
			expected:   true,
		},
		{
			name:       "429 is retried as a transient failure",
			statusCode: http.StatusTooManyRequests,
			body:       `{}`,
			expected:   true,
		},
		{
			name:       "200 is not retried",
			statusCode: http.StatusOK,
			body:       `{}`,
			expected:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			options := NewRetryOptionsForReadAfterCreate()

			actual := options.ShouldRetry(responseWithBody(testCase.statusCode, testCase.body), nil)
			if actual != testCase.expected {
				t.Fatalf("expected ShouldRetry to return %t but got %t", testCase.expected, actual)
			}
		})
	}
}

func TestNewRetryOptionsForReadAfterCreate_NilResponse(t *testing.T) {
	options := NewRetryOptionsForReadAfterCreate()

	if options.ShouldRetry(nil, nil) {
		t.Fatal("expected no retry when there is neither a response nor an error")
	}

	if !options.ShouldRetry(nil, io.ErrUnexpectedEOF) {
		t.Fatal("expected a retry when the request failed without a response")
	}
}
