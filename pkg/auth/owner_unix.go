//go:build !windows

package auth

import (
	"os"
	"syscall"
)

// ownedByCurrentUser reports whether fi is owned by the current user. When the
// platform does not expose ownership it conservatively reports false.
func ownedByCurrentUser(fi os.FileInfo) bool {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	return int64(st.Uid) == int64(os.Getuid())
}
