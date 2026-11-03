package api

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// joinIDPath builds "<base>/<escaped id>/". The id is path-escaped so that
// characters such as "/", "?" or "#" cannot change the structure of the
// request URL. Empty ids and the dot segments "." and ".." are rejected.
func joinIDPath(base, id string) (string, error) {
	if id == "" {
		return "", errors.New("invalid id: must not be empty")
	}
	if id == "." || id == ".." {
		return "", fmt.Errorf("invalid id %q", id)
	}
	return strings.TrimRight(base, "/") + "/" + url.PathEscape(id) + "/", nil
}
