//go:build windows

package autostart

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	registryKey = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName     = "NetVisionMonitor"
)

// IsEnabled checks if autostart is enabled
func IsEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.QUERY_VALUE)
	if err != nil {
		return false, nil // Key doesn't exist or can't be opened
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appName)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// Enable adds the application to Windows startup
func Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks and get absolute path
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Add quotes around path in case it contains spaces, and add --minimized flag
	value := fmt.Sprintf(`"%s" --minimized`, exePath)
	if err := key.SetStringValue(appName, value); err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	return nil
}

// Disable removes the application from Windows startup
func Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
	if err != nil {
		return nil // Key doesn't exist, nothing to disable
	}
	defer key.Close()

	if err := key.DeleteValue(appName); err != nil {
		if err == registry.ErrNotExist {
			return nil // Value doesn't exist, already disabled
		}
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	return nil
}

// Toggle toggles the autostart state
func Toggle() (bool, error) {
	enabled, err := IsEnabled()
	if err != nil {
		return false, err
	}

	if enabled {
		if err := Disable(); err != nil {
			return true, err
		}
		return false, nil
	}

	if err := Enable(); err != nil {
		return false, err
	}
	return true, nil
}
