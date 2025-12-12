package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/encryption"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// BackupData represents the exported data structure
type BackupData struct {
	Version     string    `json:"version"`
	AppVersion  string    `json:"app_version"`
	ExportDate  time.Time `json:"export_date"`
	Description string    `json:"description"`

	// Configuration data
	Credentials []CredentialExport `json:"credentials"`
	Devices     []DeviceExport     `json:"devices"`
	Switches    []SwitchExport     `json:"switches"`
	SwitchPorts []SwitchPortExport `json:"switch_ports"`
	Cameras     []CameraExport     `json:"cameras"`
	Servers     []ServerExport     `json:"servers"`
	Schemas     []SchemaExport     `json:"schemas"`
	SchemaItems []SchemaItemExport `json:"schema_items"`
	Settings    map[string]string  `json:"settings"`
}

// Export structures (with decrypted sensitive data for portability)
type CredentialExport struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type DeviceExport struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	IPAddress    string `json:"ip_address"`
	Type         string `json:"type"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	CredentialID *int64 `json:"credential_id,omitempty"`
}

type SwitchExport struct {
	DeviceID        int64  `json:"device_id"`
	SNMPCommunity   string `json:"snmp_community"`
	SNMPVersion     string `json:"snmp_version"`
	PortCount       int    `json:"port_count"`
	SFPPortCount    int    `json:"sfp_port_count"`
	SNMPv3User      string `json:"snmpv3_user,omitempty"`
	SNMPv3Security  string `json:"snmpv3_security,omitempty"`
	SNMPv3AuthProto string `json:"snmpv3_auth_proto,omitempty"`
	SNMPv3AuthPass  string `json:"snmpv3_auth_pass,omitempty"`
	SNMPv3PrivProto string `json:"snmpv3_priv_proto,omitempty"`
	SNMPv3PrivPass  string `json:"snmpv3_priv_pass,omitempty"`
}

type SwitchPortExport struct {
	ID             int64  `json:"id"`
	SwitchID       int64  `json:"switch_id"`
	PortNumber     int    `json:"port_number"`
	Name           string `json:"name"`
	PortType       string `json:"port_type"`
	LinkedCameraID *int64 `json:"linked_camera_id,omitempty"`
	LinkedSwitchID *int64 `json:"linked_switch_id,omitempty"`
}

type CameraExport struct {
	DeviceID    int64  `json:"device_id"`
	RTSPURL     string `json:"rtsp_url"`
	ONVIFPort   int    `json:"onvif_port"`
	SnapshotURL string `json:"snapshot_url"`
	StreamType  string `json:"stream_type"`
}

type ServerExport struct {
	DeviceID int64  `json:"device_id"`
	TCPPorts string `json:"tcp_ports"`
	UseSNMP  bool   `json:"use_snmp"`
}

type SchemaExport struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	BackgroundImage string    `json:"background_image"`
	CreatedAt       time.Time `json:"created_at"`
}

type SchemaItemExport struct {
	ID       int64   `json:"id"`
	DeviceID int64   `json:"device_id"`
	SchemaID int64   `json:"schema_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
}

// ExportConfiguration exports all configuration data to a JSON file
func (a *App) ExportConfiguration() (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Ask user for save location
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Экспорт конфигурации",
		DefaultFilename: fmt.Sprintf("netvision_config_%s.json", time.Now().Format("2006-01-02")),
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files", Pattern: "*.json"},
		},
	})
	if err != nil {
		return "", err
	}
	if savePath == "" {
		return "", nil // User cancelled
	}

	// Collect all data
	backup, err := a.collectBackupData()
	if err != nil {
		return "", fmt.Errorf("failed to collect backup data: %w", err)
	}

	// Write JSON file
	dataJSON, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal backup data: %w", err)
	}

	if err := os.WriteFile(savePath, dataJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write JSON file: %w", err)
	}

	return savePath, nil
}

