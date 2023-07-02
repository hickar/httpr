package httpr

import (
	"net/http"
	"time"
)

func newDefaultSettings() clientSettings {
	return clientSettings{
		redirectCheckFn:   func(_ *http.Request, _ []*http.Request) error { return nil },
		preRequestHookFn:  func(_ *http.Request) error { return nil },
		postRequestHookFn: func(_ *http.Request, _ *Response) {},
		retryConditionFn:  func(_ *Response, err error) bool { return true },
	}
}

// Option is a function type used for altering client-scoped or request-scoped settings like
// retry count, retry delay, timeout and others.
type Option func(settings *clientSettings)

// WithRetryCount sets number of retries used for request being carried. If requests number failed equals
// specified retry count, Client.Do and all shortcut methods return corresponding error.
func WithRetryCount(retries int) Option {
	return func(settings *clientSettings) {
		settings.retryCount = retries
	}
}

// WithRetryDelay is used to specify delay being taken after unsuccessful request.
// This option is ignored if retry count is not set.
func WithRetryDelay(delay time.Duration) Option {
	return func(settings *clientSettings) {
		settings.retryDelay = delay
	}
}

// WithRetryDelayDelta is used to specify delay delta being added to delay time after each unsuccessful request.
// This option is ignored if retry count is not set.
func WithRetryDelayDelta(delayDelta time.Duration) Option {
	return func(settings *clientSettings) {
		settings.retryDelayDelta = delayDelta
	}
}

// WithTransport is used to change http.Transport used.
func WithTransport(transport http.RoundTripper) Option {
	return func(settings *clientSettings) {
		if transport != nil {
			settings.transport = transport
		}
	}
}

// WithTimeout specified timeout for request being executed. If response wasn't received within specified timeout,
// Client.Do and all shortcut methods return context.DeadlineExceeded.
func WithTimeout(timeout time.Duration) Option {
	return func(settings *clientSettings) {
		settings.timeout = timeout
	}
}

// WithCheckRedirect sets middleware function for specifying request redirect policy.
func WithCheckRedirect(checkFn func(*http.Request, []*http.Request) error) Option {
	return func(settings *clientSettings) {
		if checkFn != nil {
			settings.redirectCheckFn = checkFn
		}
	}
}

// RetryConditionFunc is function, used for specifying whether request execution must be
// attempted again. Function must return true is retry is needed, false if not.
type RetryConditionFunc func(*Response, error) bool

// WithRetryCondition sets RetryConditionFunc middleware.
func WithRetryCondition(conditionFn RetryConditionFunc) Option {
	return func(settings *clientSettings) {
		settings.retryConditionFn = conditionFn
	}
}

// WithCookieJar sets http.CookieJar used by underlying http.Client.
func WithCookieJar(cookieJar http.CookieJar) Option {
	return func(settings *clientSettings) {
		settings.cookieJar = cookieJar
	}
}

// PreRequestHookFn is function, which is called before request execution. If request execution must not take place,
// PreRequestHookFn must return non-nil error.
type PreRequestHookFn func(req *http.Request) error

// WithPreRequestHook set PreRequestHookFn compliant function.
func WithPreRequestHook(hookFn PreRequestHookFn) Option {
	return func(settings *clientSettings) {
		if hookFn != nil {
			settings.preRequestHookFn = hookFn
		}
	}
}

// PostRequestHookFn is function, which is called after request execution.
type PostRequestHookFn func(req *http.Request, resp *Response)

// WithPostRequestHook sets PostRequestHookFn compliant function.
func WithPostRequestHook(hookFn PostRequestHookFn) Option {
	return func(settings *clientSettings) {
		if hookFn != nil {
			settings.postRequestHookFn = hookFn
		}
	}
}

// WithRateLimiter sets Limiter instance. Limiter is in charged for limiting rate of requests being executed.
func WithRateLimiter(limiter Limiter) Option {
	return func(settings *clientSettings) {
		if limiter != nil {
			settings.rateLimiter = limiter
		}
	}
}

// WithAutoDecompression specifies whether response body should be unarchived automatically.
// Currently only GZIP is supported.
func WithAutoDecompression(enabled bool) Option {
	return func(settings *clientSettings) {
		settings.decompressionEnabled = enabled
	}
}

// Limiter interface is used to abstract concrete types which purpose is to set and handle rate-limiting for
// request execution.
type Limiter interface {
	Take() time.Time
}

// NewUnlimitedLimiter creates dummy struct which implements Limiter interface.
// Used for testing purposes.
func NewUnlimitedLimiter() Limiter {
	return &unlimitedLimiter{}
}

type unlimitedLimiter struct{}

func (l *unlimitedLimiter) Take() time.Time {
	return time.Now()
}
