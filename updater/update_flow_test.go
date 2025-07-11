package updater

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goblinsan/mc-server-wrapper/config"
)

func TestUpdateFlow_DownloadExtractBackupCopy(t *testing.T) {
	current := "1.19.0.0"
	latest := "1.20.0.0"

	// Create a temp dir for the test
	tempDir := filepath.Join(os.TempDir(), "mc-server-test")
	defer os.RemoveAll(tempDir)

	// Mock Minecraft download page and dummy zip
	dummyZip := createDummyZip(t)
	var ts *httptest.Server
	mockHTML := `<a href="` + "REPLACE_TS_URL" + `/bedrockdedicatedserver/bin-win/bedrock-server-1.21.93.1.zip" class="MC_Button MC_Button_Hero_Outline MC_Glyph_Download_A MC_Style_Core_Green_5" aria-label="serverBedrockWindows" data-aem-contentname="primary-cta" id="MC_Download_Server_1" target="_blank" data-bi-id="MC_Download_Server_1" data-bi-ct="button" data-bi-cn="primary-cta" data-bi-ecn="primary-cta" disabled="disabled" data-bi-bhvr="DOWNLOAD"><span>Download</span></a>`
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bedrockdedicatedserver/bin-win/" {
			w.WriteHeader(200)
			w.Write([]byte(strings.ReplaceAll(mockHTML, "REPLACE_TS_URL", ts.URL)))
			return
		}
		if r.URL.Path == "/bedrockdedicatedserver/bin-win/bedrock-server-1.21.93.1.zip" {
			w.Header().Set("Content-Type", "application/zip")
			w.Write(dummyZip)
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	cfg := config.Config{
		DownloadURL: ts.URL + "/bedrockdedicatedserver/bin-win/",
		ServerDir:   tempDir,
	}

	updated, err := UpdateServerIfNew(current, latest, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Errorf("expected update to occur, but it did not")
	}
	// Check that the directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("expected server dir to exist, but it does not")
	}
}

// createDummyZip returns a minimal valid zip file as []byte
func createDummyZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("dummy.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	_, err = f.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	return buf.Bytes()
}

func TestUpdateFlow_NoUpdateNeeded(t *testing.T) {
	current := "1.20.0.0"
	latest := "1.20.0.0"

	tempDir := filepath.Join(os.TempDir(), "mc-server-test-no-update")
	defer os.RemoveAll(tempDir)

	// Mock Minecraft download page with a zip link for the current version
	var ts *httptest.Server
	mockHTMLTemplate := `<a href="{{ZIPURL}}" class="MC_Button MC_Button_Hero_Outline MC_Glyph_Download_A MC_Style_Core_Green_5" aria-label="serverBedrockWindows" id="MC_Download_Server_1"><span>Download</span></a>`
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bedrockdedicatedserver/bin-win/" {
			mockZipUrl := ts.URL + "/bedrockdedicatedserver/bin-win/bedrock-server-1.20.0.0.zip"
			mockHTML := strings.ReplaceAll(mockHTMLTemplate, "{{ZIPURL}}", mockZipUrl)
			w.WriteHeader(200)
			w.Write([]byte(mockHTML))
			return
		}
		if r.URL.Path == "/bedrockdedicatedserver/bin-win/bedrock-server-1.20.0.0.zip" {
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(200)
			w.Write([]byte("PK\x03\x04dummyzipcontent"))
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	cfg := config.Config{
		DownloadURL: ts.URL + "/bedrockdedicatedserver/bin-win/",
		ServerDir:   tempDir,
	}

	updated, err := UpdateServerIfNew(current, latest, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated {
		t.Errorf("expected no update, but update occurred")
	}
}
