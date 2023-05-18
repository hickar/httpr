package httpr

import (
	"net/http"
	"time"
)

type Option func(settings *clientSettings)

func WithRetryCount(retries int) Option {
	return func(settings *clientSettings) {
		settings.retryCount = retries
	}
}

func WithRetryDelay(delay time.Duration) Option {
	return func(settings *clientSettings) {
		settings.retryDelay = delay
	}
}

func WithRetryDelayDelta(delayDelta time.Duration) Option {
	return func(settings *clientSettings) {
		settings.retryDelayDelta = delayDelta
	}
}

func WithTransport(transport http.RoundTripper) Option {
	return func(settings *clientSettings) {
		if transport != nil {
			settings.transport = transport
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(settings *clientSettings) {
		settings.timeout = timeout
	}
}

func WithCheckRedirect(checkFn func(*http.Request, []*http.Request) error) Option {
	return func(settings *clientSettings) {
		if checkFn != nil {
			settings.redirectCheckFn = checkFn
		}
	}
}

func WithRetryCondition(conditionFn RetryConditionFunc) Option {
	return func(settings *clientSettings) {
		settings.retryConditionFn = conditionFn
	}
}

func WithCookieJar(cookieJar http.CookieJar) Option {
	return func(settings *clientSettings) {
		settings.cookieJar = cookieJar
	}
}

type PreRequestHookFn func(req *http.Request) error

func WithPreRequestHook(hookFn PreRequestHookFn) Option {
	return func(settings *clientSettings) {
		if hookFn != nil {
			settings.preRequestHookFn = hookFn
		}
	}
}

type PostRequestHookFn func(req *http.Request, resp *Response) error

func WithPostRequestHook(hookFn PostRequestHookFn) Option {
	return func(settings *clientSettings) {
		if hookFn != nil {
			settings.postRequestHookFn = hookFn
		}
	}
}

func WithRateLimiter(limiter Limiter) Option {
	return func(settings *clientSettings) {
		if limiter != nil {
			settings.rateLimiter = limiter
		}
	}
}

type unlimitedLimiter struct{}

func (l *unlimitedLimiter) Take() time.Time {
	return time.Now()
}
