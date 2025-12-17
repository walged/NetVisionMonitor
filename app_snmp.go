package main

import (
	"fmt"
	"log"
	"time"

	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/models"
	"netvisionmonitor/internal/snmp"
)

// createSNMPClient creates SNMP client based on switch settings (for read operations)
func createSNMPClient(ipAddress string, sw *models.Switch) *snmp.TFortisClient {
	version := sw.SNMPVersion
	if version == "" {
		version = "v2c"
	}

	return snmp.NewTFortisClientAuto(
		ipAddress,
		version,
		sw.SNMPCommunity,
		sw.SNMPv3User,
		sw.SNMPv3Security,
		sw.SNMPv3AuthProto,
		sw.SNMPv3AuthPass,
		sw.SNMPv3PrivProto,
		sw.SNMPv3PrivPass,
	)
}

// createSNMPWriteClient creates SNMP client for write operations (uses write community)
func createSNMPWriteClient(ipAddress string, sw *models.Switch) *snmp.TFortisClient {
	version := sw.SNMPVersion
	if version == "" {
		version = "v2c"
	}

	// Use write community if available, otherwise fall back to read community
	community := sw.SNMPWriteCommunity
	if community == "" {
		community = sw.SNMPCommunity
	}

	return snmp.NewTFortisClientAuto(
		ipAddress,
		version,
		community,
		sw.SNMPv3User,
		sw.SNMPv3Security,
		sw.SNMPv3AuthProto,
		sw.SNMPv3AuthPass,
		sw.SNMPv3PrivProto,
		sw.SNMPv3PrivPass,
	)
}

// GetSwitchSNMPData retrieves all SNMP data for a switch
func (a *App) GetSwitchSNMPData(deviceID int64) (*models.SwitchSNMPData, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	result := &models.SwitchSNMPData{
		DeviceID: deviceID,
	}

	// Get device and switch info
	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		result.Error = fmt.Sprintf("Device not found: %v", err)
		return result, nil
	}

	if device.Type != models.DeviceTypeSwitch {
		result.Error = "Device is not a switch"
		return result, nil
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		result.Error = "Switch configuration not found"
		return result, nil
	}

	// Check SNMP configuration
	log.Printf("DEBUG: Switch SNMP config - Version: '%s', User: '%s', Security: '%s', Community: '%s'",
		sw.SNMPVersion, sw.SNMPv3User, sw.SNMPv3Security, sw.SNMPCommunity)
	if sw.SNMPVersion == "v3" {
		if sw.SNMPv3User == "" {
			result.Error = "SNMPv3 user not configured"
			return result, nil
		}
	} else {
		if sw.SNMPCommunity == "" {
			result.Error = "SNMP community not configured"
			return result, nil
		}
	}

	// Create TFortis SNMP client
	client := createSNMPClient(device.IPAddress, sw)

	// Test connection first
	err = client.TestConnection()
	if err != nil {
		result.Error = fmt.Sprintf("SNMP connection failed: %v", err)
		return result, nil
	}

	// Get system info
	sysInfo, err := client.GetSystemInfo()
	if err == nil {
		result.SystemInfo = &models.SNMPSystemInfo{
			FirmwareVersion: sysInfo.FirmwareVersion,
		}
		if sysInfo.UPS != nil {
			result.SystemInfo.UPS = &models.SNMPUPSInfo{
				Present: sysInfo.UPS.Present,
				Status:  sysInfo.UPS.Status,
				Charge:  sysInfo.UPS.Charge,
			}
		}
	}

	// Get port info
	ports, err := client.GetAllPortsInfo(sw.PortCount)
	if err == nil {
		result.Ports = make([]models.SNMPPortInfo, len(ports))
		for i, p := range ports {
			result.Ports[i] = models.SNMPPortInfo{
				PortNumber:  p.PortNumber,
				Status:      p.Status,
				Speed:       p.Speed,
				SpeedStr:    p.SpeedStr,
				RxBytes:     p.RxBytes,
				TxBytes:     p.TxBytes,
				Description: p.Description,
			}
		}
	}

	// Get PoE info
	poeInfos, err := client.GetAllPoEInfo(sw.PortCount)
	if err == nil {
		result.PoE = make([]models.SNMPPoEInfo, len(poeInfos))
		for i, p := range poeInfos {
			result.PoE[i] = models.SNMPPoEInfo{
				PortNumber: p.PortNumber,
				Enabled:    p.Enabled,
				Active:     p.Active,
				Status:     p.Status,
				PowerMW:    p.PowerMW,
				PowerW:     p.PowerW,
			}
		}
	}

	return result, nil
}

