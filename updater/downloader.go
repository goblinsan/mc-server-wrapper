package updater

import (
	"io"
	"net/http"
	"os"
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

// UpdateServerIfNew checks for a new version and performs update steps if needed.
func UpdateServerIfNew(current, latest string) (bool, error) {
	// TODO: implement download, extract, backup, copy, update latest dir
	if current == latest {
		return false, nil
	}
	return true, nil
}
