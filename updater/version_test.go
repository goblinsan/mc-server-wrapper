package updater

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetLatestBedrockVersion(t *testing.T) {
	// Mock changelog HTML with a Bedrock version
	mockChangelog := `<li class="article-list-item ">
		<a href="/hc/en-us/articles/37810171798029-Minecraft-1-21-93-Bedrock" class="article-list-link" data-bi-id="n4a3" data-bi-name="minecraft - 1.21.93 (bedrock)" data-bi-type="text">Minecraft - 1.21.93 (Bedrock)</a>
	</li>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(mockChangelog))
	}))
	defer ts.Close()

	version, zipUrl, err := GetLatestBedrockVersion(ts.URL, "https://mocked-download-url/")
	if err != nil {
		t.Fatalf("Error fetching version: %v", err)
	}
	if version != "1.21.93" {
		t.Errorf("Expected version '1.21.93', got '%s'", version)
	}
	if !strings.Contains(zipUrl, "bedrock-server-1.21.93.zip") {
		t.Errorf("Expected zipUrl to contain version, got '%s'", zipUrl)
	}
}

func TestCheckForUpdate_NoNewVersion(t *testing.T) {
	current := "1.20.0.0"
	latest := "1.20.0.0"
	if IsNewVersionAvailable(current, latest) {
		t.Errorf("Expected no new version, but update was detected")
	}
}

func TestCheckForUpdate_NewVersionAvailable(t *testing.T) {
	current := "1.19.0.0"
	latest := "1.20.0.0"
	if !IsNewVersionAvailable(current, latest) {
		t.Errorf("Expected new version, but none was detected")
	}
}
