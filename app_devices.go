package main

import (
	"fmt"
	"log"

	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/models"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// DeviceInput is used for creating/updating devices from frontend
type DeviceInput struct {
	ID           int64  `json:"id,omitempty"`
	Name         string `json:"name"`
	IPAddress    string `json:"ip_address"`
	Type         string `json:"type"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	CredentialID *int64 `json:"credential_id,omitempty"`

	// Switch-specific
	SNMPCommunity string `json:"snmp_community,omitempty"`
	SNMPVersion   string `json:"snmp_version,omitempty"`
	PortCount     int    `json:"port_count,omitempty"`
	SFPPortCount  int    `json:"sfp_port_count,omitempty"`

	// SNMPv3-specific
	SNMPv3User      string `json:"snmpv3_user,omitempty"`
	SNMPv3Security  string `json:"snmpv3_security,omitempty"`
	SNMPv3AuthProto string `json:"snmpv3_auth_proto,omitempty"`
	SNMPv3AuthPass  string `json:"snmpv3_auth_pass,omitempty"`
	SNMPv3PrivProto string `json:"snmpv3_priv_proto,omitempty"`
	SNMPv3PrivPass  string `json:"snmpv3_priv_pass,omitempty"`

	// Camera-specific
	RTSPURL      string `json:"rtsp_url,omitempty"`
	ONVIFPort    int    `json:"onvif_port,omitempty"`
	SnapshotURL  string `json:"snapshot_url,omitempty"`
	StreamType   string `json:"stream_type,omitempty"`
	SwitchPortID *int64 `json:"switch_port_id,omitempty"` // Link camera to switch port

	// Server-specific
	TCPPorts string `json:"tcp_ports,omitempty"`
	UseSNMP  bool   `json:"use_snmp,omitempty"`

	// Uplink settings (for switches and servers)
	UplinkSwitchID *int64 `json:"uplink_switch_id,omitempty"` // Parent switch ID
	UplinkPortID   *int64 `json:"uplink_port_id,omitempty"`   // SFP port ID on parent switch
}

// GetDevices returns all devices
func (a *App) GetDevices() ([]models.Device, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	repo := database.NewDeviceRepository(a.db.DB())
	return repo.GetAll()
}

// DeviceFilterInput is the input for filtering devices
type DeviceFilterInput struct {
	Type      string `json:"type"`
	Status    string `json:"status"`
	Search    string `json:"search"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

// GetDevicesPaginated returns devices with pagination and filtering
func (a *App) GetDevicesPaginated(filter DeviceFilterInput) (*database.DeviceListResult, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * pageSize

	repo := database.NewDeviceRepository(a.db.DB())
	return repo.GetFiltered(database.DeviceFilter{
		Type:      filter.Type,
		Status:    filter.Status,
		Search:    filter.Search,
		Limit:     pageSize,
		Offset:    offset,
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
	})
}

// GetDevicesByType returns devices of a specific type
func (a *App) GetDevicesByType(deviceType string) ([]models.Device, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	repo := database.NewDeviceRepository(a.db.DB())
	return repo.GetByType(models.DeviceType(deviceType))
}

// GetDevice returns a device by ID with type-specific details
func (a *App) GetDevice(id int64) (*models.DeviceWithDetails, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, fmt.Errorf("device not found")
	}

	result := &models.DeviceWithDetails{Device: *device}

	// Get type-specific details
	switch device.Type {
	case models.DeviceTypeSwitch:
		switchRepo := database.NewSwitchRepository(a.db.DB())
		sw, err := switchRepo.GetByDeviceID(id)
		if err != nil {
			return nil, err
		}
		result.Switch = sw

		if sw != nil {
			ports, err := switchRepo.GetPorts(id)
			if err != nil {
				return nil, err
			}
			result.Ports = ports
		}

	case models.DeviceTypeCamera:
		cameraRepo := database.NewCameraRepository(a.db.DB())
		cam, err := cameraRepo.GetByDeviceID(id)
		if err != nil {
			return nil, err
		}
		result.Camera = cam

	case models.DeviceTypeServer:
		serverRepo := database.NewServerRepository(a.db.DB())
		srv, err := serverRepo.GetByDeviceID(id)
		if err != nil {
			return nil, err
		}
		result.Server = srv
	}

	return result, nil
}

// CreateDevice creates a new device
func (a *App) CreateDevice(input DeviceInput) (*models.Device, error) {
	log.Printf("CreateDevice called with input: %+v", input)

	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Validate input
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if input.IPAddress == "" {
		return nil, fmt.Errorf("IP address is required")
	}
	if input.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	device := &models.Device{
		Name:         input.Name,
		IPAddress:    input.IPAddress,
		Type:         models.DeviceType(input.Type),
		Manufacturer: input.Manufacturer,
		Model:        input.Model,
		CredentialID: input.CredentialID,
		Status:       models.DeviceStatusUnknown,
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	if err := deviceRepo.Create(device); err != nil {
		return nil, err
	}

	// Create type-specific data
	switch device.Type {
	case models.DeviceTypeSwitch:
		portCount := input.PortCount
		if portCount <= 0 {
			portCount = 8
		}
		sfpPortCount := input.SFPPortCount
		if sfpPortCount < 0 {
			sfpPortCount = 0
		}
		if sfpPortCount > portCount {
			sfpPortCount = 0
		}
		snmpVersion := input.SNMPVersion
		if snmpVersion == "" {
			snmpVersion = "v2c"
		}
		snmpv3Security := input.SNMPv3Security
		if snmpv3Security == "" {
			snmpv3Security = "noAuthNoPriv"
		}

		sw := &models.Switch{
			DeviceID:        device.ID,
			SNMPCommunity:   input.SNMPCommunity,
			SNMPVersion:     snmpVersion,
			PortCount:       portCount,
			SFPPortCount:    sfpPortCount,
			SNMPv3User:      input.SNMPv3User,
			SNMPv3Security:  snmpv3Security,
			SNMPv3AuthProto: input.SNMPv3AuthProto,
			SNMPv3AuthPass:  input.SNMPv3AuthPass,
			SNMPv3PrivProto: input.SNMPv3PrivProto,
			SNMPv3PrivPass:  input.SNMPv3PrivPass,
			UplinkSwitchID:  input.UplinkSwitchID,
			UplinkPortID:    input.UplinkPortID,
		}
		switchRepo := database.NewSwitchRepository(a.db.DB())
		if err := switchRepo.Create(sw); err != nil {
			// Rollback device creation
			deviceRepo.Delete(device.ID)
			return nil, err
		}

		// Link to parent switch port if uplink is set
		if input.UplinkPortID != nil {
			if err := switchRepo.LinkSwitch(*input.UplinkPortID, device.ID); err != nil {
				log.Printf("Warning: failed to link switch to uplink port: %v", err)
			}
		}

	case models.DeviceTypeCamera:
		// Camera must be linked to a switch port
		if input.SwitchPortID == nil {
			deviceRepo.Delete(device.ID)
			return nil, fmt.Errorf("camera must be linked to a switch port")
		}

		streamType := input.StreamType
		if streamType == "" {
			streamType = "jpeg"
		}
		onvifPort := input.ONVIFPort
		if onvifPort <= 0 {
			onvifPort = 80
		}

		cam := &models.Camera{
			DeviceID:    device.ID,
			RTSPURL:     input.RTSPURL,
			ONVIFPort:   onvifPort,
			SnapshotURL: input.SnapshotURL,
			StreamType:  streamType,
		}
		cameraRepo := database.NewCameraRepository(a.db.DB())
		if err := cameraRepo.Create(cam); err != nil {
			deviceRepo.Delete(device.ID)
			return nil, fmt.Errorf("failed to create camera: %w", err)
		}

		// Link camera to switch port
		switchRepo := database.NewSwitchRepository(a.db.DB())
		if err := switchRepo.LinkCamera(*input.SwitchPortID, device.ID); err != nil {
			// Rollback in correct order: camera first, then device
			cameraRepo.Delete(device.ID)
			deviceRepo.Delete(device.ID)
			return nil, fmt.Errorf("failed to link camera to port: %w", err)
		}

	case models.DeviceTypeServer:
		tcpPorts := input.TCPPorts
		if tcpPorts == "" {
			tcpPorts = "[]"
		}

		srv := &models.Server{
			DeviceID:       device.ID,
			TCPPorts:       tcpPorts,
			UseSNMP:        input.UseSNMP,
			UplinkSwitchID: input.UplinkSwitchID,
			UplinkPortID:   input.UplinkPortID,
		}
		serverRepo := database.NewServerRepository(a.db.DB())
		if err := serverRepo.Create(srv); err != nil {
			deviceRepo.Delete(device.ID)
			return nil, err
		}

		// Link to parent switch port if uplink is set
		if input.UplinkPortID != nil {
			switchRepo := database.NewSwitchRepository(a.db.DB())
			if err := switchRepo.LinkSwitch(*input.UplinkPortID, device.ID); err != nil {
				log.Printf("Warning: failed to link server to uplink port: %v", err)
			}
		}
	}

	return device, nil
}

// UpdateDevice updates an existing device
func (a *App) UpdateDevice(input DeviceInput) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if input.ID == 0 {
		return fmt.Errorf("device ID is required")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	existing, err := deviceRepo.GetByID(input.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("device not found")
	}

	// Update device fields
	existing.Name = input.Name
	existing.IPAddress = input.IPAddress
	existing.Manufacturer = input.Manufacturer
	existing.Model = input.Model
	existing.CredentialID = input.CredentialID

	if err := deviceRepo.Update(existing); err != nil {
		return err
	}

	// Update type-specific data
	switch existing.Type {
	case models.DeviceTypeSwitch:
		log.Printf("DEBUG UpdateDevice: SNMPVersion='%s', SNMPv3User='%s', SNMPv3Security='%s'",
			input.SNMPVersion, input.SNMPv3User, input.SNMPv3Security)
		snmpv3Security := input.SNMPv3Security
		if snmpv3Security == "" {
			snmpv3Security = "noAuthNoPriv"
		}
		sfpPortCount := input.SFPPortCount
		if sfpPortCount < 0 || sfpPortCount > input.PortCount {
			sfpPortCount = 0
		}

		// Get old uplink to manage port links
		switchRepo := database.NewSwitchRepository(a.db.DB())
		oldSwitch, _ := switchRepo.GetByDeviceID(existing.ID)
		var oldUplinkPortID *int64
		if oldSwitch != nil {
			oldUplinkPortID = oldSwitch.UplinkPortID
		}

		sw := &models.Switch{
			DeviceID:        existing.ID,
			SNMPCommunity:   input.SNMPCommunity,
			SNMPVersion:     input.SNMPVersion,
			PortCount:       input.PortCount,
			SFPPortCount:    sfpPortCount,
			SNMPv3User:      input.SNMPv3User,
			SNMPv3Security:  snmpv3Security,
			SNMPv3AuthProto: input.SNMPv3AuthProto,
			SNMPv3AuthPass:  input.SNMPv3AuthPass,
			SNMPv3PrivProto: input.SNMPv3PrivProto,
			SNMPv3PrivPass:  input.SNMPv3PrivPass,
			UplinkSwitchID:  input.UplinkSwitchID,
			UplinkPortID:    input.UplinkPortID,
		}
		if err := switchRepo.Update(sw); err != nil {
			return err
		}

		// Update port links if uplink changed
		if oldUplinkPortID != nil && (input.UplinkPortID == nil || *oldUplinkPortID != *input.UplinkPortID) {
			// Unlink from old port
			switchRepo.UnlinkSwitch(*oldUplinkPortID)
		}
		if input.UplinkPortID != nil && (oldUplinkPortID == nil || *oldUplinkPortID != *input.UplinkPortID) {
			// Link to new port
			switchRepo.LinkSwitch(*input.UplinkPortID, existing.ID)
		}

	case models.DeviceTypeCamera:
		cam := &models.Camera{
			DeviceID:    existing.ID,
			RTSPURL:     input.RTSPURL,
			ONVIFPort:   input.ONVIFPort,
			SnapshotURL: input.SnapshotURL,
			StreamType:  input.StreamType,
		}
		cameraRepo := database.NewCameraRepository(a.db.DB())
		if err := cameraRepo.Update(cam); err != nil {
			return err
		}

		// Update camera-to-port linking if port changed
		if input.SwitchPortID != nil {
			switchRepo := database.NewSwitchRepository(a.db.DB())
			// First, unlink from any existing port
			_, err := a.db.DB().Exec("UPDATE switch_ports SET linked_camera_id = NULL WHERE linked_camera_id = ?", existing.ID)
			if err != nil {
				return fmt.Errorf("failed to unlink camera from old port: %w", err)
			}
			// Link to new port
			if err := switchRepo.LinkCamera(*input.SwitchPortID, existing.ID); err != nil {
				return fmt.Errorf("failed to link camera to new port: %w", err)
			}
		}

	case models.DeviceTypeServer:
		// Get old uplink to manage port links
		serverRepo := database.NewServerRepository(a.db.DB())
		oldServer, _ := serverRepo.GetByDeviceID(existing.ID)
		var oldUplinkPortID *int64
		if oldServer != nil {
			oldUplinkPortID = oldServer.UplinkPortID
		}

		srv := &models.Server{
			DeviceID:       existing.ID,
			TCPPorts:       input.TCPPorts,
			UseSNMP:        input.UseSNMP,
			UplinkSwitchID: input.UplinkSwitchID,
			UplinkPortID:   input.UplinkPortID,
		}
		if err := serverRepo.Update(srv); err != nil {
			return err
		}

		// Update port links if uplink changed
		switchRepo := database.NewSwitchRepository(a.db.DB())
		if oldUplinkPortID != nil && (input.UplinkPortID == nil || *oldUplinkPortID != *input.UplinkPortID) {
			// Unlink from old port
			switchRepo.UnlinkSwitch(*oldUplinkPortID)
		}
		if input.UplinkPortID != nil && (oldUplinkPortID == nil || *oldUplinkPortID != *input.UplinkPortID) {
			// Link to new port
			switchRepo.LinkSwitch(*input.UplinkPortID, existing.ID)
		}
	}

	return nil
}

// DeleteDevice deletes a device by ID
func (a *App) DeleteDevice(id int64) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Get device to check type
	deviceRepo := database.NewDeviceRepository(a.db.DB())
	device, err := deviceRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}
	if device == nil {
		return fmt.Errorf("device not found")
	}

	// If camera, unlink from switch port first
	if device.Type == "camera" {
		_, err := a.db.DB().Exec("UPDATE switch_ports SET linked_camera_id = NULL WHERE linked_camera_id = ?", id)
		if err != nil {
			return fmt.Errorf("failed to unlink camera from port: %w", err)
		}
	}

	err = deviceRepo.Delete(id)
	if err != nil {
		return err
	}

	// Emit device deleted event so frontend can refresh
	runtime.EventsEmit(a.ctx, "device:deleted", map[string]interface{}{
		"device_id": id,
	})

	log.Printf("Device %d deleted", id)
	return nil
}

// GetDeviceStats returns device statistics
func (a *App) GetDeviceStats() (map[string]int, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	deviceRepo := database.NewDeviceRepository(a.db.DB())
	return deviceRepo.GetStats()
}
