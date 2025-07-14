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

	dummyZip := createDummyZip(t)
	ts, cfg := createMockWikiServerAndConfig(t, tempDir, "1.21.93", "1.21.93.1", dummyZip)
	defer ts.Close()

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
	// Check that the zip file has the full patch version in its name
	expectedZip := filepath.Join(tempDir, "bedrock-server-1.21.93.1.zip")
	if _, err := os.Stat(expectedZip); os.IsNotExist(err) {
		t.Errorf("expected zip file %s to exist, but it does not", expectedZip)
	}
}

func TestUpdateFlow_NoUpdateNeeded(t *testing.T) {
	current := "1.20.0.0"

	tempDir := filepath.Join(os.TempDir(), "mc-server-test-no-update")
	defer os.RemoveAll(tempDir)

	dummyZip := createDummyZip(t)
	ts, cfg := createMockWikiServerAndConfig(t, tempDir, "1.20.0.0", "1.20.0.0", dummyZip)
	defer ts.Close()

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
	createWorldsDir(t, filepath.Join(tempDir, "Latest"))

	dummyZip := createDummyZip(t)
	ts, cfg := createMockWikiServerAndConfig(t, tempDir, "1.21.93", "1.21.93.1", dummyZip)
	defer ts.Close()

	dummySymlink := func(target, link string) error { return nil }
	updated, err := UpdateServerIfNew(current, cfg, dummySymlink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Errorf("expected update to occur, but it did not")
	}

	// Check that the worlds directory was copied into the new server directory before the symlink is updated
	extractDir := filepath.Join(tempDir, "bedrock-server-1.21.93.1")
	copiedWorlds := filepath.Join(extractDir, "worlds")
	if _, err := os.Stat(filepath.Join(copiedWorlds, "level.dat")); err != nil {
		t.Errorf("expected worlds directory and file to be copied, but got error: %v", err)
	}
}

func TestUpdateFlow_NoLatestSymlink_CopiesWorldsGracefully(t *testing.T) {
	current := "1.19.0.0"

	tempDir := filepath.Join(os.TempDir(), "mc-server-test-no-latest-symlink")
	defer os.RemoveAll(tempDir)

	// Create a dummy worlds directory in the old version directory (not in 'Latest')
	oldVersionDir := filepath.Join(tempDir, "bedrock-server-1.19.0.0")
	createWorldsDir(t, oldVersionDir)

	dummyZip := createDummyZip(t)
	ts, cfg := createMockWikiServerAndConfig(t, tempDir, "1.21.93", "1.21.93.1", dummyZip)
	defer ts.Close()

	dummySymlink := func(target, link string) error { return nil }
	updated, err := UpdateServerIfNew(current, cfg, dummySymlink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Errorf("expected update to occur, but it did not")
	}
}

func TestUpdateFlow_ServerShutdownAndCopyProps(t *testing.T) {
	// Arrange
	current := "1.21.92.1"
	latest := "1.21.93.1"
	tempDir := filepath.Join(os.TempDir(), "mc-server-test-shutdown")
	defer os.RemoveAll(tempDir)

	// Create dummy Latest symlink and worlds dir
	oldServerDir := filepath.Join(tempDir, "bedrock-server-"+current)
	createWorldsDir(t, oldServerDir)
	if err := os.WriteFile(filepath.Join(oldServerDir, "server.properties"), []byte("motd=old"), 0644); err != nil {
		t.Fatalf("failed to write server.properties: %v", err)
	}
	latestLink := filepath.Join(tempDir, "Latest")
	_ = os.Remove(latestLink)
	if err := os.Symlink(oldServerDir, latestLink); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Mock process check and shutdown logic
	processChecked := false
	shutdownMessageSent := false
	killed := false
	startCalled := false
	mockCheckProcess := func() bool { processChecked = true; return true }
	mockSendShutdown := func() { shutdownMessageSent = true }
	mockKill := func() { killed = true }
	mockStart := func() { startCalled = true }

	dummyZip := createDummyZip(t)
	ts, cfg := createMockWikiServerAndConfig(t, tempDir, "1.21.93", "1.21.93.1", dummyZip)
	defer ts.Close()


	dummySymlink := func(target, link string) error { _ = os.Remove(link); return os.Symlink(target, link) }

	// Act: call the update logic with injected mocks (assume UpdateServerIfNew accepts these for TDD)
	updated, err := UpdateServerIfNewWithProcessControl(current, cfg, dummySymlink, mockCheckProcess, mockSendShutdown, mockKill, mockStart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Errorf("expected update to occur, but it did not")
	}
	// Assert process control
	if !processChecked || !shutdownMessageSent || !killed || !startCalled {
		t.Errorf("expected process control steps to be called")
	}
	// Assert worlds and server.properties copied
	extractDir := filepath.Join(tempDir, "bedrock-server-"+latest)
	if _, err := os.Stat(filepath.Join(extractDir, "worlds", "level.dat")); err != nil {
		t.Errorf("expected worlds directory and file to be copied, but got error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(extractDir, "server.properties")); err != nil {
		t.Errorf("expected server.properties to be copied, but got error: %v", err)
	}
}

func createWorldsDir(t *testing.T, baseDir string) {
	t.Helper()
	srcWorlds := filepath.Join(baseDir, "worlds")
	if err := os.MkdirAll(srcWorlds, 0755); err != nil {
		t.Fatalf("failed to create worlds dir: %v", err)
	}
	testFile := filepath.Join(srcWorlds, "level.dat")
	if err := os.WriteFile(testFile, []byte("testdata"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}

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

func createMockWikiServerAndConfig(t *testing.T, tempDir, navVersion, zipVersion string, dummyZip []byte) (*httptest.Server, config.Config) {
	t.Helper()
	mockNav := `<li id="n-Latest:-` + navVersion + `" class="mw-list-item"><a href="/w/Bedrock_Edition_` + navVersion + `" title="Bedrock Edition ` + navVersion + `"><span>Latest: ` + navVersion + `</span></a></li>`
	var ts *httptest.Server
	ts = httptest.NewServer(nil)
	mockVersionPage := `<a href="` + ts.URL + `/bedrockdedicatedserver/bin-win/bedrock-server-` + zipVersion + `.zip">Windows</a>`
	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(200)
			w.Write([]byte(mockNav))
			return
		}
		if r.URL.Path == "/w/Bedrock_Edition_"+navVersion {
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
	cfg := config.Config{
		WikiNavURL: ts.URL + "/",
		ServerDir:  tempDir,
	}
	return ts, cfg
}
