package onvif

import "strings"

// CameraPreset contains manufacturer-specific settings for cameras
type CameraPreset struct {
	Manufacturer string   `json:"manufacturer"`
	Series       string   `json:"series,omitempty"`
	Description  string   `json:"description,omitempty"`
	ONVIFPort    int      `json:"onvif_port"`
	RTSPPort     int      `json:"rtsp_port"`
	SnapshotURLs []string `json:"snapshot_urls"` // Templates with {ip}, {port}, {user}, {pass}, {channel}
	RTSPURLs     []string `json:"rtsp_urls"`     // Templates with {ip}, {port}, {user}, {pass}, {channel}
	MediaPaths   []string `json:"media_paths"`   // ONVIF media service paths
}

// ManufacturerPresets contains all presets for a manufacturer
type ManufacturerPresets struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Presets     []CameraPreset `json:"presets"`
}

// GetAllManufacturers returns list of supported manufacturers with presets
func GetAllManufacturers() []ManufacturerPresets {
	return []ManufacturerPresets{
		getLTVPresets(),
		getHikvisionPresets(),
		getDahuaPresets(),
		getGenericPresets(),
	}
}

// GetManufacturerPresets returns presets for a specific manufacturer
func GetManufacturerPresets(manufacturer string) *ManufacturerPresets {
	manufacturer = strings.ToLower(manufacturer)
	for _, m := range GetAllManufacturers() {
		if strings.ToLower(m.Name) == manufacturer || strings.Contains(strings.ToLower(m.DisplayName), manufacturer) {
			return &m
		}
	}
	return nil
}

// GetPresetForSeries returns preset for a specific manufacturer and series
func GetPresetForSeries(manufacturer, series string) *CameraPreset {
	presets := GetManufacturerPresets(manufacturer)
	if presets == nil {
		return nil
	}

	series = strings.ToLower(series)
	for _, p := range presets.Presets {
		if strings.ToLower(p.Series) == series || strings.Contains(strings.ToLower(p.Description), series) {
			return &p
		}
	}

	// Return first preset as default
	if len(presets.Presets) > 0 {
		return &presets.Presets[0]
	}
	return nil
}

// LTV camera presets based on official documentation
func getLTVPresets() ManufacturerPresets {
	return ManufacturerPresets{
		Name:        "LTV",
		DisplayName: "LTV (ЛТВ)",
		Presets: []CameraPreset{
			{
				Manufacturer: "LTV",
				Series:       "M-series",
				Description:  "LTV-CNM, LTV-ICDM и др. (M-серия)",
				ONVIFPort:    80,
				RTSPPort:     554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/Streaming/channels/1/picture",
					"http://{ip}:{port}/Streaming/channels/{channel}/picture",
					"http://{ip}:{port}/ISAPI/Streaming/channels/1/picture",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/h264/ch01/main",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/h264/ch01/sub",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/Streaming/Channels/101",
				},
				MediaPaths: []string{
					"/onvif/media_service",
					"/onvif/Media",
				},
			},
			{
				Manufacturer: "LTV",
				Series:       "E-series",
				Description:  "LTV-CNE и др. (E-серия)",
				ONVIFPort:    80,
				RTSPPort:     554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/GetSnapshot/1",
					"http://{ip}:{port}/GetSnapshot/{channel}",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/live/main",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/live/second",
				},
				MediaPaths: []string{
					"/onvif/media",
					"/onvif/Media",
					"/onvif/media_service",
				},
			},
			{
				Manufacturer: "LTV",
				Series:       "T-series",
				Description:  "LTV T-серия (ONVIF)",
				ONVIFPort:    80,
				RTSPPort:     554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/onvif/snapshot.cgi",
					"http://{ip}:{port}/onvif-http/snapshot",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/live/main",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/live/sub",
				},
				MediaPaths: []string{
					"/onvif/media_service",
					"/onvif/Media",
				},
			},
			{
				Manufacturer: "LTV",
				Series:       "5-series",
				Description:  "LTV 5-я серия",
				ONVIFPort:    80,
				RTSPPort:     554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/cgi-bin/snapshot.cgi",
					"http://{ip}:{port}/cgi-bin/snapshot.cgi?channel=1",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/live/main",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/live/second",
				},
				MediaPaths: []string{
					"/onvif/media_service",
					"/onvif/Media",
				},
			},
			{
				Manufacturer: "LTV",
				Series:       "1-3-series",
				Description:  "LTV 1-я и 3-я серия",
				ONVIFPort:    80,
				RTSPPort:     554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/images/snapshot.jpg",
					"http://{ip}:{port}/LAPI/V1.0/Channels/1/Media/Video/Streams/0/Snapshot",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/0",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/1",
				},
				MediaPaths: []string{
					"/onvif/media_service",
					"/onvif/Media",
				},
			},
			{
				Manufacturer: "LTV",
				Series:       "Pro-series",
				Description:  "LTV Pro-серия",
				ONVIFPort:    80,
				RTSPPort:     554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/Streaming/channels/1/picture",
					"http://{ip}:{port}/cgi-bin/snapshot.cgi",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/live/main",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/live/second",
				},
				MediaPaths: []string{
					"/onvif/media_service",
					"/onvif/Media",
				},
			},
			{
				Manufacturer: "LTV",
				Series:       "L-series",
				Description:  "LTV L-серия",
				ONVIFPort:    80,
				RTSPPort:     8554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/Streaming/channels/1/picture",
					"http://{ip}:{port}/cgi-bin/snapshot.cgi",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:8554/live/main",
					"rtsp://{user}:{pass}@{ip}:8555/live/second",
				},
				MediaPaths: []string{
					"/onvif/media_service",
					"/onvif/Media",
				},
			},
		},
	}
}

