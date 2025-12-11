package models

import "time"

type DeviceType string

const (
	DeviceTypeSwitch DeviceType = "switch"
	DeviceTypeServer DeviceType = "server"
	DeviceTypeCamera DeviceType = "camera"
)

type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusUnknown DeviceStatus = "unknown"
)

type Device struct {
	ID           int64        `json:"id"`
	Name         string       `json:"name"`
	IPAddress    string       `json:"ip_address"`
	Type         DeviceType   `json:"type"`
	Manufacturer string       `json:"manufacturer"`
	Model        string       `json:"model"`
	CredentialID *int64       `json:"credential_id,omitempty"`
	Status       DeviceStatus `json:"status"`
	LastCheck    *time.Time   `json:"last_check,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type Switch struct {
	DeviceID      int64  `json:"device_id"`
	SNMPCommunity string `json:"snmp_community"`
	SNMPVersion   string `json:"snmp_version"` // v1, v2c, v3
	PortCount     int    `json:"port_count"`
	SFPPortCount  int    `json:"sfp_port_count"` // Number of SFP ports (last N ports)
	// SNMPv3 settings
	SNMPv3User       string `json:"snmpv3_user,omitempty"`
	SNMPv3Security   string `json:"snmpv3_security,omitempty"`   // noAuthNoPriv, authNoPriv, authPriv
	SNMPv3AuthProto  string `json:"snmpv3_auth_proto,omitempty"` // MD5, SHA
	SNMPv3AuthPass   string `json:"snmpv3_auth_pass,omitempty"`
	SNMPv3PrivProto  string `json:"snmpv3_priv_proto,omitempty"` // DES, AES
	SNMPv3PrivPass   string `json:"snmpv3_priv_pass,omitempty"`
	// Uplink settings
	UplinkSwitchID *int64 `json:"uplink_switch_id,omitempty"` // Parent switch ID
	UplinkPortID   *int64 `json:"uplink_port_id,omitempty"`   // SFP port ID on parent switch
}

type SwitchPort struct {
	ID             int64  `json:"id"`
	SwitchID       int64  `json:"switch_id"`
	PortNumber     int    `json:"port_number"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	Speed          string `json:"speed"`
	PortType       string `json:"port_type"`                   // "copper" or "sfp"
	LinkedCameraID *int64 `json:"linked_camera_id,omitempty"`  // Only for copper ports
	LinkedSwitchID *int64 `json:"linked_switch_id,omitempty"`  // Only for SFP ports (uplink)
}

type Camera struct {
	DeviceID    int64  `json:"device_id"`
	RTSPURL     string `json:"rtsp_url"`
	ONVIFPort   int    `json:"onvif_port"`
	SnapshotURL string `json:"snapshot_url"`
	StreamType  string `json:"stream_type"`
}

type Server struct {
	DeviceID int64  `json:"device_id"`
	TCPPorts string `json:"tcp_ports"` // JSON array
	UseSNMP  bool   `json:"use_snmp"`
	// Uplink settings
	UplinkSwitchID *int64 `json:"uplink_switch_id,omitempty"` // Parent switch ID
	UplinkPortID   *int64 `json:"uplink_port_id,omitempty"`   // SFP port ID on parent switch
}

// DeviceWithDetails combines device with its type-specific details
type DeviceWithDetails struct {
	Device
	Switch *Switch      `json:"switch,omitempty"`
	Camera *Camera      `json:"camera,omitempty"`
	Server *Server      `json:"server,omitempty"`
	Ports  []SwitchPort `json:"ports,omitempty"`
}

// SNMPPortInfo contains SNMP-based information about a switch port
type SNMPPortInfo struct {
	PortNumber  int    `json:"port_number"`
	Status      string `json:"status"` // "up", "down", "unknown"
	Speed       int64  `json:"speed"`
	SpeedStr    string `json:"speed_str"`
	RxBytes     uint64 `json:"rx_bytes"`
	TxBytes     uint64 `json:"tx_bytes"`
	Description string `json:"description"`
}

// SNMPPoEInfo contains PoE information for a port
type SNMPPoEInfo struct {
	PortNumber int     `json:"port_number"`
	Enabled    bool    `json:"enabled"`
	Active     bool    `json:"active"`
	Status     string  `json:"status"` // "on", "off", "error"
	PowerMW    int     `json:"power_mw"`
	PowerW     float64 `json:"power_w"`
}

// SNMPSystemInfo contains switch system information
type SNMPSystemInfo struct {
	FirmwareVersion string       `json:"firmware_version"`
	UPS             *SNMPUPSInfo `json:"ups,omitempty"`
}

// SNMPUPSInfo contains UPS status
type SNMPUPSInfo struct {
	Present bool   `json:"present"`
	Status  string `json:"status"`
	Charge  int    `json:"charge"`
}

// SNMPAutoRestartInfo contains AutoRestart settings for a port
type SNMPAutoRestartInfo struct {
	PortNumber int    `json:"port_number"`
	Mode       int    `json:"mode"`
	ModeStr    string `json:"mode_str"`
	PingIP     string `json:"ping_ip"`
	LinkSpeed  int    `json:"link_speed"`
	Status     string `json:"status"`
}

// SwitchSNMPData contains all SNMP data for a switch
type SwitchSNMPData struct {
	DeviceID    int64                 `json:"device_id"`
	SystemInfo  *SNMPSystemInfo       `json:"system_info"`
	Ports       []SNMPPortInfo        `json:"ports"`
	PoE         []SNMPPoEInfo         `json:"poe"`
	AutoRestart []SNMPAutoRestartInfo `json:"auto_restart"`
	Error       string                `json:"error,omitempty"`
}
