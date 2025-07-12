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
// SymlinkUpdater is a function type for updating the 'Latest' symlink.
type SymlinkUpdater func(target, link string) error

// DefaultSymlinkUpdater uses os.Symlink
func DefaultSymlinkUpdater(target, link string) error {
	os.Remove(link) // Remove old symlink if exists
	return os.Symlink(target, link)
}

// UpdateServerIfNew checks for a new version and performs update steps if needed, always using config.
// Accepts a symlinkUpdater for testability.
func UpdateServerIfNew(current string, cfg config.Config, symlinkUpdater SymlinkUpdater) (bool, error) {
	// Ensure server directory exists
	if err := os.MkdirAll(cfg.ServerDir, os.ModePerm); err != nil {
		return false, fmt.Errorf("failed to create server dir: %w", err)
	}

	// Get latest version and zip URL
	version, zipUrl, err := GetLatestBedrockVersion(cfg.WikiNavURL)
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

	// Copy the 'worlds' directory from the current server to the new extracted server
	srcWorlds := filepath.Join(cfg.ServerDir, "Latest", "worlds")
	dstWorlds := filepath.Join(extractDir, "worlds")
	if _, err := os.Stat(srcWorlds); err == nil {
		if err := CopyDir(srcWorlds, dstWorlds); err != nil {
			return false, fmt.Errorf("failed to copy worlds: %w", err)
		}
	}

	// Update the 'Latest' symlink to point to the new version
	latestLink := filepath.Join(cfg.ServerDir, "Latest")
	if err := symlinkUpdater(extractDir, latestLink); err != nil {
		return false, fmt.Errorf("failed to update symlink: %w", err)
	}
	return true, nil
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
func CopyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	})
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
