package updater

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetLatestBedrockVersion(t *testing.T) {
	// Mock minecraft.wiki nav HTML and version page HTML
	mockNav := `<li id="n-Latest:-1.21.93" class="mw-list-item"><a href="/w/Bedrock_Edition_1.21.93" title="Bedrock Edition 1.21.93"><span>Latest: 1.21.93</span></a></li>`
	dummyZip := []byte("PK\x03\x04dummyzipcontent")
	var ts *httptest.Server
	ts = httptest.NewServer(nil)
	mockVersionPage := `<a href="` + ts.URL + `/bedrockdedicatedserver/bin-win/bedrock-server-1.21.93.1.zip">Windows</a>`
	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(200)
			w.Write([]byte(mockNav))
			return
		}
		if r.URL.Path == "/w/Bedrock_Edition_1.21.93" {
			w.WriteHeader(200)
			w.Write([]byte(mockVersionPage))
			return
		}
		if r.URL.Path == "/bedrockdedicatedserver/bin-win/bedrock-server-1.21.93.1.zip" {
			w.Header().Set("Content-Type", "application/zip")
			w.Write(dummyZip)
			return
		}
		w.WriteHeader(404)
	})
	defer ts.Close()

	version, zipUrl, err := GetLatestBedrockVersion(ts.URL + "/")
	if err != nil {
		t.Fatalf("Error fetching version: %v", err)
	}
	if version != "1.21.93" {
		t.Errorf("Expected version '1.21.93', got '%s'", version)
	}
	if !strings.Contains(zipUrl, "bedrock-server-1.21.93.1.zip") {
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
