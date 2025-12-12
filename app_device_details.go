package main

import (
	"fmt"
	"log"

	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/models"
)

// GetDeviceMonitoringStats returns monitoring statistics for a specific device
func (a *App) GetDeviceMonitoringStats(deviceID int64) (*models.DeviceStats, error) {
	if a.db == nil {
		return nil, nil
	}

	repo := database.NewStatusHistoryRepository(a.db.DB())
	return repo.GetStats(deviceID)
}

// GetDeviceLatencyHistory returns latency data points for graphing
func (a *App) GetDeviceLatencyHistory(deviceID int64, hours int) ([]models.LatencyPoint, error) {
	if a.db == nil {
		return nil, nil
	}

	repo := database.NewStatusHistoryRepository(a.db.DB())
	return repo.GetLatencyPoints(deviceID, hours)
}

// GetDeviceUptimeHistory returns uptime data grouped by period
func (a *App) GetDeviceUptimeHistory(deviceID int64, period string, count int) ([]models.UptimePoint, error) {
	if a.db == nil {
		return nil, nil
	}

	repo := database.NewStatusHistoryRepository(a.db.DB())
	return repo.GetUptimeByPeriod(deviceID, period, count)
}

// GetDeviceStatusHistory returns raw status history
func (a *App) GetDeviceStatusHistory(deviceID int64, limit int) ([]models.StatusHistory, error) {
	if a.db == nil {
		return nil, nil
	}

	repo := database.NewStatusHistoryRepository(a.db.DB())
	return repo.GetHistory(deviceID, limit)
}

// GetDeviceStatusChanges returns status change events
func (a *App) GetDeviceStatusChanges(deviceID int64, limit int) ([]models.StatusHistory, error) {
	if a.db == nil {
		return nil, nil
	}

	repo := database.NewStatusHistoryRepository(a.db.DB())
	return repo.GetRecentChanges(deviceID, limit)
}

// GetSwitchPorts returns port information for a switch
func (a *App) GetSwitchPorts(deviceID int64) ([]models.SwitchPort, error) {
	if a.db == nil {
		return nil, nil
	}

	repo := database.NewDeviceRepository(a.db.DB())
	return repo.GetSwitchPorts(deviceID)
}

// UpdateSwitchPort updates a switch port
func (a *App) UpdateSwitchPort(port models.SwitchPort) error {
	if a.db == nil {
		return nil
	}

	repo := database.NewDeviceRepository(a.db.DB())
	return repo.UpdateSwitchPort(&port)
}

// LinkCameraToPort links a camera device to a switch port
func (a *App) LinkCameraToPort(portID int64, cameraID *int64) error {
	if a.db == nil {
		return nil
	}

	repo := database.NewDeviceRepository(a.db.DB())
	return repo.LinkCameraToPort(portID, cameraID)
}

// GetCameraSnapshot returns the snapshot URL for a camera, auto-discovering via ONVIF if needed
func (a *App) GetCameraSnapshot(deviceID int64) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Get device
	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil || device == nil {
		return "", fmt.Errorf("device not found")
	}

	// Get camera config
	cameraRepo := database.NewCameraRepository(a.db.DB())
	cam, err := cameraRepo.GetByDeviceID(deviceID)
	if err != nil || cam == nil {
		return "", fmt.Errorf("camera configuration not found")
	}

	// If no snapshot URL, try to get via ONVIF
	if cam.SnapshotURL == "" {
		err := a.RefreshCameraStreams(deviceID)
		if err != nil {
			return "", fmt.Errorf("no snapshot URL configured and ONVIF refresh failed: %w", err)
		}
		// Reload camera config
		cam, _ = cameraRepo.GetByDeviceID(deviceID)
	}

	if cam.SnapshotURL == "" {
		return "", fmt.Errorf("no snapshot URL available")
	}

	// Build full URL if needed
	snapshotURL := cam.SnapshotURL
	if !hasScheme(snapshotURL) {
		snapshotURL = fmt.Sprintf("http://%s%s", device.IPAddress, ensureLeadingSlash(cam.SnapshotURL))
	}

	return snapshotURL, nil
}

// GetCameraStreamURL returns the RTSP stream URL for a camera, auto-discovering via ONVIF if needed
func (a *App) GetCameraStreamURL(deviceID int64) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Get camera config
	cameraRepo := database.NewCameraRepository(a.db.DB())
	cam, err := cameraRepo.GetByDeviceID(deviceID)
	if err != nil || cam == nil {
		return "", fmt.Errorf("camera configuration not found")
	}

	// If no RTSP URL, try to get via ONVIF
	if cam.RTSPURL == "" {
		err := a.RefreshCameraStreams(deviceID)
		if err != nil {
			return "", fmt.Errorf("no RTSP URL configured and ONVIF refresh failed: %w", err)
		}
		// Reload camera config
		cam, _ = cameraRepo.GetByDeviceID(deviceID)
	}

	if cam.RTSPURL == "" {
		return "", fmt.Errorf("no RTSP URL available")
	}

	return cam.RTSPURL, nil
}

