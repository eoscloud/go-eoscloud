package api

import "net/http"

// Method defines allowed http methods
type Method string

const (
	GET     Method = http.MethodGet
	POST    Method = http.MethodPost
	PUT     Method = http.MethodPut
	PATCH   Method = http.MethodPatch
	DELETE  Method = http.MethodDelete
	OPTIONS Method = http.MethodOptions
)