// GetSwitchPortSNMP retrieves SNMP data for a specific port
func (a *App) GetSwitchPortSNMP(deviceID int64, portNumber int) (*models.SNMPPortInfo, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		return nil, fmt.Errorf("switch configuration not found")
	}

	client := createSNMPClient(device.IPAddress, sw)
	portInfo, err := client.GetPortInfo(portNumber)
	if err != nil {
		return nil, err
	}

	return &models.SNMPPortInfo{
		PortNumber:  portInfo.PortNumber,
		Status:      portInfo.Status,
		Speed:       portInfo.Speed,
		SpeedStr:    portInfo.SpeedStr,
		RxBytes:     portInfo.RxBytes,
		TxBytes:     portInfo.TxBytes,
		Description: portInfo.Description,
	}, nil
}

// GetSwitchPoESNMP retrieves PoE status for all ports
func (a *App) GetSwitchPoESNMP(deviceID int64) ([]models.SNMPPoEInfo, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		return nil, fmt.Errorf("switch configuration not found")
	}

	client := createSNMPClient(device.IPAddress, sw)
	poeInfos, err := client.GetAllPoEInfo(sw.PortCount)
	if err != nil {
		return nil, err
	}

	result := make([]models.SNMPPoEInfo, len(poeInfos))
	for i, p := range poeInfos {
		result[i] = models.SNMPPoEInfo{
			PortNumber: p.PortNumber,
			Enabled:    p.Enabled,
			Active:     p.Active,
			Status:     p.Status,
			PowerMW:    p.PowerMW,
			PowerW:     p.PowerW,
		}
	}

	return result, nil
}

// SetPoEEnabled enables or disables PoE on a port
func (a *App) SetPoEEnabled(deviceID int64, portNumber int, enabled bool) error {
	log.Printf("SetPoEEnabled called: device=%d, port=%d, enabled=%v", deviceID, portNumber, enabled)

	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		return fmt.Errorf("switch configuration not found")
	}

	log.Printf("Creating SNMP write client for %s (version: %s)", device.IPAddress, sw.SNMPVersion)

	client := createSNMPWriteClient(device.IPAddress, sw)
	err = client.SetPoEEnabled(portNumber, enabled)
	if err != nil {
		log.Printf("SetPoEEnabled failed: %v", err)
		return fmt.Errorf("failed to set PoE state: %w", err)
	}

	action := "disabled"
	if enabled {
		action = "enabled"
	}
	log.Printf("PoE %s on device %d port %d - SUCCESS", action, deviceID, portNumber)

	return nil
}

// RestartPoEPort restarts PoE on a port (turns off, waits, turns on)
func (a *App) RestartPoEPort(deviceID int64, portNumber int) error {
	log.Printf("RestartPoEPort called: device=%d, port=%d", deviceID, portNumber)

	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		return fmt.Errorf("switch configuration not found")
	}

	if sw.SNMPVersion == "v3" {
		log.Printf("RestartPoEPort: IP=%s, Version=%s, User=%s, Security=%s, AuthProto=%s, PrivProto=%s",
			device.IPAddress, sw.SNMPVersion, sw.SNMPv3User, sw.SNMPv3Security, sw.SNMPv3AuthProto, sw.SNMPv3PrivProto)
	} else {
		log.Printf("RestartPoEPort: IP=%s, Version=%s, Community=%s, WriteCommunity=%s",
			device.IPAddress, sw.SNMPVersion, sw.SNMPCommunity, sw.SNMPWriteCommunity)
	}

	client := createSNMPWriteClient(device.IPAddress, sw)

	// Disable PoE
	err = client.SetPoEEnabled(portNumber, false)
	if err != nil {
		return fmt.Errorf("failed to disable PoE: %w", err)
	}

	log.Printf("PoE disabled on device %d port %d, waiting 3 seconds...", deviceID, portNumber)

	// Wait 3 seconds
	time.Sleep(3 * time.Second)

	// Enable PoE
	err = client.SetPoEEnabled(portNumber, true)
	if err != nil {
		return fmt.Errorf("failed to enable PoE: %w", err)
	}

	log.Printf("PoE enabled on device %d port %d", deviceID, portNumber)

	return nil
}

