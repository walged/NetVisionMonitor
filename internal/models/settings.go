package models

type Settings struct {
	// UI Settings
	Theme string `json:"theme"` // "light" or "dark"

	// Sound Settings
	SoundEnabled       bool    `json:"sound_enabled"`        // Master switch for sounds
	SoundOnline        bool    `json:"sound_online"`         // Play sound when device goes online
	SoundOffline       bool    `json:"sound_offline"`        // Play sound when device goes offline
	SoundVolume        float64 `json:"sound_volume"`         // Volume 0.0 - 1.0

	// Monitoring Settings
	MonitoringInterval int `json:"monitoring_interval"` // seconds
	PingTimeout        int `json:"ping_timeout"`        // milliseconds
	SNMPTimeout        int `json:"snmp_timeout"`        // milliseconds
	RetryCount         int `json:"retry_count"`

	// Camera Settings
	UseFFmpeg  bool   `json:"use_ffmpeg"`
	StreamType string `json:"stream_type"` // "jpeg", "hls", "mjpeg"

	// Application Settings
	AutoStart      bool `json:"auto_start"`
	MinimizeToTray bool `json:"minimize_to_tray"`
}

func DefaultSettings() Settings {
	return Settings{
		Theme:              "dark",
		SoundEnabled:       true,
		SoundOnline:        true,
		SoundOffline:       true,
		SoundVolume:        0.5,
		MonitoringInterval: 30,
		PingTimeout:        3000,
		SNMPTimeout:        5000,
		RetryCount:         3,
		UseFFmpeg:          true,
		StreamType:         "jpeg",
		AutoStart:          false,
		MinimizeToTray:     true,
	}
}
