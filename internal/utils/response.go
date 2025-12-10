package utils

import (
	"errors"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func ResponseErrorWasNotFound(err error) bool {
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) {
		// Check if HTTP status code is 404 (Not Found)
		if responseErr.StatusCode == http.StatusNotFound {
			return true
		}
		// Also check if the error code indicates resource not found
		// Some services return 400 Bad Request with ErrorCode "ResourceNotFound"
		if responseErr.ErrorCode == "ResourceNotFound" {
			return true
		}
	}
	return false
}

func ResponseErrorWasStatusCode(err error, statusCode int) bool {
	var responseErr *azcore.ResponseError
	return errors.As(err, &responseErr) && responseErr.StatusCode == statusCode
}
