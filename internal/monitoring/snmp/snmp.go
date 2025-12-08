package snmp

import (
	"context"
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
)

// Common SNMP OIDs
const (
	OIDSysDescr    = ".1.3.6.1.2.1.1.1.0"     // System description
	OIDSysUpTime   = ".1.3.6.1.2.1.1.3.0"     // System uptime
	OIDSysName     = ".1.3.6.1.2.1.1.5.0"     // System name
	OIDSysLocation = ".1.3.6.1.2.1.1.6.0"     // System location
	OIDIfNumber    = ".1.3.6.1.2.1.2.1.0"     // Number of interfaces
	OIDIfDescr     = ".1.3.6.1.2.1.2.2.1.2"   // Interface description
	OIDIfOperStatus = ".1.3.6.1.2.1.2.2.1.8"  // Interface operational status
	OIDIfSpeed     = ".1.3.6.1.2.1.2.2.1.5"   // Interface speed
	OIDIfInOctets  = ".1.3.6.1.2.1.2.2.1.10"  // Interface incoming octets
	OIDIfOutOctets = ".1.3.6.1.2.1.2.2.1.16"  // Interface outgoing octets
)

// Interface status constants
const (
	IfStatusUp      = 1
	IfStatusDown    = 2
	IfStatusTesting = 3
)

// Client represents an SNMP client
type Client struct {
	Target    string
	Port      uint16
	Community string
	Version   gosnmp.SnmpVersion
	Timeout   time.Duration
	Retries   int

	// SNMPv3 settings
	V3User         string
	V3Security     string // noAuthNoPriv, authNoPriv, authPriv
	V3AuthProtocol string // MD5, SHA
	V3AuthPassword string
	V3PrivProtocol string // DES, AES
	V3PrivPassword string
}

// NewClient creates a new SNMP client
func NewClient(target, community, version string, timeout time.Duration) *Client {
	var snmpVersion gosnmp.SnmpVersion
	switch version {
	case "v1":
		snmpVersion = gosnmp.Version1
	case "v3":
		snmpVersion = gosnmp.Version3
	default:
		snmpVersion = gosnmp.Version2c
	}

	return &Client{
		Target:    target,
		Port:      161,
		Community: community,
		Version:   snmpVersion,
		Timeout:   timeout,
		Retries:   2,
	}
}

// NewClientV3 creates a new SNMPv3 client
func NewClientV3(target, user, security, authProto, authPass, privProto, privPass string, timeout time.Duration) *Client {
	return &Client{
		Target:         target,
		Port:           161,
		Version:        gosnmp.Version3,
		Timeout:        timeout,
		Retries:        2,
		V3User:         user,
		V3Security:     security,
		V3AuthProtocol: authProto,
		V3AuthPassword: authPass,
		V3PrivProtocol: privProto,
		V3PrivPassword: privPass,
	}
}

// connect creates a new SNMP connection
func (c *Client) connect() (*gosnmp.GoSNMP, error) {
	snmp := &gosnmp.GoSNMP{
		Target:  c.Target,
		Port:    c.Port,
		Timeout: c.Timeout,
		Retries: c.Retries,
	}

	if c.Version == gosnmp.Version3 {
		snmp.Version = gosnmp.Version3
		snmp.SecurityModel = gosnmp.UserSecurityModel

		sp := &gosnmp.UsmSecurityParameters{
			UserName: c.V3User,
		}

		switch c.V3Security {
		case "authPriv":
			snmp.MsgFlags = gosnmp.AuthPriv
			sp.AuthenticationProtocol = c.getAuthProtocol()
			sp.AuthenticationPassphrase = c.V3AuthPassword
			sp.PrivacyProtocol = c.getPrivProtocol()
			sp.PrivacyPassphrase = c.V3PrivPassword
		case "authNoPriv":
			snmp.MsgFlags = gosnmp.AuthNoPriv
			sp.AuthenticationProtocol = c.getAuthProtocol()
			sp.AuthenticationPassphrase = c.V3AuthPassword
		default: // noAuthNoPriv
			snmp.MsgFlags = gosnmp.NoAuthNoPriv
		}

		snmp.SecurityParameters = sp
	} else {
		snmp.Version = c.Version
		snmp.Community = c.Community
	}

	if err := snmp.Connect(); err != nil {
		return nil, fmt.Errorf("SNMP connect failed: %w", err)
	}

	return snmp, nil
}

func (c *Client) getAuthProtocol() gosnmp.SnmpV3AuthProtocol {
	switch c.V3AuthProtocol {
	case "SHA":
		return gosnmp.SHA
	case "SHA224":
		return gosnmp.SHA224
	case "SHA256":
		return gosnmp.SHA256
	case "SHA384":
		return gosnmp.SHA384
	case "SHA512":
		return gosnmp.SHA512
	default:
		return gosnmp.MD5
	}
}

func (c *Client) getPrivProtocol() gosnmp.SnmpV3PrivProtocol {
	switch c.V3PrivProtocol {
	case "AES":
		return gosnmp.AES
	case "AES192":
		return gosnmp.AES192
	case "AES256":
		return gosnmp.AES256
	default:
		return gosnmp.DES
	}
}

