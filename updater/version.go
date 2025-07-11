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

// GetLatestBedrockVersion fetches and parses the latest version from the Minecraft release changelogs page, and constructs the download URL.
// GetLatestBedrockVersion fetches and parses the latest version from the Minecraft release changelogs page, and constructs the download URL using the provided base URL.
func GetLatestBedrockVersion(changelogUrl, baseDownloadUrl string) (version string, zipUrl string, err error) {
	if changelogUrl == "" {
		changelogUrl = "https://feedback.minecraft.net/hc/en-us/sections/360001186971-Release-Changelogs"
	}
	if baseDownloadUrl == "" {
		baseDownloadUrl = "https://www.minecraft.net/bedrockdedicatedserver/bin-win/"
	}
	var lastErr error
	for i := 0; i < 3; i++ {
		version, err := fetchLatestVersionFromChangelog(changelogUrl)
		if err == nil && version != "" {
			zipUrl := baseDownloadUrl + "bedrock-server-" + version + ".zip"
			return version, zipUrl, nil
		}
		lastErr = err
		time.Sleep(time.Duration(1+i) * time.Second)
	}
	return "", "", fmt.Errorf("update check failed - version info is unavailable: %w", lastErr)
}

// fetchLatestVersionFromChangelog fetches the changelog page and parses the latest Bedrock version.
func fetchLatestVersionFromChangelog(pageUrl string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Example: <a href="/hc/en-us/articles/37810171798029-Minecraft-1-21-93-Bedrock" ...>Minecraft - 1.21.93 (Bedrock)</a>
	re := regexp.MustCompile(`(?i)Minecraft\s*-\s*([\d.]+)\s*\(Bedrock\)`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", errors.New("bedrock version not found in changelog page")
	}
	return matches[1], nil
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
	re := regexp.MustCompile(`href=['"]([^'"]*bedrock-server-([\d.]+)\.zip)['"]`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 3 {
		return "", "", errors.New("version or zip url not found in page")
	}
	zipUrl := matches[1]
	version := matches[2]
	return version, zipUrl, nil
}
