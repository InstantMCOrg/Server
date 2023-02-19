package models

type PreparedContainer struct {
	Number    int    `json:"number"`
	McVersion string `json:"mc_version"`
	RamSizeMB int    `json:"ram_size_mb"`
}
