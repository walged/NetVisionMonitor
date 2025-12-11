package database

import (
	"database/sql"
	"fmt"

	"netvisionmonitor/internal/encryption"
	"netvisionmonitor/internal/models"
)

// SwitchRepository handles switch-specific database operations
type SwitchRepository struct {
	db *sql.DB
}

// NewSwitchRepository creates a new switch repository
func NewSwitchRepository(db *sql.DB) *SwitchRepository {
	return &SwitchRepository{db: db}
}

// Create inserts switch-specific data
func (r *SwitchRepository) Create(sw *models.Switch) error {
	encryptedCommunity, err := encryption.EncryptIfNotEmpty(sw.SNMPCommunity)
	if err != nil {
		return fmt.Errorf("failed to encrypt SNMP community: %w", err)
	}

	encryptedAuthPass, err := encryption.EncryptIfNotEmpty(sw.SNMPv3AuthPass)
	if err != nil {
		return fmt.Errorf("failed to encrypt SNMPv3 auth password: %w", err)
	}

	encryptedPrivPass, err := encryption.EncryptIfNotEmpty(sw.SNMPv3PrivPass)
	if err != nil {
		return fmt.Errorf("failed to encrypt SNMPv3 priv password: %w", err)
	}

	_, err = r.db.Exec(`
		INSERT INTO switches (device_id, snmp_community, snmp_version, port_count, sfp_port_count,
			snmpv3_user, snmpv3_security, snmpv3_auth_proto, snmpv3_auth_pass, snmpv3_priv_proto, snmpv3_priv_pass,
			uplink_switch_id, uplink_port_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sw.DeviceID, encryptedCommunity, sw.SNMPVersion, sw.PortCount, sw.SFPPortCount,
		sw.SNMPv3User, sw.SNMPv3Security, sw.SNMPv3AuthProto, encryptedAuthPass, sw.SNMPv3PrivProto, encryptedPrivPass,
		sw.UplinkSwitchID, sw.UplinkPortID,
	)
	if err != nil {
		return fmt.Errorf("failed to create switch: %w", err)
	}

	// Create ports - last sfp_port_count ports are SFP
	copperPorts := sw.PortCount - sw.SFPPortCount
	for i := 1; i <= sw.PortCount; i++ {
		portType := "copper"
		if i > copperPorts {
			portType = "sfp"
		}
		_, err = r.db.Exec(`
			INSERT INTO switch_ports (switch_id, port_number, name, status, port_type)
			VALUES (?, ?, ?, 'unknown', ?)`,
			sw.DeviceID, i, fmt.Sprintf("Port %d", i), portType,
		)
		if err != nil {
			return fmt.Errorf("failed to create port %d: %w", i, err)
		}
	}

	return nil
}

// GetByDeviceID retrieves switch data by device ID
func (r *SwitchRepository) GetByDeviceID(deviceID int64) (*models.Switch, error) {
	sw := &models.Switch{}
	var encryptedCommunity, encryptedAuthPass, encryptedPrivPass string
	var snmpv3User, snmpv3Security, snmpv3AuthProto, snmpv3PrivProto sql.NullString
	var uplinkSwitchID, uplinkPortID sql.NullInt64

	err := r.db.QueryRow(`
		SELECT device_id, snmp_community, snmp_version, port_count, COALESCE(sfp_port_count, 0),
			COALESCE(snmpv3_user, ''), COALESCE(snmpv3_security, 'noAuthNoPriv'),
			COALESCE(snmpv3_auth_proto, ''), COALESCE(snmpv3_auth_pass, ''),
			COALESCE(snmpv3_priv_proto, ''), COALESCE(snmpv3_priv_pass, ''),
			uplink_switch_id, uplink_port_id
		FROM switches WHERE device_id = ?`, deviceID,
	).Scan(&sw.DeviceID, &encryptedCommunity, &sw.SNMPVersion, &sw.PortCount, &sw.SFPPortCount,
		&snmpv3User, &snmpv3Security, &snmpv3AuthProto, &encryptedAuthPass, &snmpv3PrivProto, &encryptedPrivPass,
		&uplinkSwitchID, &uplinkPortID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get switch: %w", err)
	}

	sw.SNMPCommunity, err = encryption.DecryptIfNotEmpty(encryptedCommunity)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SNMP community: %w", err)
	}

	sw.SNMPv3AuthPass, err = encryption.DecryptIfNotEmpty(encryptedAuthPass)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SNMPv3 auth password: %w", err)
	}

	sw.SNMPv3PrivPass, err = encryption.DecryptIfNotEmpty(encryptedPrivPass)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SNMPv3 priv password: %w", err)
	}

	if snmpv3User.Valid {
		sw.SNMPv3User = snmpv3User.String
	}
	if snmpv3Security.Valid {
		sw.SNMPv3Security = snmpv3Security.String
	}
	if snmpv3AuthProto.Valid {
		sw.SNMPv3AuthProto = snmpv3AuthProto.String
	}
	if snmpv3PrivProto.Valid {
		sw.SNMPv3PrivProto = snmpv3PrivProto.String
	}
	if uplinkSwitchID.Valid {
		sw.UplinkSwitchID = &uplinkSwitchID.Int64
	}
	if uplinkPortID.Valid {
		sw.UplinkPortID = &uplinkPortID.Int64
	}

	return sw, nil
}

// Update updates switch-specific data
func (r *SwitchRepository) Update(sw *models.Switch) error {
	encryptedCommunity, err := encryption.EncryptIfNotEmpty(sw.SNMPCommunity)
	if err != nil {
		return fmt.Errorf("failed to encrypt SNMP community: %w", err)
	}

	encryptedAuthPass, err := encryption.EncryptIfNotEmpty(sw.SNMPv3AuthPass)
	if err != nil {
		return fmt.Errorf("failed to encrypt SNMPv3 auth password: %w", err)
	}

	encryptedPrivPass, err := encryption.EncryptIfNotEmpty(sw.SNMPv3PrivPass)
	if err != nil {
		return fmt.Errorf("failed to encrypt SNMPv3 priv password: %w", err)
	}

	// Get current port count
	var currentPortCount int
	err = r.db.QueryRow("SELECT port_count FROM switches WHERE device_id = ?", sw.DeviceID).Scan(&currentPortCount)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get current port count: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE switches SET snmp_community = ?, snmp_version = ?, port_count = ?, sfp_port_count = ?,
			snmpv3_user = ?, snmpv3_security = ?, snmpv3_auth_proto = ?, snmpv3_auth_pass = ?,
			snmpv3_priv_proto = ?, snmpv3_priv_pass = ?,
			uplink_switch_id = ?, uplink_port_id = ?
		WHERE device_id = ?`,
		encryptedCommunity, sw.SNMPVersion, sw.PortCount, sw.SFPPortCount,
		sw.SNMPv3User, sw.SNMPv3Security, sw.SNMPv3AuthProto, encryptedAuthPass,
		sw.SNMPv3PrivProto, encryptedPrivPass,
		sw.UplinkSwitchID, sw.UplinkPortID, sw.DeviceID,
	)
	if err != nil {
		return fmt.Errorf("failed to update switch: %w", err)
	}

	// Sync ports if port count changed
	if sw.PortCount != currentPortCount {
		// Add new ports if count increased
		copperPorts := sw.PortCount - sw.SFPPortCount
		if sw.PortCount > currentPortCount {
			for i := currentPortCount + 1; i <= sw.PortCount; i++ {
				portType := "copper"
				if i > copperPorts {
					portType = "sfp"
				}
				_, err = r.db.Exec(`
					INSERT INTO switch_ports (switch_id, port_number, name, status, port_type)
					VALUES (?, ?, ?, 'unknown', ?)`,
					sw.DeviceID, i, fmt.Sprintf("Port %d", i), portType,
				)
				if err != nil {
					return fmt.Errorf("failed to create port %d: %w", i, err)
				}
			}
		} else {
			// Remove extra ports if count decreased (keep camera links by unlinking first)
			_, err = r.db.Exec(`
				UPDATE switch_ports SET linked_camera_id = NULL, linked_switch_id = NULL
				WHERE switch_id = ? AND port_number > ?`,
				sw.DeviceID, sw.PortCount,
			)
			if err != nil {
				return fmt.Errorf("failed to unlink devices from removed ports: %w", err)
			}
			_, err = r.db.Exec(`
				DELETE FROM switch_ports
				WHERE switch_id = ? AND port_number > ?`,
				sw.DeviceID, sw.PortCount,
			)
			if err != nil {
				return fmt.Errorf("failed to delete extra ports: %w", err)
			}
		}
	}

	// Update port types based on SFP count
	copperPorts := sw.PortCount - sw.SFPPortCount
	_, err = r.db.Exec(`
		UPDATE switch_ports SET port_type = 'copper'
		WHERE switch_id = ? AND port_number <= ?`,
		sw.DeviceID, copperPorts,
	)
	if err != nil {
		return fmt.Errorf("failed to update copper ports: %w", err)
	}
	_, err = r.db.Exec(`
		UPDATE switch_ports SET port_type = 'sfp', linked_camera_id = NULL
		WHERE switch_id = ? AND port_number > ?`,
		sw.DeviceID, copperPorts,
	)
	if err != nil {
		return fmt.Errorf("failed to update SFP ports: %w", err)
	}

	return nil
}

