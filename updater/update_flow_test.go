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

	// Create a temp dir for the test
	tempDir := filepath.Join(os.TempDir(), "mc-server-test")
	defer os.RemoveAll(tempDir)

	// Mock wiki nav, version page, and dummy zip
	dummyZip := createDummyZip(t)
	var ts *httptest.Server
	mockNav := `<li id="n-Latest:-1.21.93" class="mw-list-item"><a href="/w/Bedrock_Edition_1.21.93" title="Bedrock Edition 1.21.93"><span>Latest: 1.21.93</span></a></li>`
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
	   if strings.HasPrefix(r.URL.Path, "/bedrockdedicatedserver/bin-win/bedrock-server-") && strings.HasSuffix(r.URL.Path, ".zip") {
		   w.Header().Set("Content-Type", "application/zip")
		   w.Write(dummyZip)
		   return
	   }
		w.WriteHeader(404)
	})
	defer ts.Close()

	cfg := config.Config{
		WikiNavURL: ts.URL + "/",
		ServerDir:  tempDir,
	}

	dummySymlink := func(target, link string) error { return nil }
	updated, err := UpdateServerIfNew(current, cfg, dummySymlink)
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

	tempDir := filepath.Join(os.TempDir(), "mc-server-test-no-update")
	defer os.RemoveAll(tempDir)

	// Mock wiki nav, version page, and dummy zip for no update needed
	dummyZip := createDummyZip(t)
	var ts *httptest.Server
	mockNav := `<li id="n-Latest:-1.20.0.0" class="mw-list-item"><a href="/w/Bedrock_Edition_1.20.0.0" title="Bedrock Edition 1.20.0.0"><span>Latest: 1.20.0.0</span></a></li>`
	ts = httptest.NewServer(nil)
	mockVersionPage := `<a href="` + ts.URL + `/bedrockdedicatedserver/bin-win/bedrock-server-1.20.0.0.zip">Windows</a>`
	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(200)
			w.Write([]byte(mockNav))
			return
		}
		if r.URL.Path == "/w/Bedrock_Edition_1.20.0.0" {
			w.WriteHeader(200)
			w.Write([]byte(mockVersionPage))
			return
		}
	   if strings.HasPrefix(r.URL.Path, "/bedrockdedicatedserver/bin-win/bedrock-server-") && strings.HasSuffix(r.URL.Path, ".zip") {
		   w.Header().Set("Content-Type", "application/zip")
		   w.Write(dummyZip)
		   return
	   }
		w.WriteHeader(404)
	})
	defer ts.Close()

	cfg := config.Config{
		WikiNavURL: ts.URL + "/",
		ServerDir:  tempDir,
	}

	dummySymlink := func(target, link string) error { return nil }
	updated, err := UpdateServerIfNew(current, cfg, dummySymlink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated {
		t.Errorf("expected no update, but update occurred")
	}
}

func TestUpdateFlow_CopiesWorldsDirectory(t *testing.T) {
	current := "1.19.0.0"

	tempDir := filepath.Join(os.TempDir(), "mc-server-test-worlds-copy")
	defer os.RemoveAll(tempDir)

	// Create a dummy worlds directory with a file
	srcWorlds := filepath.Join(tempDir, "Latest", "worlds")
	if err := os.MkdirAll(srcWorlds, 0755); err != nil {
		t.Fatalf("failed to create worlds dir: %v", err)
	}
	testFile := filepath.Join(srcWorlds, "level.dat")
	if err := os.WriteFile(testFile, []byte("testdata"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Mock wiki nav, version page, and dummy zip for the new version logic
	dummyZip := createDummyZip(t)
	var ts *httptest.Server
	mockNav := `<li id="n-Latest:-1.21.93" class="mw-list-item"><a href="/w/Bedrock_Edition_1.21.93" title="Bedrock Edition 1.21.93"><span>Latest: 1.21.93</span></a></li>`
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
	   if strings.HasPrefix(r.URL.Path, "/bedrockdedicatedserver/bin-win/bedrock-server-") && strings.HasSuffix(r.URL.Path, ".zip") {
		   w.Header().Set("Content-Type", "application/zip")
		   w.Write(dummyZip)
		   return
	   }
		w.WriteHeader(404)
	})
	defer ts.Close()

	cfg := config.Config{
		WikiNavURL: ts.URL + "/",
		ServerDir:  tempDir,
	}

	dummySymlink := func(target, link string) error { return nil }
	updated, err := UpdateServerIfNew(current, cfg, dummySymlink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Errorf("expected update to occur, but it did not")
	}

	// Check that the worlds directory was copied into the new server directory
	extractDir := filepath.Join(tempDir, "bedrock-server-1.21.93")
	copiedWorlds := filepath.Join(extractDir, "worlds")
	if _, err := os.Stat(filepath.Join(copiedWorlds, "level.dat")); err != nil {
		t.Errorf("expected worlds directory and file to be copied, but got error: %v", err)
	}
}
