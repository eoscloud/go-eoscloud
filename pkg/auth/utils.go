package auth

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func isWSL() bool {
	releaseData, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(releaseData)), "microsoft")
}

// tryOpenBrowser attempts to open the URL and returns true on success.
func tryOpenBrowser(url string) bool {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start() == nil //nolint:gosec // opens the login URL in the user's browser
	case "windows":
		// Replace & with ^& to prevent cmd.exe from interpreting it as a command separator.
		escaped := strings.ReplaceAll(url, "&", "^&")
		return exec.Command("cmd", "/c", "start", escaped).Start() == nil //nolint:gosec // opens the login URL in the user's browser
	case "linux":
		if isWSL() {
			return exec.Command("cmd.exe", "/c", "start", url).Start() == nil //nolint:gosec // opens the login URL in the user's browser
		}
		// Headless / SSH: no display available.
		if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
			return false
		}
		if path, err := exec.LookPath("xdg-open"); err == nil {
			return exec.Command(path, url).Start() == nil //nolint:gosec // opens the login URL in the user's browser
		}
		// Fallback: honour $BROWSER if set.
		if browser := os.Getenv("BROWSER"); browser != "" {
			return exec.Command(browser, url).Start() == nil //nolint:gosec // opens the login URL in the user's browser
		}
		return false

	default:
		return false
	}
}

// openBrowserFunc is the function used to open a browser URL.
// It can be replaced in tests to prevent actual browser launches.
var openBrowserFunc = defaultOpenBrowser

// defaultOpenBrowser attempts to open the given URL in a browser.
// On headless/remote sessions (SSH, no display, etc.) it prints the URL to stdout instead.
func defaultOpenBrowser(url string) {
	if tryOpenBrowser(url) {
		return
	}
	fmt.Printf("\nFailed to open browser.\n")
	fmt.Printf("Please open above URL manually and enter verification code.\n")
}
