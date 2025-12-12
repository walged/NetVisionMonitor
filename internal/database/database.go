package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

var instance *Database

// Initialize creates or opens the SQLite database
func Initialize(dataDir string) (*Database, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "netvision.db")

	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	instance = &Database{db: db}

	// Run migrations
	if err := instance.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return instance, nil
}

// Get returns the database instance
func Get() *Database {
	return instance
}

// DB returns the underlying sql.DB
func (d *Database) DB() *sql.DB {
	return d.db
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Migrate runs all database migrations
func (d *Database) Migrate() error {
	migrations := []string{
		migrationCredentials,
		migrationDevices,
		migrationSwitches,
		migrationSwitchPorts,
		migrationCameras,
		migrationServers,
		migrationEvents,
		migrationSchemas,
		migrationSchemaItems,
		migrationSettings,
		migrationStatusHistory,
	}

	for _, migration := range migrations {
		if _, err := d.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// Run optional migrations that may fail (e.g., adding columns that already exist)
	optionalMigrations := []string{
		migrationAddManufacturer,
		migrationSwitchesV3,
		migrationSwitchPortsV2,
		migrationSwitchesSFP,
		migrationSwitchesUplink,
		migrationServersUplink,
		migrationSwitchesWriteCommunity,
	}
	for _, migration := range optionalMigrations {
		d.db.Exec(migration) // Ignore errors for optional migrations
	}

	return nil
}

// Migration SQL statements
const migrationCredentials = `
CREATE TABLE IF NOT EXISTS credentials (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	type TEXT NOT NULL CHECK(type IN ('snmp', 'rtsp', 'onvif', 'ssh')),
	username TEXT NOT NULL DEFAULT '',
	password TEXT NOT NULL DEFAULT '',
	note TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationDevices = `
CREATE TABLE IF NOT EXISTS devices (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	ip_address TEXT NOT NULL,
	type TEXT NOT NULL CHECK(type IN ('switch', 'server', 'camera')),
	model TEXT DEFAULT '',
	credential_id INTEGER REFERENCES credentials(id) ON DELETE SET NULL,
	status TEXT DEFAULT 'unknown' CHECK(status IN ('online', 'offline', 'unknown')),
	last_check DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(type);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);
`

const migrationSwitches = `
CREATE TABLE IF NOT EXISTS switches (
	device_id INTEGER PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
	snmp_community TEXT DEFAULT '',
	snmp_write_community TEXT DEFAULT '',
	snmp_version TEXT DEFAULT 'v2c' CHECK(snmp_version IN ('v1', 'v2c', 'v3')),
	port_count INTEGER DEFAULT 24,
	snmpv3_user TEXT DEFAULT '',
	snmpv3_security TEXT DEFAULT 'noAuthNoPriv' CHECK(snmpv3_security IN ('noAuthNoPriv', 'authNoPriv', 'authPriv')),
	snmpv3_auth_proto TEXT DEFAULT '',
	snmpv3_auth_pass TEXT DEFAULT '',
	snmpv3_priv_proto TEXT DEFAULT '',
	snmpv3_priv_pass TEXT DEFAULT ''
);
`

const migrationSwitchesV3 = `
ALTER TABLE switches ADD COLUMN snmpv3_user TEXT DEFAULT '';
ALTER TABLE switches ADD COLUMN snmpv3_security TEXT DEFAULT 'noAuthNoPriv';
ALTER TABLE switches ADD COLUMN snmpv3_auth_proto TEXT DEFAULT '';
ALTER TABLE switches ADD COLUMN snmpv3_auth_pass TEXT DEFAULT '';
ALTER TABLE switches ADD COLUMN snmpv3_priv_proto TEXT DEFAULT '';
ALTER TABLE switches ADD COLUMN snmpv3_priv_pass TEXT DEFAULT '';
`

const migrationSwitchPorts = `
CREATE TABLE IF NOT EXISTS switch_ports (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	switch_id INTEGER NOT NULL REFERENCES switches(device_id) ON DELETE CASCADE,
	port_number INTEGER NOT NULL,
	name TEXT DEFAULT '',
	status TEXT DEFAULT 'unknown' CHECK(status IN ('up', 'down', 'unknown')),
	speed TEXT DEFAULT '',
	linked_camera_id INTEGER REFERENCES devices(id) ON DELETE SET NULL,
	UNIQUE(switch_id, port_number)
);

CREATE INDEX IF NOT EXISTS idx_switch_ports_switch ON switch_ports(switch_id);
`

const migrationCameras = `
CREATE TABLE IF NOT EXISTS cameras (
	device_id INTEGER PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
	rtsp_url TEXT DEFAULT '',
	onvif_port INTEGER DEFAULT 80,
	snapshot_url TEXT DEFAULT '',
	stream_type TEXT DEFAULT 'jpeg' CHECK(stream_type IN ('jpeg', 'hls', 'mjpeg'))
);
`

const migrationServers = `
CREATE TABLE IF NOT EXISTS servers (
	device_id INTEGER PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
	tcp_ports TEXT DEFAULT '[]',
	use_snmp INTEGER DEFAULT 0
);
`

const migrationEvents = `
CREATE TABLE IF NOT EXISTS events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	device_id INTEGER REFERENCES devices(id) ON DELETE CASCADE,
	type TEXT NOT NULL,
	level TEXT NOT NULL CHECK(level IN ('info', 'warn', 'error')),
	message TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_events_device ON events(device_id);
CREATE INDEX IF NOT EXISTS idx_events_level ON events(level);
CREATE INDEX IF NOT EXISTS idx_events_created ON events(created_at DESC);
`

const migrationSchemas = `
CREATE TABLE IF NOT EXISTS schemas (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	background_image TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationSchemaItems = `
CREATE TABLE IF NOT EXISTS schema_items (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
	schema_id INTEGER NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,
	x REAL DEFAULT 0,
	y REAL DEFAULT 0,
	width REAL DEFAULT 50,
	height REAL DEFAULT 50,
	UNIQUE(device_id, schema_id)
);

CREATE INDEX IF NOT EXISTS idx_schema_items_schema ON schema_items(schema_id);
`

const migrationSettings = `
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
`

const migrationStatusHistory = `
CREATE TABLE IF NOT EXISTS status_history (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
	status TEXT NOT NULL CHECK(status IN ('online', 'offline', 'unknown')),
	latency INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_status_history_device ON status_history(device_id);
CREATE INDEX IF NOT EXISTS idx_status_history_created ON status_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_status_history_device_created ON status_history(device_id, created_at DESC);
`

const migrationAddManufacturer = `
ALTER TABLE devices ADD COLUMN manufacturer TEXT DEFAULT '';
`

const migrationSwitchPortsV2 = `
ALTER TABLE switch_ports ADD COLUMN port_type TEXT DEFAULT 'copper' CHECK(port_type IN ('copper', 'sfp'));
ALTER TABLE switch_ports ADD COLUMN linked_switch_id INTEGER REFERENCES devices(id) ON DELETE SET NULL;
`

const migrationSwitchesSFP = `
ALTER TABLE switches ADD COLUMN sfp_port_count INTEGER DEFAULT 0;
`

const migrationSwitchesUplink = `
ALTER TABLE switches ADD COLUMN uplink_switch_id INTEGER REFERENCES devices(id) ON DELETE SET NULL;
ALTER TABLE switches ADD COLUMN uplink_port_id INTEGER REFERENCES switch_ports(id) ON DELETE SET NULL;
`

const migrationServersUplink = `
ALTER TABLE servers ADD COLUMN uplink_switch_id INTEGER REFERENCES devices(id) ON DELETE SET NULL;
ALTER TABLE servers ADD COLUMN uplink_port_id INTEGER REFERENCES switch_ports(id) ON DELETE SET NULL;
`

const migrationSwitchesWriteCommunity = `
ALTER TABLE switches ADD COLUMN snmp_write_community TEXT DEFAULT '';
`

// FixExistingPortTypes updates port_type for existing ports based on switch sfp_port_count
func (d *Database) FixExistingPortTypes() error {
	// First, fix sfp_port_count for known models where it's not set
	d.fixSfpPortCountFromModel()

	// Get all switches with their sfp_port_count
	rows, err := d.db.Query(`
		SELECT s.device_id, s.port_count, COALESCE(s.sfp_port_count, 0) as sfp_port_count
		FROM switches s
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type switchInfo struct {
		deviceID     int64
		portCount    int
		sfpPortCount int
	}

	var switches []switchInfo
	for rows.Next() {
		var s switchInfo
		if err := rows.Scan(&s.deviceID, &s.portCount, &s.sfpPortCount); err != nil {
			continue
		}
		switches = append(switches, s)
	}

	// Update port types for each switch
	for _, sw := range switches {
		if sw.sfpPortCount <= 0 {
			continue
		}

		copperPorts := sw.portCount - sw.sfpPortCount

		// Set copper ports
		_, err := d.db.Exec(`
			UPDATE switch_ports SET port_type = 'copper'
			WHERE switch_id = ? AND port_number <= ?`,
			sw.deviceID, copperPorts,
		)
		if err != nil {
			continue
		}

		// Set SFP ports
		_, err = d.db.Exec(`
			UPDATE switch_ports SET port_type = 'sfp'
			WHERE switch_id = ? AND port_number > ?`,
			sw.deviceID, copperPorts,
		)
		if err != nil {
			continue
		}
	}

	return nil
}

// fixSfpPortCountFromModel updates sfp_port_count based on known switch models
func (d *Database) fixSfpPortCountFromModel() {
	// Map of model patterns to SFP port count
	modelSfpPorts := map[string]int{
		// TFortis models
		"PSW-1G4F":          1, // 4 PoE + 1 SFP = 5 ports
		"PSW-1G4F-Box":      1,
		"PSW-1G4F-Ex":       1,
		"PSW-1G4F-UPS":      1,
		"PSW-2G4F":          2, // 4 PoE + 2 SFP = 6 ports
		"PSW-2G4F-Box":      2,
		"PSW-2G4F-Ex":       2,
		"PSW-2G4F-UPS":      2,
		"PSW-2G+":           2, // 2 PoE + 2 SFP = 4 ports
		"PSW-2G+-Box":       2,
		"PSW-2G+-Ex":        2,
		"PSW-2G+-UPS-Box":   2,
		"PSW-2G2F+-UPS":     2,
		"PSW-2G6F+":         2, // 6 PoE + 2 SFP = 8 ports
		"PSW-2G6F+-Box":     2,
		"PSW-2G6F+-UPS-Box": 2,
		"PSW-2G8F+":         2, // 8 PoE + 2 SFP = 10 ports
		"PSW-2G8F+-Box":     2,
		"PSW-2G8F+-UPS-Box": 2,
	}

	// Query all devices with switches that have 0 sfp_port_count
	rows, err := d.db.Query(`
		SELECT d.id, d.model
		FROM devices d
		JOIN switches s ON s.device_id = d.id
		WHERE d.type = 'switch' AND (s.sfp_port_count IS NULL OR s.sfp_port_count = 0)
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var deviceID int64
		var model string
		if err := rows.Scan(&deviceID, &model); err != nil {
			continue
		}

		// Check if model matches any known pattern
		sfpCount, ok := modelSfpPorts[model]
		if !ok {
			continue
		}

		// Update sfp_port_count
		_, _ = d.db.Exec(`
			UPDATE switches SET sfp_port_count = ? WHERE device_id = ?
		`, sfpCount, deviceID)
	}
}
