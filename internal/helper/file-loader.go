package helper

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	grab "github.com/cavaliergopher/grab/v3"
)

// Load file from http(s) or local disk
// Use http:// https:// as prefix for remote file
// Use file:// or no prefix for local file
func LoadFile(fileUrlOrPath string) ([]byte, error) {
	location := fileUrlOrPath
	if strings.HasPrefix(location, "http") {
		return LoadFileFromUrl(location)
	}
	location = strings.TrimPrefix(location, "file://")
	return ioutil.ReadFile(location)
}

// Load a file from a http(s) url
func LoadFileFromUrl(url string) ([]byte, error) {
	resp, err := HttpGetWrapper(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file from %s", url)
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// Download file from http(s) or local disk
// Use http:// https:// as prefix for remote file
// Use file:// or no prefix for local file
func DownloadFile(fileUrlOrPath string, dest string, showProgress bool) error {
	location := fileUrlOrPath
	if strings.HasPrefix(location, "http") {
		return DownloadFileFromUrl(location, dest, showProgress)
	}
	location = strings.TrimPrefix(location, "file://")
	return CopyLocalFile(location, dest, showProgress)
}

func DownloadFileFromUrl(url string, dest string, showProgress bool) error {
	client := grab.NewClient()

	resolvedUrl, resolved := ResolveUrl(url) // fix mac OS issue
	if resolved {
		client.HTTPClient.(*http.Client).Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	req, err := grab.NewRequest(dest, resolvedUrl)
	if err != nil {
		return fmt.Errorf("cannot get request from the server (%v)", err)
	}

	fmt.Println("Initializing download...")
	resp := client.Do(req)

	if showProgress {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()

		for !resp.IsComplete() {
			select {
			case <-t.C:
				fmt.Printf("\033[1Atransferred %.2f%%\033[K\n", 100*resp.Progress())
			default:
			}
		}

		// clear progress line
		fmt.Printf("\033[1A\033[K")
	}

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", url, err)
		return fmt.Errorf("error downloading %s: %v", url, err)
	}
	return nil
}

func CopyLocalFile(src string, dest string, showProgress bool) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destination.Close()

	// TODO: show progress here
	_, err = io.Copy(destination, source)

	return err
}
