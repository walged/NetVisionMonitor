package database

import (
	"database/sql"
	"fmt"
	"time"

	"netvisionmonitor/internal/models"
)

// DeviceRepository handles device database operations
type DeviceRepository struct {
	db *sql.DB
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(db *sql.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Create inserts a new device
func (r *DeviceRepository) Create(device *models.Device) error {
	result, err := r.db.Exec(`
		INSERT INTO devices (name, ip_address, type, manufacturer, model, credential_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		device.Name, device.IPAddress, device.Type, device.Manufacturer, device.Model,
		device.CredentialID, device.Status, time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	device.ID = id
	return nil
}

// GetByID retrieves a device by ID
func (r *DeviceRepository) GetByID(id int64) (*models.Device, error) {
	device := &models.Device{}
	err := r.db.QueryRow(`
		SELECT id, name, ip_address, type, COALESCE(manufacturer, ''), model, credential_id, status, last_check, created_at, updated_at
		FROM devices WHERE id = ?`, id,
	).Scan(
		&device.ID, &device.Name, &device.IPAddress, &device.Type, &device.Manufacturer, &device.Model,
		&device.CredentialID, &device.Status, &device.LastCheck, &device.CreatedAt, &device.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return device, nil
}

// GetAll retrieves all devices
func (r *DeviceRepository) GetAll() ([]models.Device, error) {
	rows, err := r.db.Query(`
		SELECT id, name, ip_address, type, COALESCE(manufacturer, ''), model, credential_id, status, last_check, created_at, updated_at
		FROM devices ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		err := rows.Scan(
			&d.ID, &d.Name, &d.IPAddress, &d.Type, &d.Manufacturer, &d.Model,
			&d.CredentialID, &d.Status, &d.LastCheck, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, nil
}

// DeviceFilter defines filter options for device queries
type DeviceFilter struct {
	Type       string
	Status     string
	Search     string
	Limit      int
	Offset     int
	SortBy     string
	SortOrder  string
}

// DeviceListResult contains paginated device list with total count
type DeviceListResult struct {
	Devices    []models.Device `json:"devices"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// GetFiltered retrieves devices with filtering and pagination
func (r *DeviceRepository) GetFiltered(filter DeviceFilter) (*DeviceListResult, error) {
	// Build WHERE clause
	where := "1=1"
	args := []interface{}{}

	if filter.Type != "" {
		where += " AND type = ?"
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		where += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		where += " AND (name LIKE ? OR ip_address LIKE ? OR model LIKE ?)"
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM devices WHERE " + where
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count devices: %w", err)
	}

	// Build ORDER BY
	sortBy := "name"
	if filter.SortBy != "" {
		// Whitelist allowed sort columns
		allowedSorts := map[string]bool{"name": true, "ip_address": true, "type": true, "status": true, "created_at": true}
		if allowedSorts[filter.SortBy] {
			sortBy = filter.SortBy
		}
	}
	sortOrder := "ASC"
	if filter.SortOrder == "desc" {
		sortOrder = "DESC"
	}

	// Apply pagination
	limit := 50
	if filter.Limit > 0 && filter.Limit <= 200 {
		limit = filter.Limit
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, name, ip_address, type, COALESCE(manufacturer, ''), model, credential_id, status, last_check, created_at, updated_at
		FROM devices WHERE %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		where, sortBy, sortOrder)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		err := rows.Scan(
			&d.ID, &d.Name, &d.IPAddress, &d.Type, &d.Manufacturer, &d.Model,
			&d.CredentialID, &d.Status, &d.LastCheck, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, d)
	}

	page := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit

	return &DeviceListResult{
		Devices:    devices,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}, nil
}

// GetByType retrieves devices by type
func (r *DeviceRepository) GetByType(deviceType models.DeviceType) ([]models.Device, error) {
	rows, err := r.db.Query(`
		SELECT id, name, ip_address, type, COALESCE(manufacturer, ''), model, credential_id, status, last_check, created_at, updated_at
		FROM devices WHERE type = ? ORDER BY name`, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		err := rows.Scan(
			&d.ID, &d.Name, &d.IPAddress, &d.Type, &d.Manufacturer, &d.Model,
			&d.CredentialID, &d.Status, &d.LastCheck, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, nil
}

// Update updates an existing device
func (r *DeviceRepository) Update(device *models.Device) error {
	device.UpdatedAt = time.Now()
	_, err := r.db.Exec(`
		UPDATE devices SET name = ?, ip_address = ?, type = ?, manufacturer = ?, model = ?,
		credential_id = ?, status = ?, updated_at = ?
		WHERE id = ?`,
		device.Name, device.IPAddress, device.Type, device.Manufacturer, device.Model,
		device.CredentialID, device.Status, device.UpdatedAt, device.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}
	return nil
}

// UpdateStatus updates device status and last_check time
func (r *DeviceRepository) UpdateStatus(id int64, status models.DeviceStatus) error {
	_, err := r.db.Exec(`
		UPDATE devices SET status = ?, last_check = ?, updated_at = ?
		WHERE id = ?`,
		status, time.Now(), time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update device status: %w", err)
	}
	return nil
}

// Delete removes a device by ID
func (r *DeviceRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM devices WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}
	return nil
}

// Count returns the total number of devices
func (r *DeviceRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM devices").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count devices: %w", err)
	}
	return count, nil
}

// CountByType returns the number of devices by type
func (r *DeviceRepository) CountByType(deviceType models.DeviceType) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM devices WHERE type = ?", deviceType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count devices: %w", err)
	}
	return count, nil
}

// CountByStatus returns the number of devices by status
func (r *DeviceRepository) CountByStatus(status models.DeviceStatus) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM devices WHERE status = ?", status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count devices: %w", err)
	}
	return count, nil
}

// GetStats returns device statistics
func (r *DeviceRepository) GetStats() (map[string]int, error) {
	stats := make(map[string]int)

	// Total count
	total, err := r.Count()
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Count by type
	for _, t := range []models.DeviceType{models.DeviceTypeSwitch, models.DeviceTypeServer, models.DeviceTypeCamera} {
		count, err := r.CountByType(t)
		if err != nil {
			return nil, err
		}
		stats[string(t)] = count
	}

	// Count by status
	for _, s := range []models.DeviceStatus{models.DeviceStatusOnline, models.DeviceStatusOffline, models.DeviceStatusUnknown} {
		count, err := r.CountByStatus(s)
		if err != nil {
			return nil, err
		}
		stats[string(s)] = count
	}

	return stats, nil
}

// GetSwitchPorts returns all ports for a switch
func (r *DeviceRepository) GetSwitchPorts(switchID int64) ([]models.SwitchPort, error) {
	rows, err := r.db.Query(`
		SELECT id, switch_id, port_number, name, status, COALESCE(speed, ''), linked_camera_id
		FROM switch_ports
		WHERE switch_id = ?
		ORDER BY port_number`, switchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query switch ports: %w", err)
	}
	defer rows.Close()

	var ports []models.SwitchPort
	for rows.Next() {
		var p models.SwitchPort
		var linkedCameraID sql.NullInt64
		err := rows.Scan(&p.ID, &p.SwitchID, &p.PortNumber, &p.Name, &p.Status, &p.Speed, &linkedCameraID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan switch port: %w", err)
		}
		if linkedCameraID.Valid {
			p.LinkedCameraID = &linkedCameraID.Int64
		}
		ports = append(ports, p)
	}
	return ports, nil
}

// UpdateSwitchPort updates a switch port
func (r *DeviceRepository) UpdateSwitchPort(port *models.SwitchPort) error {
	_, err := r.db.Exec(`
		UPDATE switch_ports SET name = ?, status = ?, speed = ?, linked_camera_id = ?
		WHERE id = ?`,
		port.Name, port.Status, port.Speed, port.LinkedCameraID, port.ID)
	if err != nil {
		return fmt.Errorf("failed to update switch port: %w", err)
	}
	return nil
}

// LinkCameraToPort links a camera to a switch port
func (r *DeviceRepository) LinkCameraToPort(portID int64, cameraID *int64) error {
	_, err := r.db.Exec(`UPDATE switch_ports SET linked_camera_id = ? WHERE id = ?`, cameraID, portID)
	if err != nil {
		return fmt.Errorf("failed to link camera to port: %w", err)
	}
	return nil
}

// CreateSwitchPorts creates initial ports for a switch
func (r *DeviceRepository) CreateSwitchPorts(switchID int64, portCount int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO switch_ports (switch_id, port_number, name, status)
		VALUES (?, ?, ?, 'unknown')`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= portCount; i++ {
		_, err := stmt.Exec(switchID, i, fmt.Sprintf("Port %d", i))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdatePortStatus updates the status of a switch port
func (r *DeviceRepository) UpdatePortStatus(switchID int64, portNumber int, status string, speed string) error {
	_, err := r.db.Exec(`
		UPDATE switch_ports SET status = ?, speed = ?
		WHERE switch_id = ? AND port_number = ?`,
		status, speed, switchID, portNumber)
	return err
}