// Delete removes switch data by device ID
func (r *SwitchRepository) Delete(deviceID int64) error {
	_, err := r.db.Exec("DELETE FROM switches WHERE device_id = ?", deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete switch: %w", err)
	}
	return nil
}

// GetPorts retrieves all ports for a switch
func (r *SwitchRepository) GetPorts(switchID int64) ([]models.SwitchPort, error) {
	rows, err := r.db.Query(`
		SELECT id, switch_id, port_number, name, status, COALESCE(speed, ''),
			COALESCE(port_type, 'copper'), linked_camera_id, linked_switch_id
		FROM switch_ports WHERE switch_id = ? ORDER BY port_number`, switchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ports: %w", err)
	}
	defer rows.Close()

	var ports []models.SwitchPort
	for rows.Next() {
		var p models.SwitchPort
		var linkedCameraID, linkedSwitchID sql.NullInt64
		err := rows.Scan(&p.ID, &p.SwitchID, &p.PortNumber, &p.Name, &p.Status, &p.Speed,
			&p.PortType, &linkedCameraID, &linkedSwitchID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan port: %w", err)
		}
		if linkedCameraID.Valid {
			p.LinkedCameraID = &linkedCameraID.Int64
		}
		if linkedSwitchID.Valid {
			p.LinkedSwitchID = &linkedSwitchID.Int64
		}
		ports = append(ports, p)
	}
	return ports, nil
}