// SystemInfo contains basic system information
type SystemInfo struct {
	Description string
	Uptime      time.Duration
	Name        string
	Location    string
}

// GetSystemInfo retrieves basic system information
func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	snmp, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer snmp.Conn.Close()

	oids := []string{OIDSysDescr, OIDSysUpTime, OIDSysName, OIDSysLocation}

	result, err := snmp.Get(oids)
	if err != nil {
		return nil, fmt.Errorf("SNMP get failed: %w", err)
	}

	info := &SystemInfo{}

	for _, variable := range result.Variables {
		switch variable.Name {
		case OIDSysDescr:
			info.Description = string(variable.Value.([]byte))
		case OIDSysUpTime:
			// Uptime is in hundredths of a second
			ticks := gosnmp.ToBigInt(variable.Value).Uint64()
			info.Uptime = time.Duration(ticks) * time.Millisecond * 10
		case OIDSysName:
			info.Name = string(variable.Value.([]byte))
		case OIDSysLocation:
			info.Location = string(variable.Value.([]byte))
		}
	}

	return info, nil
}

// InterfaceInfo contains interface information
type InterfaceInfo struct {
	Index       int
	Description string
	Status      string // "up", "down", "testing"
	Speed       uint64 // bits per second
	InOctets    uint64
	OutOctets   uint64
}

// GetInterfaceCount returns the number of interfaces
func (c *Client) GetInterfaceCount(ctx context.Context) (int, error) {
	snmp, err := c.connect()
	if err != nil {
		return 0, err
	}
	defer snmp.Conn.Close()

	result, err := snmp.Get([]string{OIDIfNumber})
	if err != nil {
		return 0, fmt.Errorf("SNMP get failed: %w", err)
	}

	if len(result.Variables) == 0 {
		return 0, fmt.Errorf("no interface count returned")
	}

	count := gosnmp.ToBigInt(result.Variables[0].Value).Int64()
	return int(count), nil
}

// GetInterfaceStatus retrieves status for a specific interface
func (c *Client) GetInterfaceStatus(ctx context.Context, ifIndex int) (*InterfaceInfo, error) {
	snmp, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer snmp.Conn.Close()

	oids := []string{
		fmt.Sprintf("%s.%d", OIDIfDescr, ifIndex),
		fmt.Sprintf("%s.%d", OIDIfOperStatus, ifIndex),
		fmt.Sprintf("%s.%d", OIDIfSpeed, ifIndex),
	}

	result, err := snmp.Get(oids)
	if err != nil {
		return nil, fmt.Errorf("SNMP get failed: %w", err)
	}

	info := &InterfaceInfo{
		Index: ifIndex,
	}

	for _, variable := range result.Variables {
		switch {
		case variable.Name == fmt.Sprintf("%s.%d", OIDIfDescr, ifIndex):
			if bytes, ok := variable.Value.([]byte); ok {
				info.Description = string(bytes)
			}
		case variable.Name == fmt.Sprintf("%s.%d", OIDIfOperStatus, ifIndex):
			status := gosnmp.ToBigInt(variable.Value).Int64()
			switch status {
			case IfStatusUp:
				info.Status = "up"
			case IfStatusDown:
				info.Status = "down"
			default:
				info.Status = "unknown"
			}
		case variable.Name == fmt.Sprintf("%s.%d", OIDIfSpeed, ifIndex):
			info.Speed = gosnmp.ToBigInt(variable.Value).Uint64()
		}
	}

	return info, nil
}

// GetAllInterfaceStatuses retrieves status for all interfaces
func (c *Client) GetAllInterfaceStatuses(ctx context.Context, maxPorts int) ([]InterfaceInfo, error) {
	snmp, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer snmp.Conn.Close()

	var interfaces []InterfaceInfo

	// Walk interface status OID
	err = snmp.Walk(OIDIfOperStatus, func(pdu gosnmp.SnmpPDU) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Extract interface index from OID
		var ifIndex int
		_, err := fmt.Sscanf(pdu.Name, OIDIfOperStatus+".%d", &ifIndex)
		if err != nil {
			return nil
		}

		if maxPorts > 0 && ifIndex > maxPorts {
			return nil
		}

		status := gosnmp.ToBigInt(pdu.Value).Int64()
		statusStr := "unknown"
		switch status {
		case IfStatusUp:
			statusStr = "up"
		case IfStatusDown:
			statusStr = "down"
		}

		interfaces = append(interfaces, InterfaceInfo{
			Index:  ifIndex,
			Status: statusStr,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("SNMP walk failed: %w", err)
	}

	return interfaces, nil
}

// CheckAvailability performs a simple availability check
func (c *Client) CheckAvailability(ctx context.Context) (bool, time.Duration, error) {
	start := time.Now()

	snmp, err := c.connect()
	if err != nil {
		return false, time.Since(start), err
	}
	defer snmp.Conn.Close()

	// Try to get system uptime
	_, err = snmp.Get([]string{OIDSysUpTime})
	latency := time.Since(start)

	if err != nil {
		return false, latency, err
	}

	return true, latency, nil
}
