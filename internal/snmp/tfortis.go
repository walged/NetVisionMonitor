package snmp

import (
	"fmt"
	"strconv"
	"strings"
)

// TFortis OID definitions
const (
	// Base OID for TFortis devices
	TFortisBaseOID = ".1.3.6.1.4.1.42019"

	// System info
	OIDFirmwareVersion = ".1.3.6.1.4.1.42019.3.2.2.3.1.0"
	OIDUPSStatus       = ".1.3.6.1.4.1.42019.3.2.2.1.2.0"
	OIDUPSCharge       = ".1.3.6.1.4.1.42019.3.2.2.1.3.0"

	// Standard MIB-2 OIDs for port info
	OIDifOperStatus = ".1.3.6.1.2.1.2.2.1.8"  // Port operational status
	OIDifSpeed      = ".1.3.6.1.2.1.2.2.1.5"  // Port speed
	OIDifInOctets   = ".1.3.6.1.2.1.2.2.1.10" // RX bytes
	OIDifOutOctets  = ".1.3.6.1.2.1.2.2.1.16" // TX bytes
	OIDifDescr      = ".1.3.6.1.2.1.2.2.1.2"  // Interface description

	// TFortis PoE OIDs
	OIDPoEStateBase  = ".1.3.6.1.4.1.42019.3.2.1.3.1.1.2" // PoE enable/disable (RW)
	OIDPoEStatusBase = ".1.3.6.1.4.1.42019.3.2.2.5.1.1.2" // PoE actual status (RO)
	OIDPoEPowerBase  = ".1.3.6.1.4.1.42019.3.2.2.5.1.1.3" // PoE power in mW (RO)

	// TFortis AutoRestart OIDs
	OIDAutoRestartModeBase   = ".1.3.6.1.4.1.42019.3.2.1.2.1.1.2" // AutoRestart mode (RW)
	OIDAutoRestartIPBase     = ".1.3.6.1.4.1.42019.3.2.1.2.1.1.3" // IP to ping (RW)
	OIDAutoRestartSpeedBase  = ".1.3.6.1.4.1.42019.3.2.1.2.1.1.4" // Link speed setting (RW)
	OIDAutoRestartStatusBase = ".1.3.6.1.4.1.42019.3.2.2.2.1.1.2" // AutoRestart status (RO)

	// ComfortStart
	OIDComfortStartBase = ".1.3.6.1.4.1.42019.3.2.1.2.1.1.5" // ComfortStart delay (RW)
)

// Port status values
const (
	PortStatusUp      = 1
	PortStatusDown    = 2
	PortStatusTesting = 3
)

// PoE config values (for OIDPoEStateBase - RW)
// Согласно TFortis MIB: portPoeState { enabled(1), disabled(2) }
const (
	PoEEnabled  = 1 // PoE включено
	PoEDisabled = 2 // PoE выключено
)

// PoE status values (for OIDPoEStatusBase - RO)
// Согласно TFortis MIB: portPoeStatusState { up(1), down(2) }
const (
	PoEStatusUp   = 1 // Питание выдаётся
	PoEStatusDown = 2 // Питание не выдаётся
)

// AutoRestart mode values
const (
	AutoRestartDisabled  = 0
	AutoRestartAlways    = 1
	AutoRestartOnLink    = 2
	AutoRestartOnPing    = 3
	AutoRestartLinkSpeed = 4
)

// PortInfo contains information about a switch port
type PortInfo struct {
	PortNumber  int    `json:"port_number"`
	Status      string `json:"status"` // "up", "down", "testing"
	Speed       int64  `json:"speed"`  // bps
	SpeedStr    string `json:"speed_str"`
	RxBytes     uint64 `json:"rx_bytes"`
	TxBytes     uint64 `json:"tx_bytes"`
	Description string `json:"description"`
}

// PoEInfo contains PoE information for a port
type PoEInfo struct {
	PortNumber int     `json:"port_number"`
	Enabled    bool    `json:"enabled"`     // Is PoE enabled (config)
	Active     bool    `json:"active"`      // Is PoE actually active
	Status     string  `json:"status"`      // "on", "off", "error"
	PowerMW    int     `json:"power_mw"`    // Power consumption in milliwatts
	PowerW     float64 `json:"power_w"`     // Power consumption in watts
}

