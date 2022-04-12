package helper

import "net/http"

func Is2xx(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode <= http.StatusIMUsed
}