// collectBackupData gathers all configuration data from the database
func (a *App) collectBackupData() (*BackupData, error) {
	backup := &BackupData{
		Version:     "1.0",
		AppVersion:  "1.1.0",
		ExportDate:  time.Now(),
		Description: "NetVisionMonitor configuration backup",
		Settings:    make(map[string]string),
	}

	db := a.db.DB()

	// Export credentials with decrypted passwords
	credRepo := database.NewCredentialRepository(db)
	creds, err := a.getAllCredentialsWithPasswords(credRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to export credentials: %w", err)
	}
	backup.Credentials = creds

	// Export devices
	devices, err := a.exportDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to export devices: %w", err)
	}
	backup.Devices = devices

	// Export switches with decrypted SNMP data
	switches, err := a.exportSwitches()
	if err != nil {
		return nil, fmt.Errorf("failed to export switches: %w", err)
	}
	backup.Switches = switches

	// Export switch ports
	ports, err := a.exportSwitchPorts()
	if err != nil {
		return nil, fmt.Errorf("failed to export switch ports: %w", err)
	}
	backup.SwitchPorts = ports

	// Export cameras with decrypted URLs
	cameras, err := a.exportCameras()
	if err != nil {
		return nil, fmt.Errorf("failed to export cameras: %w", err)
	}
	backup.Cameras = cameras

	// Export servers
	servers, err := a.exportServers()
	if err != nil {
		return nil, fmt.Errorf("failed to export servers: %w", err)
	}
	backup.Servers = servers

	// Export schemas
	schemas, err := a.exportSchemas()
	if err != nil {
		return nil, fmt.Errorf("failed to export schemas: %w", err)
	}
	backup.Schemas = schemas

	// Export schema items
	schemaItems, err := a.exportSchemaItems()
	if err != nil {
		return nil, fmt.Errorf("failed to export schema items: %w", err)
	}
	backup.SchemaItems = schemaItems

	// Export settings
	settingsRepo := database.NewSettingsRepository(db)
	settings, err := settingsRepo.GetAll()
	if err == nil {
		backup.Settings = settings
	}

	return backup, nil
}