// AutoRestartInfo contains AutoRestart settings for a port
type AutoRestartInfo struct {
	PortNumber int    `json:"port_number"`
	Mode       int    `json:"mode"`        // 0-4
	ModeStr    string `json:"mode_str"`    // Human readable mode
	PingIP     string `json:"ping_ip"`     // IP for ping mode
	LinkSpeed  int    `json:"link_speed"`  // Expected link speed
	Status     string `json:"status"`      // Current status
}

// UPSInfo contains UPS status
type UPSInfo struct {
	Present bool   `json:"present"`
	Status  string `json:"status"` // "charging", "discharging", "full", "none"
	Charge  int    `json:"charge"` // Battery percentage
}

// SystemInfo contains switch system information
type SystemInfo struct {
	FirmwareVersion string   `json:"firmware_version"`
	UPS             *UPSInfo `json:"ups,omitempty"`
}

// TFortisClient provides TFortis-specific SNMP operations
type TFortisClient struct {
	client *Client
}

// NewTFortisClient creates a new TFortis SNMP client (v2c)
func NewTFortisClient(ipAddress, community string) *TFortisClient {
	return &TFortisClient{
		client: NewClient(ipAddress, community),
	}
}

// NewTFortisClientV3 creates a new TFortis SNMP client with v3 support
func NewTFortisClientV3(ipAddress, user, security, authProto, authPass, privProto, privPass string) *TFortisClient {
	client := NewClientV3(ipAddress, user, SNMPv3Security(security))
	if authProto != "" {
		client.SetV3Auth(authProto, authPass)
	}
	if privProto != "" {
		client.SetV3Priv(privProto, privPass)
	}
	return &TFortisClient{
		client: client,
	}
}

// NewTFortisClientAuto creates TFortis client based on version
func NewTFortisClientAuto(ipAddress, version, community, v3User, v3Security, v3AuthProto, v3AuthPass, v3PrivProto, v3PrivPass string) *TFortisClient {
	if version == "v3" {
		return NewTFortisClientV3(ipAddress, v3User, v3Security, v3AuthProto, v3AuthPass, v3PrivProto, v3PrivPass)
	}
	client := NewClient(ipAddress, community)
	if version == "v1" {
		client.SetVersion(SNMPv1)
	}
	return &TFortisClient{client: client}
}

// GetSystemInfo retrieves system information
func (t *TFortisClient) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// Get firmware version
	fw, err := t.client.Get(OIDFirmwareVersion)
	if err == nil {
		info.FirmwareVersion = fmt.Sprintf("%v", fw)
	}

	// Try to get UPS status (may not be available on all models)
	upsStatus, err := t.client.Get(OIDUPSStatus)
	if err == nil {
		info.UPS = &UPSInfo{
			Present: true,
			Status:  decodeUPSStatus(upsStatus),
		}

		// Get charge level
		charge, err := t.client.Get(OIDUPSCharge)
		if err == nil {
			if c, ok := charge.(int); ok {
				info.UPS.Charge = c
			}
		}
	}

	return info, nil
}

// GetPortInfo retrieves information about a specific port
func (t *TFortisClient) GetPortInfo(portNum int) (*PortInfo, error) {
	info := &PortInfo{PortNumber: portNum}

	// Get operational status
	status, err := t.client.Get(fmt.Sprintf("%s.%d", OIDifOperStatus, portNum))
	if err == nil {
		info.Status = decodePortStatus(status)
	}

	// Get speed
	speed, err := t.client.Get(fmt.Sprintf("%s.%d", OIDifSpeed, portNum))
	if err == nil {
		if s, ok := speed.(uint); ok {
			info.Speed = int64(s)
			info.SpeedStr = formatSpeed(int64(s))
		}
	}

	// Get RX bytes
	rx, err := t.client.Get(fmt.Sprintf("%s.%d", OIDifInOctets, portNum))
	if err == nil {
		switch v := rx.(type) {
		case uint:
			info.RxBytes = uint64(v)
		case uint64:
			info.RxBytes = v
		}
	}

	// Get TX bytes
	tx, err := t.client.Get(fmt.Sprintf("%s.%d", OIDifOutOctets, portNum))
	if err == nil {
		switch v := tx.(type) {
		case uint:
			info.TxBytes = uint64(v)
		case uint64:
			info.TxBytes = v
		}
	}

	// Get description
	descr, err := t.client.Get(fmt.Sprintf("%s.%d", OIDifDescr, portNum))
	if err == nil {
		info.Description = fmt.Sprintf("%v", descr)
	}

	return info, nil
}