// SwitchWithPorts combines switch device info with its ports
type SwitchWithPorts struct {
	DeviceID     int64               `json:"device_id"`
	DeviceName   string              `json:"device_name"`
	IPAddress    string              `json:"ip_address"`
	PortCount    int                 `json:"port_count"`
	SFPPortCount int                 `json:"sfp_port_count"`
	Ports        []models.SwitchPort `json:"ports"`
}

// GetCameraPort returns the switch port ID that a camera is linked to
func (a *App) GetCameraPort(cameraDeviceID int64) (*int64, error) {
	if a.db == nil {
		return nil, nil
	}

	var portID int64
	err := a.db.DB().QueryRow(`
		SELECT id FROM switch_ports WHERE linked_camera_id = ?
	`, cameraDeviceID).Scan(&portID)

	if err != nil {
		return nil, nil // Not found or error
	}

	return &portID, nil
}

// GetSwitchesWithPorts returns all switches with their ports for camera linking
func (a *App) GetSwitchesWithPorts() ([]SwitchWithPorts, error) {
	log.Printf("GetSwitchesWithPorts: called")
	if a.db == nil {
		log.Printf("GetSwitchesWithPorts: db is nil!")
		return []SwitchWithPorts{}, nil
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	switchRepo := database.NewSwitchRepository(a.db.DB())

	// Get all switch devices
	devices, err := deviceRepo.GetByType(models.DeviceTypeSwitch)
	if err != nil {
		log.Printf("GetSwitchesWithPorts: error getting devices: %v", err)
		return []SwitchWithPorts{}, err
	}
	log.Printf("GetSwitchesWithPorts: found %d switch devices", len(devices))

	result := make([]SwitchWithPorts, 0)
	for _, d := range devices {
		sw, err := switchRepo.GetByDeviceID(d.ID)
		if err != nil {
			continue
		}
		if sw == nil {
			continue
		}

		ports, err := switchRepo.GetPorts(d.ID)
		if err != nil {
			ports = []models.SwitchPort{} // Initialize empty array on error
		}
		if ports == nil {
			ports = []models.SwitchPort{} // Ensure not nil
		}

		log.Printf("GetSwitchesWithPorts: switch %s (ID:%d) has %d ports", d.Name, d.ID, len(ports))

		result = append(result, SwitchWithPorts{
			DeviceID:     d.ID,
			DeviceName:   d.Name,
			IPAddress:    d.IPAddress,
			PortCount:    sw.PortCount,
			SFPPortCount: sw.SFPPortCount,
			Ports:        ports,
		})
	}

	log.Printf("GetSwitchesWithPorts: returning %d switches", len(result))
	return result, nil
}

// UplinkConnection represents an uplink connection between devices
type UplinkConnection struct {
	FromDeviceID int64  `json:"from_device_id"`
	ToDeviceID   int64  `json:"to_device_id"`
	PortID       int64  `json:"port_id"`
	DeviceType   string `json:"device_type"` // "switch" or "server"
}

// GetAllUplinkConnections returns all uplink connections (switches and servers to their parent switches)
func (a *App) GetAllUplinkConnections() ([]UplinkConnection, error) {
	if a.db == nil {
		return []UplinkConnection{}, nil
	}

	connections := make([]UplinkConnection, 0)

	// Get switch uplinks
	switchRows, err := a.db.DB().Query(`
		SELECT s.device_id, s.uplink_switch_id, s.uplink_port_id
		FROM switches s
		WHERE s.uplink_switch_id IS NOT NULL AND s.uplink_port_id IS NOT NULL
	`)
	if err == nil {
		defer switchRows.Close()
		for switchRows.Next() {
			var deviceID, uplinkSwitchID, uplinkPortID int64
			if err := switchRows.Scan(&deviceID, &uplinkSwitchID, &uplinkPortID); err == nil {
				connections = append(connections, UplinkConnection{
					FromDeviceID: uplinkSwitchID, // Parent switch
					ToDeviceID:   deviceID,       // Child switch
					PortID:       uplinkPortID,
					DeviceType:   "switch",
				})
			}
		}
	}

	// Get server uplinks
	serverRows, err := a.db.DB().Query(`
		SELECT s.device_id, s.uplink_switch_id, s.uplink_port_id
		FROM servers s
		WHERE s.uplink_switch_id IS NOT NULL AND s.uplink_port_id IS NOT NULL
	`)
	if err == nil {
		defer serverRows.Close()
		for serverRows.Next() {
			var deviceID, uplinkSwitchID, uplinkPortID int64
			if err := serverRows.Scan(&deviceID, &uplinkSwitchID, &uplinkPortID); err == nil {
				connections = append(connections, UplinkConnection{
					FromDeviceID: uplinkSwitchID, // Parent switch
					ToDeviceID:   deviceID,       // Server
					PortID:       uplinkPortID,
					DeviceType:   "server",
				})
			}
		}
	}

	return connections, nil
}
