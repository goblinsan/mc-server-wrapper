package updater

import (
	"testing"
)

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