// GetAllPortsInfo retrieves information about all ports
func (t *TFortisClient) GetAllPortsInfo(portCount int) ([]PortInfo, error) {
	ports := make([]PortInfo, 0, portCount)

	for i := 1; i <= portCount; i++ {
		info, err := t.GetPortInfo(i)
		if err != nil {
			// Skip ports that fail
			continue
		}
		ports = append(ports, *info)
	}

	return ports, nil
}

// GetPoEInfo retrieves PoE information for a specific port
func (t *TFortisClient) GetPoEInfo(portNum int) (*PoEInfo, error) {
	info := &PoEInfo{PortNumber: portNum}

	// Get PoE config: enabled(1), disabled(2)
	state, err := t.client.Get(fmt.Sprintf("%s.%d", OIDPoEStateBase, portNum))
	if err == nil {
		if s, ok := state.(int); ok {
			info.Enabled = s == PoEEnabled // 1 = PoE включено
			fmt.Printf("GetPoEInfo port %d: config=%d, enabled=%v\n", portNum, s, info.Enabled)
		}
	}

	// Get PoE actual status: up(1), down(2)
	status, err := t.client.Get(fmt.Sprintf("%s.%d", OIDPoEStatusBase, portNum))
	if err == nil {
		info.Status = decodePoEStatus(status)
		if s, ok := status.(int); ok {
			info.Active = s == PoEStatusUp // 1 = питание выдаётся
			fmt.Printf("GetPoEInfo port %d: status=%d, active=%v\n", portNum, s, info.Active)
		}
	}

	// Get power consumption
	power, err := t.client.Get(fmt.Sprintf("%s.%d", OIDPoEPowerBase, portNum))
	if err == nil {
		if p, ok := power.(int); ok {
			info.PowerMW = p
			info.PowerW = float64(p) / 1000.0
		}
	}

	return info, nil
}

// GetAllPoEInfo retrieves PoE information for all ports
func (t *TFortisClient) GetAllPoEInfo(portCount int) ([]PoEInfo, error) {
	poeInfos := make([]PoEInfo, 0, portCount)

	for i := 1; i <= portCount; i++ {
		info, err := t.GetPoEInfo(i)
		if err != nil {
			continue
		}
		poeInfos = append(poeInfos, *info)
	}

	return poeInfos, nil
}

// SetPoEEnabled sets PoE state on a port
// TFortis MIB: enabled(1), disabled(2)
func (t *TFortisClient) SetPoEEnabled(portNum int, enabled bool) error {
	value := PoEDisabled // 2 = выключить
	if enabled {
		value = PoEEnabled // 1 = включить
	}

	oid := fmt.Sprintf("%s.%d", OIDPoEStateBase, portNum)
	fmt.Printf("TFortis SetPoEEnabled: port=%d, enabled=%v, value=%d (1=On, 2=Off), oid=%s\n", portNum, enabled, value, oid)
	err := t.client.SetInteger(oid, value)
	if err != nil {
		fmt.Printf("TFortis SetPoEEnabled error: %v\n", err)
	} else {
		fmt.Printf("TFortis SetPoEEnabled: SUCCESS\n")
	}
	return err
}

// RestartPoE restarts PoE on a port (disable then enable after delay)
func (t *TFortisClient) RestartPoE(portNum int) error {
	// Disable PoE
	err := t.SetPoEEnabled(portNum, false)
	if err != nil {
		return fmt.Errorf("failed to disable PoE: %w", err)
	}

	// Note: The actual delay should be handled by the caller
	// or we could add a goroutine here

	return nil
}

