package config

type Config struct {
	ServerDir       string `json:"server_dir"`
	BackupDir       string `json:"backup_dir"`
	WorldDir        string `json:"world_dir"`
	NetworkShare    string `json:"network_share"`
	DownloadURL     string `json:"download_url"`
	LastVersionFile string `json:"last_version_file"`
}