func (a *App) getAllCredentialsWithPasswords(repo *database.CredentialRepository) ([]CredentialExport, error) {
	rows, err := a.db.DB().Query(`
		SELECT id, name, type, username, password, note, created_at
		FROM credentials ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []CredentialExport
	for rows.Next() {
		var c CredentialExport
		var encUser, encPass string
		err := rows.Scan(&c.ID, &c.Name, &c.Type, &encUser, &encPass, &c.Note, &c.CreatedAt)
		if err != nil {
			continue
		}
		c.Username, _ = encryption.DecryptIfNotEmpty(encUser)
		c.Password, _ = encryption.DecryptIfNotEmpty(encPass)
		creds = append(creds, c)
	}
	return creds, nil
}

func (a *App) exportDevices() ([]DeviceExport, error) {
	rows, err := a.db.DB().Query(`
		SELECT id, name, ip_address, type, COALESCE(manufacturer, ''), COALESCE(model, ''), credential_id
		FROM devices ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []DeviceExport
	for rows.Next() {
		var d DeviceExport
		var credID *int64
		err := rows.Scan(&d.ID, &d.Name, &d.IPAddress, &d.Type, &d.Manufacturer, &d.Model, &credID)
		if err != nil {
			continue
		}
		d.CredentialID = credID
		devices = append(devices, d)
	}
	return devices, nil
}

func (a *App) exportSwitches() ([]SwitchExport, error) {
	rows, err := a.db.DB().Query(`
		SELECT device_id, snmp_community, snmp_version, port_count, COALESCE(sfp_port_count, 0),
			COALESCE(snmpv3_user, ''), COALESCE(snmpv3_security, ''),
			COALESCE(snmpv3_auth_proto, ''), COALESCE(snmpv3_auth_pass, ''),
			COALESCE(snmpv3_priv_proto, ''), COALESCE(snmpv3_priv_pass, '')
		FROM switches ORDER BY device_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var switches []SwitchExport
	for rows.Next() {
		var s SwitchExport
		var encCommunity, encAuthPass, encPrivPass string
		err := rows.Scan(&s.DeviceID, &encCommunity, &s.SNMPVersion, &s.PortCount, &s.SFPPortCount,
			&s.SNMPv3User, &s.SNMPv3Security, &s.SNMPv3AuthProto, &encAuthPass, &s.SNMPv3PrivProto, &encPrivPass)
		if err != nil {
			continue
		}
		s.SNMPCommunity, _ = encryption.DecryptIfNotEmpty(encCommunity)
		s.SNMPv3AuthPass, _ = encryption.DecryptIfNotEmpty(encAuthPass)
		s.SNMPv3PrivPass, _ = encryption.DecryptIfNotEmpty(encPrivPass)
		switches = append(switches, s)
	}
	return switches, nil
}

func (a *App) exportSwitchPorts() ([]SwitchPortExport, error) {
	rows, err := a.db.DB().Query(`
		SELECT id, switch_id, port_number, name, COALESCE(port_type, 'copper'), linked_camera_id, linked_switch_id
		FROM switch_ports ORDER BY switch_id, port_number`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []SwitchPortExport
	for rows.Next() {
		var p SwitchPortExport
		err := rows.Scan(&p.ID, &p.SwitchID, &p.PortNumber, &p.Name, &p.PortType, &p.LinkedCameraID, &p.LinkedSwitchID)
		if err != nil {
			continue
		}
		ports = append(ports, p)
	}
	return ports, nil
}

func (a *App) exportCameras() ([]CameraExport, error) {
	rows, err := a.db.DB().Query(`
		SELECT device_id, COALESCE(rtsp_url, ''), COALESCE(onvif_port, 80),
			COALESCE(snapshot_url, ''), COALESCE(stream_type, 'jpeg')
		FROM cameras ORDER BY device_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cameras []CameraExport
	for rows.Next() {
		var c CameraExport
		var encRTSP string
		err := rows.Scan(&c.DeviceID, &encRTSP, &c.ONVIFPort, &c.SnapshotURL, &c.StreamType)
		if err != nil {
			continue
		}
		c.RTSPURL, _ = encryption.DecryptIfNotEmpty(encRTSP)
		cameras = append(cameras, c)
	}
	return cameras, nil
}

func (a *App) exportServers() ([]ServerExport, error) {
	rows, err := a.db.DB().Query(`
		SELECT device_id, COALESCE(tcp_ports, '[]'), COALESCE(use_snmp, 0)
		FROM servers ORDER BY device_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []ServerExport
	for rows.Next() {
		var s ServerExport
		var useSNMP int
		err := rows.Scan(&s.DeviceID, &s.TCPPorts, &useSNMP)
		if err != nil {
			continue
		}
		s.UseSNMP = useSNMP == 1
		servers = append(servers, s)
	}
	return servers, nil
}

func (a *App) exportSchemas() ([]SchemaExport, error) {
	rows, err := a.db.DB().Query(`
		SELECT id, name, COALESCE(background_image, ''), created_at
		FROM schemas ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []SchemaExport
	for rows.Next() {
		var s SchemaExport
		err := rows.Scan(&s.ID, &s.Name, &s.BackgroundImage, &s.CreatedAt)
		if err != nil {
			continue
		}
		schemas = append(schemas, s)
	}
	return schemas, nil
}

func (a *App) exportSchemaItems() ([]SchemaItemExport, error) {
	rows, err := a.db.DB().Query(`
		SELECT id, device_id, schema_id, x, y, width, height
		FROM schema_items ORDER BY schema_id, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SchemaItemExport
	for rows.Next() {
		var i SchemaItemExport
		err := rows.Scan(&i.ID, &i.DeviceID, &i.SchemaID, &i.X, &i.Y, &i.Width, &i.Height)
		if err != nil {
			continue
		}
		items = append(items, i)
	}
	return items, nil
}

// ImportConfiguration imports configuration from a JSON file
func (a *App) ImportConfiguration() (bool, error) {
	if a.db == nil {
		return false, fmt.Errorf("database not initialized")
	}

	// Ask user for file
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Импорт конфигурации",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files", Pattern: "*.json"},
		},
	})
	if err != nil {
		return false, err
	}
	if filePath == "" {
		return false, nil // User cancelled
	}

	// Read JSON file
	configData, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var backup BackupData
	if err := json.Unmarshal(configData, &backup); err != nil {
		return false, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Confirm import
	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Подтверждение импорта",
		Message:       "Текущие данные будут заменены данными из резервной копии.\n\nПродолжить?",
		Buttons:       []string{"Да", "Нет"},
		DefaultButton: "Нет",
	})
	if err != nil || result != "Да" {
		return false, nil
	}

	// Stop monitoring
	if a.monitor != nil {
		a.monitor.Stop()
	}

	// Import data
	if err := a.importBackupData(&backup); err != nil {
		// Restart monitoring on error
		a.monitor.Start()
		return false, fmt.Errorf("failed to import data: %w", err)
	}

	// Fix port types after import
	a.db.FixExistingPortTypes()

	// Restart monitoring
	a.initMonitoring()
	a.monitor.Start()

	return true, nil
}

// importBackupData imports all data from backup into the database
func (a *App) importBackupData(backup *BackupData) error {
	db := a.db.DB()

	// Clear existing data (in reverse dependency order)
	tables := []string{
		"schema_items", "schemas", "switch_ports", "cameras", "servers", "switches", "devices", "credentials",
	}
	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table)
		if err != nil {
			return fmt.Errorf("failed to clear %s: %w", table, err)
		}
	}

	// Import credentials
	for _, c := range backup.Credentials {
		encUser, _ := encryption.EncryptIfNotEmpty(c.Username)
		encPass, _ := encryption.EncryptIfNotEmpty(c.Password)
		_, err := db.Exec(`
			INSERT INTO credentials (id, name, type, username, password, note, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			c.ID, c.Name, c.Type, encUser, encPass, c.Note, c.CreatedAt, time.Now())
		if err != nil {
			return fmt.Errorf("failed to import credential %s: %w", c.Name, err)
		}
	}

	// Import devices
	for _, d := range backup.Devices {
		_, err := db.Exec(`
			INSERT INTO devices (id, name, ip_address, type, manufacturer, model, credential_id, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, 'unknown', ?, ?)`,
			d.ID, d.Name, d.IPAddress, d.Type, d.Manufacturer, d.Model, d.CredentialID, time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to import device %s: %w", d.Name, err)
		}
	}

	// Import switches
	for _, s := range backup.Switches {
		encCommunity, _ := encryption.EncryptIfNotEmpty(s.SNMPCommunity)
		encAuthPass, _ := encryption.EncryptIfNotEmpty(s.SNMPv3AuthPass)
		encPrivPass, _ := encryption.EncryptIfNotEmpty(s.SNMPv3PrivPass)
		_, err := db.Exec(`
			INSERT INTO switches (device_id, snmp_community, snmp_version, port_count, sfp_port_count,
				snmpv3_user, snmpv3_security, snmpv3_auth_proto, snmpv3_auth_pass, snmpv3_priv_proto, snmpv3_priv_pass)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.DeviceID, encCommunity, s.SNMPVersion, s.PortCount, s.SFPPortCount,
			s.SNMPv3User, s.SNMPv3Security, s.SNMPv3AuthProto, encAuthPass, s.SNMPv3PrivProto, encPrivPass)
		if err != nil {
			return fmt.Errorf("failed to import switch %d: %w", s.DeviceID, err)
		}
	}

	// Import cameras
	for _, c := range backup.Cameras {
		encRTSP, _ := encryption.EncryptIfNotEmpty(c.RTSPURL)
		_, err := db.Exec(`
			INSERT INTO cameras (device_id, rtsp_url, onvif_port, snapshot_url, stream_type)
			VALUES (?, ?, ?, ?, ?)`,
			c.DeviceID, encRTSP, c.ONVIFPort, c.SnapshotURL, c.StreamType)
		if err != nil {
			return fmt.Errorf("failed to import camera %d: %w", c.DeviceID, err)
		}
	}

	// Import servers
	for _, s := range backup.Servers {
		useSNMP := 0
		if s.UseSNMP {
			useSNMP = 1
		}
		_, err := db.Exec(`
			INSERT INTO servers (device_id, tcp_ports, use_snmp)
			VALUES (?, ?, ?)`,
			s.DeviceID, s.TCPPorts, useSNMP)
		if err != nil {
			return fmt.Errorf("failed to import server %d: %w", s.DeviceID, err)
		}
	}

	// Import switch ports
	for _, p := range backup.SwitchPorts {
		_, err := db.Exec(`
			INSERT INTO switch_ports (id, switch_id, port_number, name, status, port_type, linked_camera_id, linked_switch_id)
			VALUES (?, ?, ?, ?, 'unknown', ?, ?, ?)`,
			p.ID, p.SwitchID, p.PortNumber, p.Name, p.PortType, p.LinkedCameraID, p.LinkedSwitchID)
		if err != nil {
			return fmt.Errorf("failed to import port %d: %w", p.ID, err)
		}
	}

	// Import schemas
	for _, s := range backup.Schemas {
		_, err := db.Exec(`
			INSERT INTO schemas (id, name, background_image, created_at)
			VALUES (?, ?, ?, ?)`,
			s.ID, s.Name, s.BackgroundImage, s.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to import schema %s: %w", s.Name, err)
		}
	}

	// Import schema items
	for _, i := range backup.SchemaItems {
		_, err := db.Exec(`
			INSERT INTO schema_items (id, device_id, schema_id, x, y, width, height)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			i.ID, i.DeviceID, i.SchemaID, i.X, i.Y, i.Width, i.Height)
		if err != nil {
			return fmt.Errorf("failed to import schema item %d: %w", i.ID, err)
		}
	}

	// Import settings
	settingsRepo := database.NewSettingsRepository(db)
	for key, value := range backup.Settings {
		settingsRepo.Set(key, value)
	}

	return nil
}