// EnablePoE enables PoE on a port
func (t *TFortisClient) EnablePoE(portNum int) error {
	return t.SetPoEEnabled(portNum, true)
}

// GetAutoRestartInfo retrieves AutoRestart settings for a port
func (t *TFortisClient) GetAutoRestartInfo(portNum int) (*AutoRestartInfo, error) {
	info := &AutoRestartInfo{PortNumber: portNum}

	// Get mode
	mode, err := t.client.Get(fmt.Sprintf("%s.%d", OIDAutoRestartModeBase, portNum))
	if err == nil {
		if m, ok := mode.(int); ok {
			info.Mode = m
			info.ModeStr = decodeAutoRestartMode(m)
		}
	}

	// Get ping IP
	ip, err := t.client.Get(fmt.Sprintf("%s.%d", OIDAutoRestartIPBase, portNum))
	if err == nil {
		info.PingIP = fmt.Sprintf("%v", ip)
	}

	// Get link speed setting
	speed, err := t.client.Get(fmt.Sprintf("%s.%d", OIDAutoRestartSpeedBase, portNum))
	if err == nil {
		if s, ok := speed.(int); ok {
			info.LinkSpeed = s
		}
	}

	// Get current status
	status, err := t.client.Get(fmt.Sprintf("%s.%d", OIDAutoRestartStatusBase, portNum))
	if err == nil {
		info.Status = fmt.Sprintf("%v", status)
	}

	return info, nil
}

// SetAutoRestartMode sets the AutoRestart mode for a port
func (t *TFortisClient) SetAutoRestartMode(portNum int, mode int) error {
	oid := fmt.Sprintf("%s.%d", OIDAutoRestartModeBase, portNum)
	return t.client.SetInteger(oid, mode)
}

// SetAutoRestartPingIP sets the IP address for ping-based AutoRestart
func (t *TFortisClient) SetAutoRestartPingIP(portNum int, ip string) error {
	oid := fmt.Sprintf("%s.%d", OIDAutoRestartIPBase, portNum)
	// Parse IP to bytes
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	ipBytes := make([]byte, 4)
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 || v > 255 {
			return fmt.Errorf("invalid IP address: %s", ip)
		}
		ipBytes[i] = byte(v)
	}

	return t.client.Set(oid, string(ipBytes), 4) // IPAddress type
}

// TestConnection tests if the switch is accessible via SNMP
func (t *TFortisClient) TestConnection() error {
	return t.client.TestConnection()
}

// Helper functions

func decodePortStatus(value interface{}) string {
	if v, ok := value.(int); ok {
		switch v {
		case PortStatusUp:
			return "up"
		case PortStatusDown:
			return "down"
		case PortStatusTesting:
			return "testing"
		}
	}
	return "unknown"
}

func decodePoEStatus(value interface{}) string {
	if v, ok := value.(int); ok {
		switch v {
		case PoEStatusUp:
			return "on"
		case PoEStatusDown:
			return "off"
		}
	}
	return "unknown"
}

func decodeUPSStatus(value interface{}) string {
	if v, ok := value.(int); ok {
		switch v {
		case 0:
			return "none"
		case 1:
			return "charging"
		case 2:
			return "discharging"
		case 3:
			return "full"
		}
	}
	return "unknown"
}

func decodeAutoRestartMode(mode int) string {
	switch mode {
	case AutoRestartDisabled:
		return "Выключено"
	case AutoRestartAlways:
		return "Всегда"
	case AutoRestartOnLink:
		return "По состоянию линка"
	case AutoRestartOnPing:
		return "По пингу"
	case AutoRestartLinkSpeed:
		return "По скорости линка"
	default:
		return "Неизвестно"
	}
}

func formatSpeed(bps int64) string {
	if bps >= 1000000000 {
		return fmt.Sprintf("%d Гбит/с", bps/1000000000)
	}
	if bps >= 1000000 {
		return fmt.Sprintf("%d Мбит/с", bps/1000000)
	}
	if bps >= 1000 {
		return fmt.Sprintf("%d Кбит/с", bps/1000)
	}
	return fmt.Sprintf("%d бит/с", bps)
}
