package prechecks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrecheckURLsAccess(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			// HEAD and GET -> 200
			w.WriteHeader(200)
		case "/etagfail":
			// HEAD -> 500, GET -> 200
			if r.Method == http.MethodHead {
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
		case "/bad":
			// HEAD and GET -> 500
			w.WriteHeader(500)
		case "/redirect":
			// HEAD and GET -> 301 redirect
			w.Header().Set("Location", "/ok")
			w.WriteHeader(301)
		default:
			w.WriteHeader(404)
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("empty slice", func(t *testing.T) {
		if err := PrecheckURLsAccess([]string{}); err != nil {
			t.Fatalf("expected nil error for empty slice, got: %v", err)
		}
	})

	t.Run("accessible via HEAD", func(t *testing.T) {
		urls := []string{server.URL + "/ok"}
		if err := PrecheckURLsAccess(urls); err != nil {
			t.Fatalf("expected nil error for accessible url, got: %v", err)
		}
	})

	t.Run("fallback to GET", func(t *testing.T) {
		urls := []string{server.URL + "/etagfail"}
		if err := PrecheckURLsAccess(urls); err != nil {
			t.Fatalf("expected nil error when GET succeeds after HEAD fails, got: %v", err)
		}
	})

	t.Run("redirect considered accessible", func(t *testing.T) {
		urls := []string{server.URL + "/redirect"}
		if err := PrecheckURLsAccess(urls); err != nil {
			t.Fatalf("expected nil error for redirect (3xx) response, got: %v", err)
		}
	})

	t.Run("inaccessible url reported", func(t *testing.T) {
		bad := server.URL + "/bad"
		urls := []string{server.URL + "/ok", bad, "   ", ""}
		err := PrecheckURLsAccess(urls)
		if err == nil {
			t.Fatalf("expected error for inaccessible url, got nil")
		}
		if !strings.Contains(err.Error(), bad) {
			t.Fatalf("error message does not contain failing url; got: %v", err)
		}
	})
}
