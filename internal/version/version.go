package version

// Version information set at build time via ldflags
var (
	GitHash = "dev"
)

// GetUserAgent returns the user agent string for HTTP requests
func GetUserAgent() string {
	return "NewsGoat " + GitHash + "; +https://github.com/jarv/newsgoat"
}