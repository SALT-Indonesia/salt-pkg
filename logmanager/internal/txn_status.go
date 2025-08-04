package internal

import (
	"net/http"
)

var warningHttpStatusCodes = []int{
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusNotFound,
	http.StatusMethodNotAllowed,
	http.StatusNotAcceptable,
	http.StatusProxyAuthRequired,
	http.StatusRequestTimeout,
	http.StatusConflict,
	http.StatusGone,
}

var errorHttpStatusCodes = []int{
	http.StatusInternalServerError,
	http.StatusNotImplemented,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
	http.StatusHTTPVersionNotSupported,
}

func HasErrorBusinessFromHttpStatusCode(statusCode int) bool {
	if isResponseSuccess(statusCode) {
		return false
	}

	return isWarningStatusCode(statusCode)
}

func HasErrorInternalFromHttpStatusCode(statusCode int) bool {
	if isResponseSuccess(statusCode) {
		return false
	}

	if isErrorStatusCode(statusCode) {
		return true
	}

	return !isWarningStatusCode(statusCode)
}

func isResponseSuccess(code int) bool {
	return (code >= 200 && code <= 299) || code == http.StatusTemporaryRedirect
}

func isWarningStatusCode(code int) bool {
	for _, c := range warningHttpStatusCodes {
		if code == c {
			return true
		}
	}
	return false
}

func isErrorStatusCode(code int) bool {
	for _, c := range errorHttpStatusCodes {
		if code == c {
			return true
		}
	}
	return false
}