// Hikvision camera presets
func getHikvisionPresets() ManufacturerPresets {
	return ManufacturerPresets{
		Name:        "Hikvision",
		DisplayName: "Hikvision",
		Presets: []CameraPreset{
			{
				Manufacturer: "Hikvision",
				Series:       "default",
				Description:  "Стандартные камеры Hikvision",
				ONVIFPort:    80,
				RTSPPort:     554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/Streaming/channels/1/picture",
					"http://{ip}:{port}/ISAPI/Streaming/channels/101/picture",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/Streaming/Channels/101",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/Streaming/Channels/102",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/h264/ch1/main/av_stream",
				},
				MediaPaths: []string{
					"/onvif/media_service",
					"/onvif/Media",
				},
			},
		},
	}
}

// Dahua camera presets
func getDahuaPresets() ManufacturerPresets {
	return ManufacturerPresets{
		Name:        "Dahua",
		DisplayName: "Dahua",
		Presets: []CameraPreset{
			{
				Manufacturer: "Dahua",
				Series:       "default",
				Description:  "Стандартные камеры Dahua",
				ONVIFPort:    80,
				RTSPPort:     554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/cgi-bin/snapshot.cgi",
					"http://{ip}:{port}/cgi-bin/snapshot.cgi?channel=1",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/cam/realmonitor?channel=1&subtype=0",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/cam/realmonitor?channel=1&subtype=1",
				},
				MediaPaths: []string{
					"/onvif/media_service",
					"/onvif/Media",
				},
			},
		},
	}
}

// Generic camera presets for unknown manufacturers
func getGenericPresets() ManufacturerPresets {
	return ManufacturerPresets{
		Name:        "Generic",
		DisplayName: "Другие / Универсальные",
		Presets: []CameraPreset{
			{
				Manufacturer: "Generic",
				Series:       "default",
				Description:  "Универсальные настройки ONVIF",
				ONVIFPort:    80,
				RTSPPort:     554,
				SnapshotURLs: []string{
					"http://{ip}:{port}/onvif-http/snapshot",
					"http://{ip}:{port}/cgi-bin/snapshot.cgi",
					"http://{ip}:{port}/Streaming/channels/1/picture",
					"http://{ip}:{port}/snap.jpg",
					"http://{ip}:{port}/snapshot.jpg",
				},
				RTSPURLs: []string{
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/live/main",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/stream1",
					"rtsp://{user}:{pass}@{ip}:{rtsp_port}/h264",
				},
				MediaPaths: []string{
					"/onvif/media_service",
					"/onvif/Media",
					"/onvif/media",
					"/onvif/services/media",
					"/Media",
					"/media",
				},
			},
		},
	}
}

// ApplyTemplate replaces placeholders in URL template
func ApplyTemplate(template, ip string, port, rtspPort int, username, password string, channel int) string {
	result := template

	// Replace placeholders
	result = strings.ReplaceAll(result, "{ip}", ip)
	result = strings.ReplaceAll(result, "{port}", intToString(port))
	result = strings.ReplaceAll(result, "{rtsp_port}", intToString(rtspPort))
	result = strings.ReplaceAll(result, "{user}", username)
	result = strings.ReplaceAll(result, "{pass}", password)
	result = strings.ReplaceAll(result, "{channel}", intToString(channel))

	return result
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	negative := false
	if n < 0 {
		negative = true
		n = -n
	}
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	if negative {
		result = "-" + result
	}
	return result
}
