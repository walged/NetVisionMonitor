package main

import (
	"os"
	"strings"

	"netvisionmonitor/internal/logger"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// MinimizeToTray minimizes the window to system tray
func (a *App) MinimizeToTray() {
	wailsRuntime.WindowHide(a.ctx)
	logger.Info("Window minimized to tray")
}

// ShowFromTray shows the window from system tray
func (a *App) ShowFromTray() {
	wailsRuntime.WindowShow(a.ctx)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
	logger.Info("Window restored from tray")
}

// QuitApp completely exits the application
func (a *App) QuitApp() {
	logger.Info("Quit requested from tray")
	wailsRuntime.Quit(a.ctx)
}

// GetTrayStatus returns current monitoring status for tray tooltip
func (a *App) GetTrayStatus() map[string]interface{} {
	status := map[string]interface{}{
		"monitoring": false,
		"online":     0,
		"offline":    0,
		"total":      0,
	}

	if a.monitor != nil {
		monStatus := a.GetMonitoringStatus()
		status["monitoring"] = monStatus.Running

		// Count online/offline devices
		if a.db != nil {
			devices, err := a.GetDevices()
			if err == nil {
				online := 0
				offline := 0
				for _, d := range devices {
					if d.Status == "online" {
						online++
					} else if d.Status == "offline" {
						offline++
					}
				}
				status["online"] = online
				status["offline"] = offline
				status["total"] = len(devices)
			}
		}
	}

	return status
}

// ShouldMinimizeToTray checks if app should minimize to tray on close
func (a *App) ShouldMinimizeToTray() bool {
	if a.db == nil {
		return true // Default behavior
	}

	settings := a.GetSettings()
	return settings.MinimizeToTray
}

// GetMinimizeToTray returns the minimize to tray setting
func (a *App) GetMinimizeToTray() bool {
	return a.ShouldMinimizeToTray()
}

// SetMinimizeToTray updates the minimize to tray setting
func (a *App) SetMinimizeToTray(enable bool) error {
	settings, err := a.GetAppSettings()
	if err != nil {
		return err
	}

	settings.MinimizeToTray = enable
	return a.SaveAppSettings(settings)
}

// IsStartedMinimized checks if app was started with --minimized flag
func IsStartedMinimized() bool {
	for _, arg := range os.Args[1:] {
		if strings.TrimPrefix(arg, "-") == "-minimized" || arg == "--minimized" {
			return true
		}
	}
	return false
}