// SetPortEnabled enables or disables a port (via ifAdminStatus)
func (a *App) SetPortEnabled(deviceID int64, portNumber int, enabled bool) error {
	log.Printf("SetPortEnabled called: device=%d, port=%d, enabled=%v", deviceID, portNumber, enabled)

	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		return fmt.Errorf("switch configuration not found")
	}

	log.Printf("Creating SNMP write client for %s (version: %s)", device.IPAddress, sw.SNMPVersion)

	client := createSNMPWriteClient(device.IPAddress, sw)
	err = client.SetPortEnabled(portNumber, enabled)
	if err != nil {
		log.Printf("SetPortEnabled failed: %v", err)
		return fmt.Errorf("failed to set port state: %w", err)
	}

	action := "disabled"
	if enabled {
		action = "enabled"
	}
	log.Printf("Port %s on device %d port %d - SUCCESS", action, deviceID, portNumber)

	return nil
}

// RestartPort restarts a port (turns off, waits, turns on)
func (a *App) RestartPort(deviceID int64, portNumber int) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		return fmt.Errorf("switch configuration not found")
	}

	client := createSNMPWriteClient(device.IPAddress, sw)

	// Disable port
	err = client.SetPortEnabled(portNumber, false)
	if err != nil {
		return fmt.Errorf("failed to disable port: %w", err)
	}

	log.Printf("Port disabled on device %d port %d, waiting 3 seconds...", deviceID, portNumber)

	// Wait 3 seconds
	time.Sleep(3 * time.Second)

	// Enable port
	err = client.SetPortEnabled(portNumber, true)
	if err != nil {
		return fmt.Errorf("failed to enable port: %w", err)
	}

	log.Printf("Port enabled on device %d port %d", deviceID, portNumber)

	return nil
}

// TestSNMPConnection tests SNMP connection to a switch
func (a *App) TestSNMPConnection(deviceID int64) (bool, error) {
	if a.db == nil {
		return false, fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		return false, fmt.Errorf("device not found: %w", err)
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		return false, fmt.Errorf("switch configuration not found")
	}

	// Check SNMP configuration
	if sw.SNMPVersion == "v3" {
		if sw.SNMPv3User == "" {
			return false, fmt.Errorf("SNMPv3 user not configured")
		}
	} else {
		if sw.SNMPCommunity == "" {
			return false, fmt.Errorf("SNMP community not configured")
		}
	}

	client := createSNMPClient(device.IPAddress, sw)
	err = client.TestConnection()
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetSwitchAutoRestartSettings retrieves AutoRestart settings for all ports
func (a *App) GetSwitchAutoRestartSettings(deviceID int64) ([]models.SNMPAutoRestartInfo, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		return nil, fmt.Errorf("switch configuration not found")
	}

	client := createSNMPClient(device.IPAddress, sw)

	result := make([]models.SNMPAutoRestartInfo, 0, sw.PortCount)
	for i := 1; i <= sw.PortCount; i++ {
		info, err := client.GetAutoRestartInfo(i)
		if err != nil {
			continue
		}
		result = append(result, models.SNMPAutoRestartInfo{
			PortNumber: info.PortNumber,
			Mode:       info.Mode,
			ModeStr:    info.ModeStr,
			PingIP:     info.PingIP,
			LinkSpeed:  info.LinkSpeed,
			Status:     info.Status,
		})
	}

	return result, nil
}

// SetAutoRestartMode sets AutoRestart mode for a port
func (a *App) SetAutoRestartMode(deviceID int64, portNumber int, mode int) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	switchRepo := database.NewSwitchRepository(a.db.DB())
	sw, err := switchRepo.GetByDeviceID(deviceID)
	if err != nil || sw == nil {
		return fmt.Errorf("switch configuration not found")
	}

	client := createSNMPWriteClient(device.IPAddress, sw)
	err = client.SetAutoRestartMode(portNumber, mode)
	if err != nil {
		return fmt.Errorf("failed to set AutoRestart mode: %w", err)
	}

	log.Printf("AutoRestart mode set to %d on device %d port %d", mode, deviceID, portNumber)
	return nil
}
