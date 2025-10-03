package version

// Version information set at build time via ldflags
var (
	GitHash = "dev"
)

// GetVersion returns just the version string
func GetVersion() string {
	return GitHash
}

// GetUserAgent returns the user agent string for HTTP requests
func GetUserAgent() string {
	return "NewsGoat " + GitHash + "; +https://github.com/jarv/newsgoat"
}