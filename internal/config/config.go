package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	DataDir    string `json:"data_dir"`
	LogDir     string `json:"log_dir"`
	CacheDir   string `json:"cache_dir"`
	SchemaDir  string `json:"schema_dir"`
	ExportDir  string `json:"export_dir"`
	FFmpegPath string `json:"ffmpeg_path"`
	IsPortable bool   `json:"is_portable"`
}

var (
	config     *Config
	configOnce sync.Once
)

// GetExecutableDir returns the directory containing the executable
func GetExecutableDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

// Initialize sets up the configuration
func Initialize() (*Config, error) {
	var initErr error

	configOnce.Do(func() {
		exeDir, err := GetExecutableDir()
		if err != nil {
			initErr = err
			return
		}

		// Check if running in portable mode
		// Portable mode: data directory is next to executable
		// Installed mode: data directory is in ProgramData
		portableDataDir := filepath.Join(exeDir, "data")
		isPortable := true

		// Check if portable data dir exists or if we can create it
		if _, err := os.Stat(portableDataDir); os.IsNotExist(err) {
			// Check if we're in a Program Files directory
			if isInProgramFiles(exeDir) {
				isPortable = false
			}
		}

		var dataDir, logDir string
		if isPortable {
			dataDir = portableDataDir
			logDir = filepath.Join(exeDir, "logs")
		} else {
			programData := os.Getenv("ProgramData")
			if programData == "" {
				programData = "C:\\ProgramData"
			}
			dataDir = filepath.Join(programData, "NetVisionMonitor", "data")
			logDir = filepath.Join(programData, "NetVisionMonitor", "logs")
		}

		config = &Config{
			DataDir:    dataDir,
			LogDir:     logDir,
			CacheDir:   filepath.Join(dataDir, "cache"),
			SchemaDir:  filepath.Join(exeDir, "schemas"),
			ExportDir:  filepath.Join(exeDir, "exports"),
			FFmpegPath: filepath.Join(exeDir, "ffmpeg", "ffmpeg.exe"),
			IsPortable: isPortable,
		}

		// Create directories
		dirs := []string{
			config.DataDir,
			config.LogDir,
			config.CacheDir,
			config.SchemaDir,
			config.ExportDir,
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				initErr = err
				return
			}
		}
	})

	return config, initErr
}

// Get returns the current configuration
func Get() *Config {
	return config
}

// isInProgramFiles checks if path is under Program Files
func isInProgramFiles(path string) bool {
	programFiles := os.Getenv("ProgramFiles")
	programFilesX86 := os.Getenv("ProgramFiles(x86)")

	if programFiles != "" && len(path) >= len(programFiles) {
		if path[:len(programFiles)] == programFiles {
			return true
		}
	}
	if programFilesX86 != "" && len(path) >= len(programFilesX86) {
		if path[:len(programFilesX86)] == programFilesX86 {
			return true
		}
	}
	return false
}

// SaveJSON saves the config to a JSON file
func (c *Config) SaveJSON() error {
	path := filepath.Join(c.DataDir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
