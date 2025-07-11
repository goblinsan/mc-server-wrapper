package updater

// IsNewVersionAvailable compares the current and latest version strings and returns true if a new version is available.
func IsNewVersionAvailable(current, latest string) bool {
	return current != latest
}
