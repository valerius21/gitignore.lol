package lib

var CLI struct {
	Port            int    `help:"Port the server listens on." name:"port" default:"4444"`
	BaseRepository  string `help:"Gitignore repository where the .gitignore files are versioned." name:"repo" default:"https://github.com/github/gitignore.git" type:"url"`
	ClonePath       string `help:"Location of the locally stored gitignore repository" name:"clone-path" default:"./store" type:"path"`
	UpdateInterval  int    `help:"Interval (seconds) in which the linked repository gets updated" name:"fetch-interval" default:"300"`
	RateLimit       int    `help:"Maximum requests per window per IP" name:"rate-limit" default:"100"`
	RateWindow      int    `help:"Rate limiting window in seconds" name:"rate-window" default:"60"`
	RateCleanupMs   int    `help:"Cleanup interval for rate limiter in milliseconds" name:"rate-cleanup" default:"30000"`
	EnableRateLimit bool   `help:"Enable rate limiting" name:"enable-rate-limit" default:"true"`

	// Enhanced rate limiting for scanner protection
	UseEnhancedLimiter bool `help:"Use enhanced rate limiter with scanner protection" name:"enhanced-limiter" default:"false"`
	ErrorRateLimit     int  `help:"Maximum 404/error requests per window per IP" name:"error-rate-limit" default:"10"`
	BlockMinutes       int  `help:"Minutes to block IPs that exceed limits" name:"block-minutes" default:"5"`
	MaxViolations      int  `help:"Max violations before longer blocks" name:"max-violations" default:"3"`
}
