package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"netvisionmonitor/internal/autostart"
	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/models"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AppSettings holds all application settings
type AppSettings struct {
	// Theme settings
	Theme string `json:"theme"` // "dark", "light", "system"

	// Monitoring settings
	MonitoringInterval int  `json:"monitoring_interval"` // seconds
	PingTimeout        int  `json:"ping_timeout"`        // seconds
	SNMPTimeout        int  `json:"snmp_timeout"`        // seconds
	MonitoringWorkers  int  `json:"monitoring_workers"`
	AutoStartMonitor   bool `json:"auto_start_monitor"`

	// Notification settings
	SoundEnabled       bool    `json:"sound_enabled"`
	SoundVolume        float64 `json:"sound_volume"` // 0.0 - 1.0
	NotifyOnOffline    bool    `json:"notify_on_offline"`
	NotifyOnOnline     bool    `json:"notify_on_online"`
	NotifyOnPortChange bool    `json:"notify_on_port_change"`

	// Data settings
	EventRetentionDays int `json:"event_retention_days"`

	// Camera settings
	CameraSnapshotInterval int    `json:"camera_snapshot_interval"` // seconds
	CameraStreamType       string `json:"camera_stream_type"`       // "jpeg", "mjpeg", "hls"

	// System settings
	MinimizeToTray bool `json:"minimize_to_tray"` // Minimize to tray on close
}

// DefaultAppSettings returns default settings
func DefaultAppSettings() AppSettings {
	return AppSettings{
		Theme:                  "dark",
		MonitoringInterval:     30,
		PingTimeout:            3,
		SNMPTimeout:            5,
		MonitoringWorkers:      10,
		AutoStartMonitor:       true,
		SoundEnabled:           true,
		SoundVolume:            0.5,
		NotifyOnOffline:        true,
		NotifyOnOnline:         true,
		NotifyOnPortChange:     false,
		EventRetentionDays:     30,
		CameraSnapshotInterval: 60,
		CameraStreamType:       "jpeg",
		MinimizeToTray:         true,
	}
}

// GetAppSettings returns current application settings
func (a *App) GetAppSettings() (AppSettings, error) {
	if a.db == nil {
		return DefaultAppSettings(), nil
	}

	repo := database.NewSettingsRepository(a.db.DB())
	settings := DefaultAppSettings()

	if err := repo.GetJSON("app_settings", &settings); err != nil {
		return settings, nil // Return defaults on error
	}

	return settings, nil
}

// SaveAppSettings saves application settings
func (a *App) SaveAppSettings(settings AppSettings) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	repo := database.NewSettingsRepository(a.db.DB())
	if err := repo.SetJSON("app_settings", settings); err != nil {
		return err
	}

	// Apply monitoring settings if monitor is running
	if a.monitor != nil {
		a.monitor.SetInterval(time.Duration(settings.MonitoringInterval) * time.Second)
	}

	// Emit settings changed event
	runtime.EventsEmit(a.ctx, "settings:changed", settings)

	return nil
}

// ExportData exports all data to a ZIP file
func (a *App) ExportData() (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Ask user for save location
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Экспорт данных",
		DefaultFilename: fmt.Sprintf("netvision_backup_%s.zip", time.Now().Format("2006-01-02")),
		Filters: []runtime.FileFilter{
			{DisplayName: "ZIP Archive", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return "", err
	}
	if savePath == "" {
		return "", nil // User cancelled
	}

	// Create ZIP file
	zipFile, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Export database
	dbPath := filepath.Join(a.cfg.DataDir, "netvision.db")
	if err := addFileToZip(zipWriter, dbPath, "netvision.db"); err != nil {
		return "", fmt.Errorf("failed to add database to zip: %w", err)
	}

	// Export config
	configData, _ := json.MarshalIndent(map[string]interface{}{
		"version":     "1.0.0",
		"export_date": time.Now().Format(time.RFC3339),
		"is_portable": a.cfg.IsPortable,
	}, "", "  ")

	configWriter, err := zipWriter.Create("export_info.json")
	if err != nil {
		return "", err
	}
	configWriter.Write(configData)

	// Export schema backgrounds
	schemasDir := filepath.Join(a.cfg.DataDir, "schemas")
	if _, err := os.Stat(schemasDir); err == nil {
		filepath.Walk(schemasDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			relPath, _ := filepath.Rel(a.cfg.DataDir, path)
			return addFileToZip(zipWriter, path, relPath)
		})
	}

	return savePath, nil
}

