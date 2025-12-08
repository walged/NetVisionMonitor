package main

import (
	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/models"
	"netvisionmonitor/internal/monitoring"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// MonitoringStatus represents current monitoring state
type MonitoringStatus struct {
	Running  bool   `json:"running"`
	Interval int    `json:"interval"` // seconds
	Workers  int    `json:"workers"`
}

// initMonitoring initializes the monitoring system
func (a *App) initMonitoring() {
	cfg := monitoring.DefaultConfig()
	a.monitor = monitoring.NewMonitor(a.db, cfg)

	// Set up event handlers
	a.monitor.SetStatusChangeHandler(a.onDeviceStatusChange)
	a.monitor.SetEventHandler(a.onMonitoringEvent)
}

// onDeviceStatusChange handles device status changes
func (a *App) onDeviceStatusChange(deviceID int64, oldStatus, newStatus string) {
	// Emit event to frontend
	runtime.EventsEmit(a.ctx, "device:status", map[string]interface{}{
		"device_id":  deviceID,
		"old_status": oldStatus,
		"new_status": newStatus,
	})
}

// onMonitoringEvent handles monitoring events
func (a *App) onMonitoringEvent(event *models.Event) {
	// Save event to database
	eventRepo := database.NewEventRepository(a.db.DB())
	eventRepo.Create(event)

	// Emit event to frontend
	runtime.EventsEmit(a.ctx, "event:new", map[string]interface{}{
		"id":         event.ID,
		"device_id":  event.DeviceID,
		"type":       event.Type,
		"level":      event.Level,
		"message":    event.Message,
		"created_at": event.CreatedAt,
	})
}

// StartMonitoring starts the monitoring cycle
func (a *App) StartMonitoring() error {
	if a.monitor == nil {
		a.initMonitoring()
	}
	a.monitor.Start()
	runtime.EventsEmit(a.ctx, "monitoring:started", nil)
	return nil
}

// StopMonitoring stops the monitoring cycle
func (a *App) StopMonitoring() error {
	if a.monitor != nil {
		a.monitor.Stop()
		runtime.EventsEmit(a.ctx, "monitoring:stopped", nil)
	}
	return nil
}

// GetMonitoringStatus returns current monitoring status
func (a *App) GetMonitoringStatus() MonitoringStatus {
	if a.monitor == nil {
		return MonitoringStatus{
			Running:  false,
			Interval: 30,
			Workers:  10,
		}
	}

	return MonitoringStatus{
		Running:  a.monitor.IsRunning(),
		Interval: 30, // TODO: get from config
		Workers:  10,
	}
}

// SetMonitoringInterval updates the monitoring interval
func (a *App) SetMonitoringInterval(seconds int) error {
	if a.monitor != nil {
		a.monitor.SetInterval(time.Duration(seconds) * time.Second)
	}
	return nil
}

// RunMonitoringOnce performs a single monitoring cycle
func (a *App) RunMonitoringOnce() error {
	if a.monitor == nil {
		a.initMonitoring()
	}
	a.monitor.RunOnce()
	return nil
}
