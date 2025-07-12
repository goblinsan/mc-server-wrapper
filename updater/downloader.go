package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/goblinsan/mc-server-wrapper/config"
)

func DownloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

type SymlinkUpdater func(target, link string) error

func DefaultSymlinkUpdater(target, link string) error {
	os.Remove(link) // Remove old symlink if exists
	return os.Symlink(target, link)
}

func UpdateServerIfNew(current string, cfg config.Config, symlinkUpdater SymlinkUpdater) (bool, error) {
	// Ensure server directory exists
	if err := os.MkdirAll(cfg.ServerDir, os.ModePerm); err != nil {
		return false, fmt.Errorf("failed to create server dir: %w", err)
	}

	version, zipUrl, err := GetLatestBedrockVersion(cfg.WikiNavURL)
	if err != nil {
		return false, err
	}

	if current == version {
		return false, nil
	}

	zipFileName := filepath.Base(zipUrl)
	zipPath := filepath.Join(cfg.ServerDir, zipFileName)
	err = DownloadFile(zipPath, zipUrl)
	if err != nil {
		return false, fmt.Errorf("failed to download: %w", err)
	}

	extractDir := filepath.Join(cfg.ServerDir, "bedrock-server-"+ParseBedrockVersion(zipFileName))
	if err := os.MkdirAll(extractDir, os.ModePerm); err != nil {
		return false, fmt.Errorf("failed to create extract dir: %w", err)
	}
	err = ExtractZip(zipPath, extractDir)
	if err != nil {
		return false, fmt.Errorf("failed to extract: %w", err)
	}

	dstWorlds := filepath.Join(extractDir, "worlds")
	srcWorlds := filepath.Join(cfg.ServerDir, "Latest", "worlds")
	if _, err := os.Stat(srcWorlds); os.IsNotExist(err) {
		// Fallback: find the most recent bedrock-server-* directory (excluding the new one)
		entries, _ := os.ReadDir(cfg.ServerDir)
		var latestDir string
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "bedrock-server-") && entry.Name() != "bedrock-server-"+ParseBedrockVersion(zipFileName) {
				candidate := filepath.Join(cfg.ServerDir, entry.Name(), "worlds")
				if _, err := os.Stat(candidate); err == nil {
					latestDir = candidate
				}
			}
		}
		if latestDir != "" {
			srcWorlds = latestDir
		}
	}

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
		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			return err
		}
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
