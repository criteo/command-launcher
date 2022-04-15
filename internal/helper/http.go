package helper

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
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

func HttpEtag(url string) (int, string, error) {
	resolvedUrl, resolved := ResolveUrl(url)
	if resolved {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	resp, err := http.Head(resolvedUrl)
	if err != nil {
		return 0, "", err
	}

	if !Is2xx(resp.StatusCode) {
		return resp.StatusCode, "", fmt.Errorf("failed to read etage from %s, status code %d", url, resp.StatusCode)
	}

	return resp.StatusCode, strings.Trim(resp.Header.Get("etag"), "\""), nil
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
	return url, false
}

func BodyAsString(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
