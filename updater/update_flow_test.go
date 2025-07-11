package updater

import (
	"testing"
)

func TestUpdateFlow_DownloadExtractBackupCopy(t *testing.T) {
	// Arrange
	current := "1.19.0.0"
	latest := "1.20.0.0"

	// Act
	updated, err := UpdateServerIfNew(current, latest)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Errorf("expected update to occur, but it did not")
	}
}

func TestUpdateFlow_NoUpdateNeeded(t *testing.T) {
	current := "1.20.0.0"
	latest := "1.20.0.0"

	updated, err := UpdateServerIfNew(current, latest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated {
		t.Errorf("expected no update, but update occurred")
	}
}
