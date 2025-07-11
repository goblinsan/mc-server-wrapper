package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/goblinsan/mc-server-wrapper/config"
)

func DownloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// UpdateServerIfNew checks for a new version and performs update steps if needed, always using config.
func UpdateServerIfNew(current, _ string, cfg config.Config) (bool, error) {
	// Ensure server directory exists
	if err := os.MkdirAll(cfg.ServerDir, os.ModePerm); err != nil {
		return false, fmt.Errorf("failed to create server dir: %w", err)
	}

	// Get latest version and zip URL
	version, zipUrl, err := GetLatestBedrockVersion(cfg.DownloadURL)
	if err != nil {
		return false, err
	}

	// Use the provided 'current' version for comparison
	if current == version {
		return false, nil // Already up to date, skip download/extract
	}

	// Download the zip to its versioned name
	zipName := fmt.Sprintf("bedrock-server-%s.zip", version)
	zipPath := filepath.Join(cfg.ServerDir, zipName)
	err = DownloadFile(zipPath, zipUrl)
	if err != nil {
		return false, fmt.Errorf("failed to download: %w", err)
	}

	// Extract the zip to a versioned directory
	extractDir := filepath.Join(cfg.ServerDir, "bedrock-server-"+version)
	if err := os.MkdirAll(extractDir, os.ModePerm); err != nil {
		return false, fmt.Errorf("failed to create extract dir: %w", err)
	}
	err = ExtractZip(zipPath, extractDir)
	if err != nil {
		return false, fmt.Errorf("failed to extract: %w", err)
	}

	// Update the 'Latest' symlink to point to the new version
	latestLink := filepath.Join(cfg.ServerDir, "Latest")
	os.Remove(latestLink) // Remove old symlink if exists
	if err := os.Symlink(extractDir, latestLink); err != nil {
		return false, fmt.Errorf("failed to update symlink: %w", err)
	}

	return true, nil
}

// ExtractZip extracts a zip file to the target directory
func ExtractZip(zipPath, targetDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(targetDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
