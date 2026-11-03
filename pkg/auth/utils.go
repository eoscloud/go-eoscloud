package auth

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"unicode"
)

func isWSL() bool {
	releaseData, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(releaseData)), "microsoft")
}

// isLoopbackHost reports whether host (without port) is localhost or a loopback IP.
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// checkURLScheme requires https, allowing http only for loopback hosts.
func checkURLScheme(u *url.URL) error {
	if u.Hostname() == "" {
		return errors.New("missing host")
	}
	if u.User != nil {
		return errors.New("userinfo is not allowed")
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return nil
	case "http":
		if isLoopbackHost(u.Hostname()) {
			return nil
		}
		return errors.New("http is only allowed for loopback hosts, use https")
	default:
		return fmt.Errorf("unsupported scheme %q, use https", u.Scheme)
	}
}

// validateAuthBaseURL checks that the auth base URL parses, uses https (http is
// accepted for loopback hosts only) and has a host.
func validateAuthBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid auth base URL %q: %w", raw, err)
	}
	if err := checkURLScheme(u); err != nil {
		return fmt.Errorf("invalid auth base URL %q: %w", raw, err)
	}
	return nil
}

// validateVerificationURL checks that a verification URL returned by the
// server is safe to hand to a browser: https (http for loopback only) and the
// same host as the auth server. authURL is any URL on the auth server.
func validateVerificationURL(raw, authURL string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid verification URL: %w", err)
	}
	if err := checkURLScheme(u); err != nil {
		return fmt.Errorf("invalid verification URL: %w", err)
	}
	base, err := url.Parse(authURL)
	if err != nil || base.Hostname() == "" {
		return errors.New("cannot determine auth server host")
	}
	if !strings.EqualFold(u.Hostname(), base.Hostname()) {
		return fmt.Errorf("verification URL host %q does not match auth server host %q", u.Hostname(), base.Hostname())
	}
	return nil
}

// sanitizeForTerminal drops control characters so server-supplied text cannot
// inject terminal escape sequences.
func sanitizeForTerminal(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// start launches cmd without waiting and reaps it in the background.
func start(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

// tryOpenBrowser attempts to open the URL. None of the launchers go through a
// shell, so the URL is never interpreted as a command line.
func tryOpenBrowser(u string) error {
	switch runtime.GOOS {
	case "darwin":
		return start(exec.Command("open", u)) //nolint:gosec // opens the validated login URL in the user's browser
	case "windows":
		return start(exec.Command("rundll32", "url.dll,FileProtocolHandler", u)) //nolint:gosec // opens the validated login URL in the user's browser
	case "linux":
		if isWSL() {
			return start(exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", u)) //nolint:gosec // opens the validated login URL in the user's browser
		}
		// Headless / SSH: no display available.
		if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
			return errors.New("no display available")
		}
		if path, err := exec.LookPath("xdg-open"); err == nil {
			return start(exec.Command(path, u)) //nolint:gosec // opens the validated login URL in the user's browser
		}
		// Fallback: honour $BROWSER if set.
		if browser := os.Getenv("BROWSER"); browser != "" {
			return start(exec.Command(browser, u)) //nolint:gosec // opens the validated login URL in the user's browser
		}
		return errors.New("no browser launcher found")
	default:
		return fmt.Errorf("opening a browser is not supported on %s", runtime.GOOS)
	}
}

// openBrowserFunc is the built-in browser opener. It can be replaced in tests
// to prevent actual browser launches.
var openBrowserFunc = tryOpenBrowser
