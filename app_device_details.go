package main

import (
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

// GetCameraSnapshot returns the current snapshot URL for a camera
func (a *App) GetCameraSnapshot(deviceID int64) (string, error) {
	if a.db == nil {
		return "", nil
	}

	cameraRepo := database.NewCameraRepository(a.db.DB())
	cam, err := cameraRepo.GetByDeviceID(deviceID)
	if err != nil {
		return "", err
	}

	if cam != nil && cam.SnapshotURL != "" {
		return cam.SnapshotURL, nil
	}

	return "", nil
}

// GetCameraStreamURL returns the RTSP stream URL for a camera
func (a *App) GetCameraStreamURL(deviceID int64) (string, error) {
	if a.db == nil {
		return "", nil
	}

	cameraRepo := database.NewCameraRepository(a.db.DB())
	cam, err := cameraRepo.GetByDeviceID(deviceID)
	if err != nil {
		return "", err
	}

	if cam != nil && cam.RTSPURL != "" {
		return cam.RTSPURL, nil
	}

	return "", nil
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
	if a.db == nil {
		return []SwitchWithPorts{}, nil
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	switchRepo := database.NewSwitchRepository(a.db.DB())

	// Get all switch devices
	devices, err := deviceRepo.GetByType(models.DeviceTypeSwitch)
	if err != nil {
		return []SwitchWithPorts{}, err
	}

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
