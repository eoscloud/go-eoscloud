package auth

// SharedFlowConfig holds configuration fields that are common to all OAuth2 flow configs.
type SharedFlowConfig struct {
	clientID               string
	authBaseURL            string
	keyringServiceName     string
	fallbackTokenStoreDir  string
	allowOrganizationScope bool
}

// AuthBaseURL returns the auth base URL.
func (f *SharedFlowConfig) AuthBaseURL() string {
	return f.authBaseURL
}

// ClientID returns the OAuth2 client ID.
func (f *SharedFlowConfig) ClientID() string {
	return f.clientID
}

// KeyRingServiceName returns the keyring service name.
func (f *SharedFlowConfig) KeyRingServiceName() string {
	return f.keyringServiceName
}

// FallbackTokenStoreDir returns the directory used for file-based token storage
// when the OS keyring is not available.
func (f *SharedFlowConfig) FallbackTokenStoreDir() string {
	return f.fallbackTokenStoreDir
}

// defaultFlowConfig returns a SharedFlowConfig populated with package-level defaults.
func defaultFlowConfig() SharedFlowConfig {
	return SharedFlowConfig{
		authBaseURL:           DefaultAuthBaseURL,
		keyringServiceName:    DefaultKeyringServiceName,
		fallbackTokenStoreDir: DefaultTokenStoreDir(),
	}
}
