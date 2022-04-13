package helper

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	u "net/url"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

func HttpGetWithBasicAuth(url, user, password string) (int, []byte, error) {
	return HttpDoWithBasicAuth("GET", url, user, password, nil)
}

func HttpPostWithBasicAuth(url, user, password string) (int, []byte, error) {
	return HttpDoWithBasicAuth("POST", url, user, password, nil)
}

func HttpPostInputWithBasicAuth(url, user, password string, input io.Reader) (int, []byte, error) {
	return HttpDoWithBasicAuth("POST", url, user, password, input)
}

func HttpDoWithBasicAuth(method, url, user, password string, input io.Reader) (int, []byte, error) {
	req, err := http.NewRequest(method, url, input)
	if err != nil {
		return 0, nil, err
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	req.SetBasicAuth(user, password)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	if !Is2xx(resp.StatusCode) {
		return resp.StatusCode, nil, fmt.Errorf("failed to get resources %s, status code %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	return resp.StatusCode, body, nil
}

func HttpGet(url string) (int, []byte, error) {
	resp, err := HttpGetWrapper(url)
	if err != nil {
		return 0, nil, err
	}
	if !Is2xx(resp.StatusCode) {
		return resp.StatusCode, nil, fmt.Errorf("failed to get resource %s, status code %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	return resp.StatusCode, body, nil
}

func HttpGetWrapper(url string) (*http.Response, error) {
	resolvedUrl, resolved := ResolveUrl(url)
	if resolved {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return http.Get(resolvedUrl)
}

func HttpNewRequestWrapper(method, url string, body io.Reader) (*http.Request, error) {
	resolvedUrl, resolved := ResolveUrl(url)
	if resolved {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return http.NewRequest(method, resolvedUrl, body)
}

func ResolveUrl(url string) (string, bool) {
	// disable mac os dns resolver
	// TODO: remove this function, when the testing of macosx resolver is done
	return url, false

	if runtime.GOOS != "darwin" {
		return url, false
	}

	log.Debugf("[Darwin] try to resolve %s", url)

	// use dscacheutil tool to resolve the dns correctly in darwin
	// this is because macOS uses a local dns resolver, which is not used by golang's native resolver
	urlObj, err := u.Parse(url)
	if err != nil {
		log.Debugf("Parsing %s has failed", url)
		return url, false
	}

	ip, resolved := DarwinDnsResolve(urlObj.Host)
	if !resolved {
		return url, false
	}
	urlObj.Host = ip
	return urlObj.String(), true
}

func DarwinDnsResolve(host string) (string, bool) {
	if runtime.GOOS != "darwin" {
		return host, false
	}

	dir, err := os.Getwd()
	if err != nil {
		return host, false
	}
	code, output, err := CallExternalWithOutput([]string{}, dir, "dscacheutil", "-q", "host", "-a", "name", host)
	if err != nil {
		return host, false
	}
	if code != 0 {
		return host, false
	}
	for _, line := range strings.Split(output, "\n") {
		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			if strings.TrimSpace(parts[0]) == "ip_address" {
				ip := strings.TrimSpace(parts[1])
				return ip, true
			}
		}
	}
	return host, false
}

func BodyAsString(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
