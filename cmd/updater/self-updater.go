package updater

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"time"

	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/criteo/command-launcher/internal/user"
	"github.com/inconshreveable/go-update"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type LatestVersion struct {
	Version        string `json:"version" yaml:"version"`
	ReleaseNotes   string `json:"releaseNotes" yaml:"releaseNotes"`
	StartPartition uint8  `json:"startPartition" yaml:"startPartition"`
	EndPartition   uint8  `json:"endPartition" yaml:"endPartition"`
}

type SelfUpdater struct {
	selfUpdateChan <-chan bool
	latestVersion  LatestVersion

	BinaryName        string
	LatestVersionUrl  string
	SelfUpdateRootUrl string
	User              user.User
	CurrentVersion    string
	Timeout           time.Duration
}

func (u *SelfUpdater) CheckUpdateAsync() {
	ch := make(chan bool, 1)
	u.selfUpdateChan = ch
	go func() {
		select {
		case value := <-u.checkSelfUpdate():
			ch <- value
		case <-time.After(u.Timeout):
			ch <- false
		}
	}()
}

func (u *SelfUpdater) Update() error {
	canBeSelfUpdated := <-u.selfUpdateChan || helper.LoadDebugFlags().ForceSelfUpdate
	if !canBeSelfUpdated {
		return nil
	}

	fmt.Println("\n-----------------------------------")
	fmt.Printf("🚀 %s version %s \n", u.BinaryName, u.CurrentVersion)
	fmt.Printf("\nan update of %s (%s) is available:\n\n", u.BinaryName, u.latestVersion.Version)
	fmt.Println(u.latestVersion.ReleaseNotes)
	fmt.Println()
	console.Reminder("do you want to update it? [yN]")
	var resp int
	if _, err := fmt.Scanf("%c", &resp); err != nil || (resp != 'y' && resp != 'Y') {
		fmt.Println("aborted by user")
		return fmt.Errorf("Aborted by user")
	}

	fmt.Printf("update and install the latest version of %s (%s)\n", u.BinaryName, u.latestVersion.Version)
	downloadUrl, err := u.downloadUrl(u.latestVersion.Version)
	if err != nil {
		console.Error("update failed: %s\n", err)
		return err
	}
	if err = u.doSelfUpdate(downloadUrl); err != nil {
		// fallback to legacy self update
		if err = u.legacySelfUpdate(); err != nil {
			console.Error("update failed: %s\n", err)
			return err
		}
	}

	return nil
}

func (u *SelfUpdater) checkSelfUpdate() <-chan bool {
	ch := make(chan bool, 1)
	go func() {
		data, err := helper.LoadFile(u.LatestVersionUrl)
		if err != nil {
			log.Infof(err.Error())
			ch <- false
			return
		}

		u.latestVersion = LatestVersion{}
		// YAML is a supper set of json, should work with JSON as well.
		err = yaml.Unmarshal(data, &u.latestVersion)
		if err != nil {
			log.Errorf(err.Error())
			ch <- false
			return
		}

		ch <- u.latestVersion.Version != u.CurrentVersion &&
			u.User.InPartition(u.latestVersion.StartPartition, u.latestVersion.EndPartition)
	}()
	return ch
}

func (u *SelfUpdater) doSelfUpdate(url string) error {
	log.Debugf("Update %s version %s from %s", u.BinaryName, u.latestVersion.Version, url)
	resp, err := helper.HttpGetWrapper(url)
	if err != nil {
		return fmt.Errorf("cannot download the new version from %s: %v", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cannot download the new version from %s: code %d", url, resp.StatusCode)
	}

	defer resp.Body.Close()
	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		if err = update.RollbackError(err); err != nil {
			return fmt.Errorf("update failed, unfortunately, the rollback did not work neither: %v\nplease contact #build-services team", err)
		}
		console.Warn("update failed, rollback to previous version: %v\n", err)
	}

	return nil
}

func (u *SelfUpdater) downloadUrl(version string) (string, error) {
	updateUrl, err := url.Parse(u.SelfUpdateRootUrl)
	if err != nil {
		return "", err
	}

	// the download url convention: [self_update_base_url]/[version]/[binaryName]_[OS]_[ARCH]_[version][extension]
	// Example: https://github.com/criteo/command-launcher/releases/download/1.6.0/cdt_darwin_arm64_1.6.0"
	updateUrl.Path = path.Join(updateUrl.Path, version, u.binaryFileName(version))
	return updateUrl.String(), nil
}

func (u *SelfUpdater) binaryFileName(version string) string {
	downloadFileName := fmt.Sprintf("%s_%s_%s_%s", u.BinaryName, runtime.GOOS, runtime.GOARCH, version)
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s.exe", downloadFileName)
	}
	return downloadFileName
}

// deprecated. Keep it here for backward compatibility, will be removed in 1.8.0
func (u *SelfUpdater) legacySelfUpdate() error {
	legacyUrl, err := u.legacyLatestDownloadUrl()
	if err != nil {
		return err
	}
	if err := u.doSelfUpdate(legacyUrl); err != nil {
		return err
	}
	return nil
}

func (u *SelfUpdater) legacyLatestDownloadUrl() (string, error) {
	updateUrl, err := url.Parse(u.SelfUpdateRootUrl)
	if err != nil {
		return "", err
	}

	updateUrl.Path = path.Join(updateUrl.Path, "current", runtime.GOOS, runtime.GOARCH, u.binaryFileNameWithoutVersion())
	return updateUrl.String(), nil
}

func (u *SelfUpdater) binaryFileNameWithoutVersion() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s.exe", u.BinaryName)
	}
	return u.BinaryName
}