// UpdatePort updates a switch port
func (r *SwitchRepository) UpdatePort(port *models.SwitchPort) error {
	_, err := r.db.Exec(`
		UPDATE switch_ports SET name = ?, status = ?, speed = ?, port_type = ?,
			linked_camera_id = ?, linked_switch_id = ?
		WHERE id = ?`,
		port.Name, port.Status, port.Speed, port.PortType,
		port.LinkedCameraID, port.LinkedSwitchID, port.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update port: %w", err)
	}
	return nil
}

// UpdatePortStatus updates port status
func (r *SwitchRepository) UpdatePortStatus(portID int64, status string) error {
	_, err := r.db.Exec("UPDATE switch_ports SET status = ? WHERE id = ?", status, portID)
	if err != nil {
		return fmt.Errorf("failed to update port status: %w", err)
	}
	return nil
}

// LinkCamera links a camera to a port
func (r *SwitchRepository) LinkCamera(portID, cameraID int64) error {
	_, err := r.db.Exec("UPDATE switch_ports SET linked_camera_id = ? WHERE id = ?", cameraID, portID)
	if err != nil {
		return fmt.Errorf("failed to link camera: %w", err)
	}
	return nil
}

// UnlinkCamera unlinks a camera from a port
func (r *SwitchRepository) UnlinkCamera(portID int64) error {
	_, err := r.db.Exec("UPDATE switch_ports SET linked_camera_id = NULL WHERE id = ?", portID)
	if err != nil {
		return fmt.Errorf("failed to unlink camera: %w", err)
	}
	return nil
}

// LinkSwitch links a switch to an SFP port (uplink)
func (r *SwitchRepository) LinkSwitch(portID, switchID int64) error {
	_, err := r.db.Exec("UPDATE switch_ports SET linked_switch_id = ? WHERE id = ?", switchID, portID)
	if err != nil {
		return fmt.Errorf("failed to link switch: %w", err)
	}
	return nil
}

// UnlinkSwitch unlinks a switch from an SFP port
func (r *SwitchRepository) UnlinkSwitch(portID int64) error {
	_, err := r.db.Exec("UPDATE switch_ports SET linked_switch_id = NULL WHERE id = ?", portID)
	if err != nil {
		return fmt.Errorf("failed to unlink switch: %w", err)
	}
	return nil
}
