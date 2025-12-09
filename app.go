package main

import (
	"context"
	"fmt"
	"path/filepath"

	"netvisionmonitor/internal/config"
	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/encryption"
	"netvisionmonitor/internal/logger"
	"netvisionmonitor/internal/monitoring"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	db      *database.Database
	cfg     *config.Config
	monitor *monitoring.Monitor
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize configuration
	cfg, err := config.Initialize()
	if err != nil {
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Ошибка инициализации",
			Message: fmt.Sprintf("Не удалось инициализировать конфигурацию: %v", err),
		})
		return
	}
	a.cfg = cfg

	// Initialize logger
	logDir := filepath.Join(cfg.DataDir, "logs")
	if err := logger.Init(logDir, logger.LevelInfo); err != nil {
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.WarningDialog,
			Title:   "Предупреждение",
			Message: fmt.Sprintf("Не удалось инициализировать логгер: %v", err),
		})
		// Continue without file logging
	}

	// Clean old logs (keep 30 days)
	logger.CleanOldLogs(logDir, 30)

	logger.Info("Starting NetVisionMonitor v1.1.0")
	logger.Info("Data directory: %s", cfg.DataDir)
	logger.Info("Portable mode: %v", cfg.IsPortable)

	// Initialize encryption
	if err := encryption.Initialize(cfg.DataDir); err != nil {
		logger.Error("Failed to initialize encryption: %v", err)
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Ошибка инициализации",
			Message: fmt.Sprintf("Не удалось инициализировать шифрование: %v", err),
		})
		return
	}
	logger.Info("Encryption initialized")

	// Initialize database
	db, err := database.Initialize(cfg.DataDir)
	if err != nil {
		logger.Error("Failed to initialize database: %v", err)
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Ошибка инициализации",
			Message: fmt.Sprintf("Не удалось инициализировать базу данных: %v", err),
		})
		return
	}
	a.db = db
	logger.Info("Database initialized")

	// Fix existing port types based on sfp_port_count
	if err := db.FixExistingPortTypes(); err != nil {
		logger.Warn("Failed to fix existing port types: %v", err)
	} else {
		logger.Info("Port types verified")
	}

	// Initialize monitoring
	a.initMonitoring()

	// Auto-start monitoring
	a.monitor.Start()
	logger.Info("Monitoring started")

	// Initialize system tray
	InitTray(a)

	logger.Info("Application started successfully")
}

// beforeClose is called when the user tries to close the window
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	// Check if we should minimize to tray instead of closing
	if a.ShouldMinimizeToTray() {
		a.MinimizeToTray()
		return true // Prevent closing, just hide
	}
	return false // Allow closing
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	logger.Info("Application shutting down...")

	// Stop system tray
	StopTray()

	// Stop monitoring
	if a.monitor != nil {
		a.monitor.Stop()
		logger.Info("Monitoring stopped")
	}

	if a.db != nil {
		a.db.Close()
		logger.Info("Database closed")
	}

	logger.Info("Application shutdown complete")

	// Close logger
	logger.Get().Close()
}

// GetAppInfo returns application information
func (a *App) GetAppInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":       "NetVisionMonitor",
		"version":    "1.1.0",
		"isPortable": a.cfg.IsPortable,
	}
}

// GetLogPath returns the path to the current log file
func (a *App) GetLogPath() string {
	return logger.Get().GetFilePath()
}

// OpenLogFolder opens the log folder in file explorer
func (a *App) OpenLogFolder() error {
	if a.cfg == nil {
		return fmt.Errorf("config not initialized")
	}
	logDir := filepath.Join(a.cfg.DataDir, "logs")
	runtime.BrowserOpenURL(a.ctx, "file:///"+filepath.ToSlash(logDir))
	return nil
}

