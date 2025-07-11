package updater

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// IsNewVersionAvailable compares the current and latest version strings and returns true if a new version is available.
func IsNewVersionAvailable(current, latest string) bool {
	return current != latest
}

// GetLatestBedrockVersion fetches and parses the latest version and zip URL from the download page, with retries.
func GetLatestBedrockVersion(cfgUrl string) (version string, zipUrl string, err error) {
	var lastErr error
	for i := 0; i < 3; i++ {
		version, zipUrl, err := fetchAndParseVersion(cfgUrl)
		if err == nil {
			return version, zipUrl, nil
		}
		lastErr = err
		time.Sleep(time.Duration(1+i) * time.Second)
	}
	return "", "", fmt.Errorf("update check failed - version info is unavailable: %w", lastErr)
}

func fetchAndParseVersion(baseUrl string) (string, string, error) {
	resp, err := http.Get(baseUrl)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	version, zipUrl, err := ParseBedrockVersionAndUrl(baseUrl, string(body))
	if err != nil {
		return "", "", err
	}
	return version, zipUrl, nil
}

// ParseBedrockVersion extracts the version from a filename or string.
func ParseBedrockVersion(s string) string {
	re := regexp.MustCompile(`bedrock-server-([\d.]+)\.zip`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// ParseBedrockVersionAndUrl extracts the version and zip URL from the HTML body using the base URL.
func ParseBedrockVersionAndUrl(baseUrl, body string) (string, string, error) {
	re := regexp.MustCompile(`href="([^"]*bedrock-server-([\d.]+)\.zip)"`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 3 {
		return "", "", errors.New("version or zip url not found in page")
	}
	zipUrl := matches[1]
	version := matches[2]
	return version, zipUrl, nil
}
