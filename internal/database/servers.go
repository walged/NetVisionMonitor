package database

import (
	"database/sql"
	"fmt"

	"netvisionmonitor/internal/models"
)

// ServerRepository handles server-specific database operations
type ServerRepository struct {
	db *sql.DB
}

// NewServerRepository creates a new server repository
func NewServerRepository(db *sql.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

// Create inserts server-specific data
func (r *ServerRepository) Create(srv *models.Server) error {
	_, err := r.db.Exec(`
		INSERT INTO servers (device_id, tcp_ports, use_snmp, uplink_switch_id, uplink_port_id)
		VALUES (?, ?, ?, ?, ?)`,
		srv.DeviceID, srv.TCPPorts, srv.UseSNMP, srv.UplinkSwitchID, srv.UplinkPortID,
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	return nil
}

// GetByDeviceID retrieves server data by device ID
func (r *ServerRepository) GetByDeviceID(deviceID int64) (*models.Server, error) {
	srv := &models.Server{}
	var uplinkSwitchID, uplinkPortID sql.NullInt64

	err := r.db.QueryRow(`
		SELECT device_id, tcp_ports, use_snmp, uplink_switch_id, uplink_port_id
		FROM servers WHERE device_id = ?`, deviceID,
	).Scan(&srv.DeviceID, &srv.TCPPorts, &srv.UseSNMP, &uplinkSwitchID, &uplinkPortID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	if uplinkSwitchID.Valid {
		srv.UplinkSwitchID = &uplinkSwitchID.Int64
	}
	if uplinkPortID.Valid {
		srv.UplinkPortID = &uplinkPortID.Int64
	}

	return srv, nil
}

// Update updates server-specific data
func (r *ServerRepository) Update(srv *models.Server) error {
	_, err := r.db.Exec(`
		UPDATE servers SET tcp_ports = ?, use_snmp = ?, uplink_switch_id = ?, uplink_port_id = ?
		WHERE device_id = ?`,
		srv.TCPPorts, srv.UseSNMP, srv.UplinkSwitchID, srv.UplinkPortID, srv.DeviceID,
	)
	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}
	return nil
}

// Delete removes server data by device ID
func (r *ServerRepository) Delete(deviceID int64) error {
	_, err := r.db.Exec("DELETE FROM servers WHERE device_id = ?", deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}
	return nil
}
