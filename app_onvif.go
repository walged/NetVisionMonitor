package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/onvif"
)

// ONVIFProfile contains ONVIF media profile info
type ONVIFProfile struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

// ONVIFDiscoveryResult contains the result of ONVIF camera discovery
type ONVIFDiscoveryResult struct {
	Success      bool           `json:"success"`
	Error        string         `json:"error,omitempty"`
	Manufacturer string         `json:"manufacturer,omitempty"`
	Model        string         `json:"model,omitempty"`
	Firmware     string         `json:"firmware,omitempty"`
	SerialNumber string         `json:"serial_number,omitempty"`
	StreamURI    string         `json:"stream_uri,omitempty"`
	SnapshotURI  string         `json:"snapshot_uri,omitempty"`
	Profiles     []ONVIFProfile `json:"profiles,omitempty"`
}

// DiscoverONVIFCamera discovers camera info via ONVIF protocol
func (a *App) DiscoverONVIFCamera(ipAddress string, port int, username, password string) *ONVIFDiscoveryResult {
	log.Printf("DiscoverONVIFCamera: ip=%s, port=%d, user=%s", ipAddress, port, username)

	if ipAddress == "" {
		return &ONVIFDiscoveryResult{
			Success: false,
			Error:   "IP address is required",
		}
	}

	if port <= 0 {
		port = 80
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := onvif.NewClient(ipAddress, port, username, password)
	info, err := client.DiscoverCamera(ctx)
	if err != nil {
		log.Printf("ONVIF discovery failed: %v", err)
		return &ONVIFDiscoveryResult{
			Success: false,
			Error:   fmt.Sprintf("ONVIF discovery failed: %v", err),
		}
	}

	result := &ONVIFDiscoveryResult{
		Success: true,
	}

	if info.DeviceInfo != nil {
		result.Manufacturer = info.DeviceInfo.Manufacturer
		result.Model = info.DeviceInfo.Model
		result.Firmware = info.DeviceInfo.FirmwareVersion
		result.SerialNumber = info.DeviceInfo.SerialNumber
	}

	result.StreamURI = info.StreamURI
	result.SnapshotURI = info.SnapshotURI

	for _, p := range info.Profiles {
		result.Profiles = append(result.Profiles, ONVIFProfile{
			Token: p.Token,
			Name:  p.Name,
		})
	}

	log.Printf("ONVIF discovery success: manufacturer=%s, model=%s, stream=%s, snapshot=%s",
		result.Manufacturer, result.Model, result.StreamURI, result.SnapshotURI)

	return result
}

// TestONVIFConnection tests ONVIF connectivity for a camera
func (a *App) TestONVIFConnection(ipAddress string, port int, username, password string) (bool, string) {
	if ipAddress == "" {
		return false, "IP address is required"
	}

	if port <= 0 {
		port = 80
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := onvif.NewClient(ipAddress, port, username, password)
	err := client.TestConnection(ctx)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}

	return true, "ONVIF connection successful"
}

// GetCameraONVIFInfo retrieves ONVIF info for an existing camera device
func (a *App) GetCameraONVIFInfo(deviceID int64) *ONVIFDiscoveryResult {
	if a.db == nil {
		return &ONVIFDiscoveryResult{
			Success: false,
			Error:   "Database not initialized",
		}
	}

	// Get device
	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil || device == nil {
		return &ONVIFDiscoveryResult{
			Success: false,
			Error:   "Device not found",
		}
	}

	// Get camera config
	cameraRepo := database.NewCameraRepository(a.db.DB())
	cam, err := cameraRepo.GetByDeviceID(deviceID)
	if err != nil || cam == nil {
		return &ONVIFDiscoveryResult{
			Success: false,
			Error:   "Camera configuration not found",
		}
	}

	// Get credentials if available
	var username, password string
	if device.CredentialID != nil {
		credRepo := database.NewCredentialRepository(a.db.DB())
		cred, err := credRepo.GetByID(*device.CredentialID)
		if err == nil && cred != nil {
			username = cred.Username
			password = cred.Password
		}
	}

	return a.DiscoverONVIFCamera(device.IPAddress, cam.ONVIFPort, username, password)
}

// RefreshCameraStreams updates camera RTSP and snapshot URLs from ONVIF
func (a *App) RefreshCameraStreams(deviceID int64) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Get device
	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil || device == nil {
		return fmt.Errorf("device not found")
	}

	// Get camera config
	cameraRepo := database.NewCameraRepository(a.db.DB())
	cam, err := cameraRepo.GetByDeviceID(deviceID)
	if err != nil || cam == nil {
		return fmt.Errorf("camera configuration not found")
	}

	// Get credentials if available
	var username, password string
	if device.CredentialID != nil {
		credRepo := database.NewCredentialRepository(a.db.DB())
		cred, err := credRepo.GetByID(*device.CredentialID)
		if err == nil && cred != nil {
			username = cred.Username
			password = cred.Password
		}
	}

	// Discover via ONVIF
	result := a.DiscoverONVIFCamera(device.IPAddress, cam.ONVIFPort, username, password)
	if !result.Success {
		return fmt.Errorf("ONVIF discovery failed: %s", result.Error)
	}

	// Update camera with discovered URLs
	if result.StreamURI != "" {
		cam.RTSPURL = result.StreamURI
	}
	if result.SnapshotURI != "" {
		cam.SnapshotURL = result.SnapshotURI
	}

	// Update in database
	if err := cameraRepo.Update(cam); err != nil {
		return fmt.Errorf("failed to update camera: %w", err)
	}

	// Update device manufacturer/model if discovered
	if result.Manufacturer != "" || result.Model != "" {
		if result.Manufacturer != "" {
			device.Manufacturer = result.Manufacturer
		}
		if result.Model != "" {
			device.Model = result.Model
		}
		if err := deviceRepo.Update(device); err != nil {
			log.Printf("Warning: failed to update device info: %v", err)
		}
	}

	log.Printf("Camera %d streams refreshed: rtsp=%s, snapshot=%s", deviceID, cam.RTSPURL, cam.SnapshotURL)
	return nil
}

