package prechecks

import (
	"fmt"
	"strings"
	"sync"

	"github.com/criteo/command-launcher/internal/helper"
)

// PrecheckURLsAccess checks accessibility of the provided URLs using helper functions.
// It returns nil when all URLs are accessible (HTTP status 2xx or 3xx).
// If any URL is inaccessible an error is returned that lists the failing URLs.
func PrecheckURLsAccess(urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	var (
		mu     sync.Mutex
		failed []string
		wg     sync.WaitGroup
	)

	for _, raw := range urls {
		u := strings.TrimSpace(raw)
		if u == "" {
			continue
		}
		wg.Add(1)

		go func(u string) {
			defer wg.Done()

			// Try HttpEtag first (lightweight HEAD request)
			// which leverages URL resolution and TLS configuration from helper
			statusCode, _, err := helper.HttpEtag(u)
			if err == nil && helper.Is2xx(statusCode) {
				// URL is accessible
				return
			}

			// Fallback to HttpGet if HttpEtag fails or returns non-2xx
			statusCode, _, err = helper.HttpGet(u)
			if err != nil || !helper.Is2xx(statusCode) {
				mu.Lock()
				failed = append(failed, u)
				mu.Unlock()
			}
		}(u)
	}

	wg.Wait()

	if len(failed) == 0 {
		return nil
	}
	return fmt.Errorf("inaccessible urls: %s", strings.Join(failed, ", "))
}
