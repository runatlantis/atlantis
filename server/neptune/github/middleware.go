package github

import (
	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/uber-go/tally/v4"
	"net/http"
	"strconv"
)

const (
	MetricsKeyRequests    = "github.requests"
	MetricsKeyRequests2xx = "github.requests.2xx"
	MetricsKeyRequests3xx = "github.requests.3xx"
	MetricsKeyRequests4xx = "github.requests.4xx"
	MetricsKeyRequests5xx = "github.requests.5xx"

	MetricsKeyRequestsCached = "github.requests.cached"

	MetricsKeyRateLimit          = "github.rate.limit"
	MetricsKeyRateLimitRemaining = "github.rate.remaining"
)

// ClientMetrics creates client middleware that records metrics about all
// requests.
//
// Pretty much copied from:
// https://github.com/palantir/go-githubapp/blob/develop/githubapp/middleware.go#L41
func ClientMetrics(scope tally.Scope) githubapp.ClientMiddleware {

	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			res, err := next.RoundTrip(r)

			if res != nil {
				scope.Counter(MetricsKeyRequests).Inc(1)
				if key := bucketStatus(res.StatusCode); key != "" {
					scope.Counter(key).Inc(1)
				}

				if res.Header.Get(httpcache.XFromCache) != "" {
					scope.Counter(MetricsKeyRequestsCached).Inc(1)
				}

				// Headers from https://developer.github.com/v3/#rate-limiting
				updateRegistryForHeader(res.Header, "X-RateLimit-Limit", scope.Gauge(MetricsKeyRateLimit))
				updateRegistryForHeader(res.Header, "X-RateLimit-Remaining", scope.Gauge(MetricsKeyRateLimitRemaining))
			}

			return res, err
		})
	}
}

func updateRegistryForHeader(headers http.Header, header string, metric tally.Gauge) {
	headerString := headers.Get(header)
	if headerString != "" {
		headerVal, err := strconv.ParseFloat(headerString, 64)
		if err == nil {
			metric.Update(headerVal)
		}
	}
}

func bucketStatus(status int) string {
	switch {
	case status >= 200 && status < 300:
		return MetricsKeyRequests2xx
	case status >= 300 && status < 400:
		return MetricsKeyRequests3xx
	case status >= 400 && status < 500:
		return MetricsKeyRequests4xx
	case status >= 500 && status < 600:
		return MetricsKeyRequests5xx
	}
	return ""
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
