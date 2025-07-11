package updater

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetLatestBedrockVersion(t *testing.T) {
	// Mock HTML from the current Minecraft Bedrock Server download page
	mockHTML := `
	<a href="https://www.minecraft.net/bedrockdedicatedserver/bin-win/bedrock-server-1.21.93.1.zip" class="MC_Button MC_Button_Hero_Outline MC_Glyph_Download_A MC_Style_Core_Green_5" aria-label="serverBedrockWindows" data-aem-contentname="primary-cta" id="MC_Download_Server_1" target="_blank" data-bi-id="MC_Download_Server_1" data-bi-ct="button" data-bi-cn="primary-cta" data-bi-ecn="primary-cta" disabled="disabled" data-bi-bhvr="DOWNLOAD">
	<span>Download</span>
	</a>
	<a href="https://www.minecraft.net/bedrockdedicatedserver/bin-linux/bedrock-server-1.21.93.1.zip" class="MC_Button MC_Button_Hero_Outline MC_Glyph_Download_A MC_Style_Core_Green_5" aria-label="serverBedrockLinux" data-aem-contentname="primary-cta" id="MC_Download_Server_2" target="_blank" data-bi-id="MC_Download_Server_2" data-bi-ct="button" data-bi-cn="primary-cta" data-bi-ecn="primary-cta" disabled="disabled" data-bi-bhvr="DOWNLOAD">
	<span>Download</span>
	</a>
	`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(mockHTML))
	}))
	defer ts.Close()

	version, zipUrl, err := GetLatestBedrockVersion(ts.URL + "/bedrockdedicatedserver/bin-win/")
	if err != nil {
		t.Fatalf("Error fetching version: %v", err)
	}
	if version != "1.21.93.1" {
		t.Errorf("Expected version '1.21.93.1', got '%s'", version)
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
