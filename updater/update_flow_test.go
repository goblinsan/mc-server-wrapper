package updater

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/goblinsan/mc-server-wrapper/config"
)

func TestUpdateFlow_DownloadExtractBackupCopy(t *testing.T) {
	current := "1.19.0.0"
	latest := "1.20.0.0"

	// Create a temp dir for the test
	tempDir := filepath.Join(os.TempDir(), "mc-server-test")
	defer os.RemoveAll(tempDir)

	// Mock changelog and dummy zip
	dummyZip := createDummyZip(t)
	var ts *httptest.Server
	mockChangelog := `<li class="article-list-item ">
		<a href="/hc/en-us/articles/37810171798029-Minecraft-1-21-93-Bedrock" class="article-list-link" data-bi-id="n4a3" data-bi-name="minecraft - 1.21.93 (bedrock)" data-bi-type="text">Minecraft - 1.21.93 (Bedrock)</a>
	</li>`
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/changelog" {
			w.WriteHeader(200)
			w.Write([]byte(mockChangelog))
			return
		}
		if r.URL.Path == "/bedrockdedicatedserver/bin-win/bedrock-server-1.21.93.zip" {
			w.Header().Set("Content-Type", "application/zip")
			w.Write(dummyZip)
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	cfg := config.Config{
		ChangelogURL: ts.URL + "/changelog",
		DownloadURL:  ts.URL + "/bedrockdedicatedserver/bin-win/",
		ServerDir:    tempDir,
	}

	dummySymlink := func(target, link string) error { return nil }
	updated, err := UpdateServerIfNew(current, latest, cfg, dummySymlink)
	if err != nil {
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

	// Mock changelog and dummy zip for no update needed
	dummyZip := createDummyZip(t)
	var ts *httptest.Server
	mockChangelog := `<li class="article-list-item ">
		<a href="/hc/en-us/articles/37810171798029-Minecraft-1-20-0-0-Bedrock" class="article-list-link" data-bi-id="n4a3" data-bi-name="minecraft - 1.20.0.0 (bedrock)" data-bi-type="text">Minecraft - 1.20.0.0 (Bedrock)</a>
	</li>`
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/changelog" {
			w.WriteHeader(200)
			w.Write([]byte(mockChangelog))
			return
		}
		if r.URL.Path == "/bedrockdedicatedserver/bin-win/bedrock-server-1.20.0.0.zip" {
			w.Header().Set("Content-Type", "application/zip")
			w.Write(dummyZip)
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	cfg := config.Config{
		ChangelogURL: ts.URL + "/changelog",
		DownloadURL:  ts.URL + "/bedrockdedicatedserver/bin-win/",
		ServerDir:    tempDir,
	}

	dummySymlink := func(target, link string) error { return nil }
	updated, err := UpdateServerIfNew(current, latest, cfg, dummySymlink)
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

	// Mock changelog and dummy zip for the new version logic
	dummyZip := createDummyZip(t)
	var ts *httptest.Server
	mockChangelog := `<li class="article-list-item ">
		<a href="/hc/en-us/articles/37810171798029-Minecraft-1-21-93-Bedrock" class="article-list-link" data-bi-id="n4a3" data-bi-name="minecraft - 1.21.93 (bedrock)" data-bi-type="text">Minecraft - 1.21.93 (Bedrock)</a>
	</li>`
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/changelog" {
			w.WriteHeader(200)
			w.Write([]byte(mockChangelog))
			return
		}
		if r.URL.Path == "/bedrockdedicatedserver/bin-win/bedrock-server-1.21.93.zip" {
			w.Header().Set("Content-Type", "application/zip")
			w.Write(dummyZip)
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	cfg := config.Config{
		ChangelogURL: ts.URL + "/changelog",
		DownloadURL:  ts.URL + "/bedrockdedicatedserver/bin-win/",
		ServerDir:    tempDir,
	}

	dummySymlink := func(target, link string) error { return nil }
	updated, err := UpdateServerIfNew(current, "", cfg, dummySymlink)
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