// ImportData imports data from a ZIP file
func (a *App) ImportData() (bool, error) {
	if a.db == nil {
		return false, fmt.Errorf("database not initialized")
	}

	// Ask user for file
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Импорт данных",
		Filters: []runtime.FileFilter{
			{DisplayName: "ZIP Archive", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return false, err
	}
	if filePath == "" {
		return false, nil // User cancelled
	}

	// Confirm import
	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Подтверждение импорта",
		Message:       "Текущие данные будут заменены. Продолжить?",
		Buttons:       []string{"Да", "Нет"},
		DefaultButton: "Нет",
	})
	if err != nil || result != "Да" {
		return false, nil
	}

	// Stop monitoring
	if a.monitor != nil {
		a.monitor.Stop()
	}

	// Close database
	a.db.Close()

	// Open ZIP file
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipReader.Close()

	// Extract files
	for _, file := range zipReader.File {
		destPath := filepath.Join(a.cfg.DataDir, file.Name)

		// Create directory if needed
		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		// Create parent directory
		os.MkdirAll(filepath.Dir(destPath), 0755)

		// Extract file
		srcFile, err := file.Open()
		if err != nil {
			continue
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			continue
		}

		io.Copy(destFile, srcFile)
		destFile.Close()
		srcFile.Close()
	}

	// Reinitialize database
	db, err := database.Initialize(a.cfg.DataDir)
	if err != nil {
		return false, fmt.Errorf("failed to reinitialize database: %w", err)
	}
	a.db = db

	// Restart monitoring
	a.initMonitoring()
	a.monitor.Start()

	return true, nil
}

// GetDataPath returns the data directory path
func (a *App) GetDataPath() string {
	if a.cfg != nil {
		return a.cfg.DataDir
	}
	return ""
}

// OpenDataFolder opens the data folder in file explorer
func (a *App) OpenDataFolder() error {
	if a.cfg == nil {
		return fmt.Errorf("config not initialized")
	}
	runtime.BrowserOpenURL(a.ctx, "file:///"+filepath.ToSlash(a.cfg.DataDir))
	return nil
}

// ClearOldData removes old events and cache
func (a *App) ClearOldData(daysToKeep int) (int64, error) {
	if a.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	if daysToKeep <= 0 {
		daysToKeep = 30
	}

	repo := database.NewEventRepository(a.db.DB())
	before := time.Now().AddDate(0, 0, -daysToKeep)
	return repo.DeleteOlderThan(before)
}

// Helper function to add file to ZIP
func addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipPath
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// GetSettings returns current settings (legacy compatibility)
func (a *App) GetSettings() models.Settings {
	settings, _ := a.GetAppSettings()
	return models.Settings{
		Theme:              settings.Theme,
		MonitoringInterval: settings.MonitoringInterval,
		PingTimeout:        settings.PingTimeout * 1000, // Convert to ms
		SNMPTimeout:        settings.SNMPTimeout * 1000, // Convert to ms
		StreamType:         settings.CameraStreamType,
	}
}

// SaveSettings saves settings (legacy compatibility)
func (a *App) SaveSettings(settings models.Settings) error {
	appSettings, _ := a.GetAppSettings()
	appSettings.Theme = settings.Theme
	appSettings.MonitoringInterval = settings.MonitoringInterval
	if settings.PingTimeout > 0 {
		appSettings.PingTimeout = settings.PingTimeout / 1000 // Convert to seconds
	}
	if settings.SNMPTimeout > 0 {
		appSettings.SNMPTimeout = settings.SNMPTimeout / 1000 // Convert to seconds
	}
	appSettings.CameraStreamType = settings.StreamType
	return a.SaveAppSettings(appSettings)
}

// GetAutostartEnabled checks if autostart is enabled
func (a *App) GetAutostartEnabled() bool {
	enabled, _ := autostart.IsEnabled()
	return enabled
}

// SetAutostartEnabled enables or disables autostart
func (a *App) SetAutostartEnabled(enabled bool) error {
	if enabled {
		return autostart.Enable()
	}
	return autostart.Disable()
}

// ToggleAutostart toggles autostart state and returns new state
func (a *App) ToggleAutostart() (bool, error) {
	return autostart.Toggle()
}
