//go:build windows

package auth

import "os"

// ownedByCurrentUser is a no-op on Windows, where ownership and permission
// bits are not expressed through POSIX uid/mode semantics.
func ownedByCurrentUser(_ os.FileInfo) bool {
	return true
}
