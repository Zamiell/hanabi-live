package httpmain

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

// googleAnalyticsMiddleware is a Gin middleware for sending page views to Google Analytics.
// (we do this on the server because client-side blocking is common via uBlock Origin, Adblock Plus,
// etc.)
func (m *Manager) googleAnalyticsMiddleware(c *gin.Context) {
	// Local variables
	r := c.Request
	w := c.Writer

	// We only want to track page views for "/", "/scores/Alice", etc.
	// (this goroutine will be entered for requests to "/public/css/main.min.css", for example)
	path := c.Request.URL.Path
	if path != "/" &&
		!strings.HasPrefix(path, "/scores/") &&
		!strings.HasPrefix(path, "/profile/") &&
		!strings.HasPrefix(path, "/history/") &&
		!strings.HasPrefix(path, "/missing-scores/") &&
		!strings.HasPrefix(path, "/stats") &&
		!strings.HasPrefix(path, "/variant/") &&
		!strings.HasPrefix(path, "/videos") {

		return
	}

	// Get their Google Analytics cookie, if any
	// If they do not have one, set a new cookie
	var clientID string
	if cookie, err := r.Cookie("_ga"); err != nil {
		// They don't have a cookie set, so set a new one
		clientID = uuid.NewV4().String()
		http.SetCookie(w, &http.Cookie{ // nolint: exhaustivestruct
			// This is the standard cookie name used by the Google Analytics JavaScript library
			Name:  "_ga",
			Value: clientID,

			// The standard library does not have definitions for units of day
			// or larger to avoid confusion across daylight savings
			// We use 2 years because it is recommended by Google:
			// https://developers.google.com/analytics/devguides/collection/analyticsjs/cookie-usage
			Expires: time.Now().Add(2 * 365 * 24 * time.Hour), // 2 years

			// Bind the cookie to this specific domain for security purposes
			Domain: m.Domain,

			// Only send the cookie over HTTPS:
			// https://www.owasp.org/index.php/Testing_for_cookies_attributes_(OTG-SESS-002)
			Secure: m.UseTLS,

			// Mitigate XSS attacks:
			// https://www.owasp.org/index.php/HttpOnly
			HttpOnly: true,

			// Mitigate CSRF attacks:
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#SameSite_cookies
			SameSite: http.SameSiteStrictMode,
		})
	} else {
		clientID = cookie.Value
	}

	// Now we need to make a POST to google-analytics.com, but we need to do
	// that in a new goroutine to avoid locking up the server
	// According to the Gin documentation, we have to make a copy of the context
	// before using it inside of a goroutine
	contextCopy := c.Copy()
	go m.googleAnalyticsSend(contextCopy, clientID)
	c.Next()
}

func (m *Manager) googleAnalyticsSend(c *gin.Context, clientID string) {
	// Local variables
	r := c.Request

	// Prepare the request
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	data := url.Values{
		"v":   {"1"},            // API version
		"tid": {m.gaTrackingID}, // Tracking ID
		"cid": {clientID},       // Anonymous client ID
		"t":   {"pageview"},     // Hit type
		"dh":  {r.Host},         // Document hostname
		"dp":  {r.URL.Path},     // Document page/path
		"uip": {ip},             // IP address override
		"ua":  {r.UserAgent()},  // User agent override
	}
	url := "https://www.google-analytics.com/collect"
	var req *http.Request
	if v, err := http.NewRequestWithContext(
		c,
		http.MethodPost,
		url,
		strings.NewReader(data.Encode()),
	); err != nil {
		m.logger.Infof("Failed to prepare a request for a page hit to Google Analytics: %v", err)
		return
	} else {
		req = v
	}

	// Send it
	if resp, err := m.httpClientWithTimeout.Do(req); err != nil {
		// POSTs to Google Analytics will occasionally time out; if this occurs,
		// do not bother retrying, since losing a single page view is fairly meaningless
		m.logger.Infof("Failed to send a page hit to Google Analytics: %v", err)
		return
	} else {
		defer resp.Body.Close()
	}
}