// FetchCameraSnapshotBase64 fetches camera snapshot and returns as base64 data URI
func (a *App) FetchCameraSnapshotBase64(deviceID int64) (string, error) {
	log.Printf("FetchCameraSnapshotBase64: deviceID=%d", deviceID)

	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Get device
	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		log.Printf("FetchCameraSnapshotBase64: error getting device: %v", err)
		return "", fmt.Errorf("device not found: %w", err)
	}
	if device == nil {
		return "", fmt.Errorf("device not found")
	}
	log.Printf("FetchCameraSnapshotBase64: device=%s, ip=%s", device.Name, device.IPAddress)

	// Get camera config
	cameraRepo := database.NewCameraRepository(a.db.DB())
	cam, err := cameraRepo.GetByDeviceID(deviceID)
	if err != nil {
		log.Printf("FetchCameraSnapshotBase64: error getting camera: %v", err)
		return "", fmt.Errorf("camera configuration not found: %w", err)
	}
	if cam == nil {
		return "", fmt.Errorf("camera configuration not found")
	}
	log.Printf("FetchCameraSnapshotBase64: onvif_port=%d, snapshot_url=%s, rtsp_url=%s", cam.ONVIFPort, cam.SnapshotURL, cam.RTSPURL)

	// If no snapshot URL, try to get via ONVIF
	if cam.SnapshotURL == "" {
		log.Printf("No snapshot URL, trying ONVIF discovery for device %d", deviceID)
		err := a.RefreshCameraStreams(deviceID)
		if err != nil {
			log.Printf("FetchCameraSnapshotBase64: ONVIF refresh failed: %v", err)
			return "", fmt.Errorf("Нет snapshot URL. ONVIF ошибка: %v", err)
		}
		// Reload camera config
		cam, _ = cameraRepo.GetByDeviceID(deviceID)
		if cam.SnapshotURL == "" {
			return "", fmt.Errorf("ONVIF не вернул snapshot URL. Проверьте ONVIF порт и учётные данные")
		}
	}

	// Build full URL
	snapshotURL := cam.SnapshotURL
	if !hasScheme(snapshotURL) {
		snapshotURL = fmt.Sprintf("http://%s%s", device.IPAddress, ensureLeadingSlash(cam.SnapshotURL))
	}

	log.Printf("Fetching snapshot from: %s", snapshotURL)

	// Get credentials if available
	var username, password string
	if device.CredentialID != nil {
		credRepo := database.NewCredentialRepository(a.db.DB())
		cred, err := credRepo.GetByID(*device.CredentialID)
		if err == nil && cred != nil {
			username = cred.Username
			password = cred.Password
		}
	}

	// Create HTTP client with SSL skip for self-signed certificates (common in IP cameras)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	req, err := http.NewRequest("GET", snapshotURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if credentials available
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return "", fmt.Errorf("требуется авторизация (401). Проверьте учётные данные камеры")
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ошибка камеры: статус %d", resp.StatusCode)
	}

	// Read image data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read snapshot data: %w", err)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("empty snapshot response")
	}

	// Determine content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	// Return as data URI
	b64 := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, b64)

	log.Printf("Snapshot fetched successfully: %d bytes, type: %s", len(data), contentType)
	return dataURI, nil
}

// FetchSnapshotFromURL fetches snapshot from a specific URL with optional auth
func (a *App) FetchSnapshotFromURL(snapshotURL, username, password string) (string, error) {
	if snapshotURL == "" {
		return "", fmt.Errorf("snapshot URL is required")
	}

	log.Printf("Fetching snapshot from URL: %s", snapshotURL)

	// Create HTTP client with SSL skip for self-signed certificates
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	req, err := http.NewRequest("GET", snapshotURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if credentials available
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return "", fmt.Errorf("authentication required (401)")
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// Read image data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("empty response")
	}

	// Determine content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	// Return as data URI
	b64 := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, b64)

	log.Printf("Snapshot fetched: %d bytes, type: %s", len(data), contentType)
	return dataURI, nil
}

// Helper functions for URL processing
func hasScheme(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme != ""
}

func ensureLeadingSlash(path string) string {
	if len(path) > 0 && path[0] != '/' {
		return "/" + path
	}
	return path
}
