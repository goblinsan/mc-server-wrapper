package updater

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)


// GetLatestBedrockVersion fetches the latest Bedrock version and download link from minecraft.wiki
func GetLatestBedrockVersion(wikiNavUrl string) (version string, zipUrl string, err error) {
	if wikiNavUrl == "" {
		wikiNavUrl = "https://minecraft.wiki/"
	}
	var lastErr error
	for i := 0; i < 3; i++ {
		version, pageUrl, err := fetchLatestBedrockVersionFromWikiNav(wikiNavUrl)
		if err == nil && version != "" && pageUrl != "" {
			zipUrl, err := fetchBedrockDownloadLinkFromWikiPage(pageUrl)
			if err == nil && zipUrl != "" {
				return version, zipUrl, nil
			}
			lastErr = err
		} else {
			lastErr = err
		}
		time.Sleep(time.Duration(1+i) * time.Second)
	}
	return "", "", fmt.Errorf("update check failed - version info is unavailable: %w", lastErr)
}

func fetchLatestBedrockVersionFromWikiNav(wikiNavUrl string) (string, string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", wikiNavUrl, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
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
	nav := string(body)
	bedrockNavMarker := `<li id="n-Bedrock-Edition"`
	bedrockIdx := strings.Index(nav, bedrockNavMarker)
	if bedrockIdx == -1 {
		bedrockNavMarker = `<li id="n-Bedrock-Edition" class="mw-list-item"><a href="/w/Bedrock_Edition" title="Bedrock Edition"><span>Bedrock Edition</span></a></li>`
		bedrockIdx = strings.Index(nav, bedrockNavMarker)
		if bedrockIdx == -1 {
			bedrockIdx = strings.Index(nav, ">Bedrock Edition<")
			reTest := regexp.MustCompile(`<li id="n-Latest:-([\d.]+)"[^>]*><a href="([^"]+)"[^>]*>.*?</a></li>`)
			matchesTest := reTest.FindStringSubmatch(nav)
			if bedrockIdx == -1 && len(matchesTest) >= 3 {
				bedrockIdx = 0
			} else if bedrockIdx == -1 {
				return "", "", errors.New("Bedrock Edition section not found in wiki nav")
			}
		}
	}
	nav = nav[bedrockIdx:]
	re := regexp.MustCompile(`<li id="n-Latest:-([\d.]+)"[^>]*><a href="([^"]+)"[^>]*><span>Latest: [\d.]+</span></a></li>`)
	matches := re.FindStringSubmatch(nav)
	if len(matches) < 3 {
		re = regexp.MustCompile(`<li id="n-Latest:-([\d.]+)"[^>]*><a href="([^"]+)"[^>]*>.*?</a></li>`)
		matches = re.FindStringSubmatch(nav)
		if len(matches) < 3 {
			return "", "", errors.New("latest bedrock version not found in wiki nav")
		}
	}
	version := matches[1]
	pagePath := matches[2]
	
	base := wikiNavUrl
	if strings.HasPrefix(wikiNavUrl, "http://") || strings.HasPrefix(wikiNavUrl, "https://") {
		u := wikiNavUrl
		if strings.HasSuffix(u, "/") {
			u = u[:len(u)-1]
		}
		base = u
	}
	
	var pageUrl string
	if strings.HasPrefix(pagePath, "http") {
		pageUrl = pagePath
	} else {
		pageUrl = base + pagePath
	}
	return version, pageUrl, nil
}

func fetchBedrockDownloadLinkFromWikiPage(pageUrl string) (string, error) {
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
	
	re := regexp.MustCompile(`<a[^>]+href="([^"]*bedrockdedicatedserver/bin-win/bedrock-server-[\d.]+\.zip)"[^>]*>Windows</a>`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", errors.New("bedrock server download link not found in wiki page")
	}
	return matches[1], nil
}

func ParseBedrockVersion(s string) string {
	re := regexp.MustCompile(`bedrock-server-([\d.]+)\.zip`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

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
