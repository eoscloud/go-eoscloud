package api

import "net/http"

// HTTPStatusCheck defines a function to check for valid response status codes
type HTTPStatusCheck func(i int) bool

// IsHTTPStatusCreated check if the response status code is http.StatusCreated
func IsHTTPStatusCreated(i int) bool {
	return i == http.StatusCreated
}

// IsHTTPStatusOK check if the response status code is http.StatusOK
func IsHTTPStatusOK(i int) bool {
	return i == http.StatusOK
}

// IsHTTPStatusNoContent check if the response status code is http.StatusNoContent
func IsHTTPStatusNoContent(i int) bool {
	return i == http.StatusNoContent
}

// IsHTTPStatusAccepted checks if the response status code is http.StatusAccepted
func IsHTTPStatusAccepted(i int) bool {
	return i == http.StatusAccepted
}

// IsHTTPStatusOKOrAccepted checks if the response status code is either
// http.StatusOK or http.StatusAccepted. It is intended for endpoints that may
// answer synchronously (200) or asynchronously (202) depending on server-side
// state — for example PATCH/PUT on resources that can transition into an
// intermediate state.
func IsHTTPStatusOKOrAccepted(i int) bool {
	return i == http.StatusOK || i == http.StatusAccepted
}
