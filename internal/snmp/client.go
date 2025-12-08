package snmp

import (
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
)

// SNMPVersion represents SNMP protocol version
type SNMPVersion string

const (
	SNMPv1  SNMPVersion = "v1"
	SNMPv2c SNMPVersion = "v2c"
	SNMPv3  SNMPVersion = "v3"
)

// SNMPv3Security represents SNMPv3 security level
type SNMPv3Security string

const (
	NoAuthNoPriv SNMPv3Security = "noAuthNoPriv"
	AuthNoPriv   SNMPv3Security = "authNoPriv"
	AuthPriv     SNMPv3Security = "authPriv"
)

// Client wraps gosnmp for easier use
type Client struct {
	target    string
	community string
	port      uint16
	timeout   time.Duration
	retries   int
	version   SNMPVersion

	// SNMPv3 settings
	v3User         string
	v3Security     SNMPv3Security
	v3AuthProtocol string // MD5, SHA
	v3AuthPassword string
	v3PrivProtocol string // DES, AES
	v3PrivPassword string
}

// NewClient creates a new SNMP client (defaults to v2c)
func NewClient(target, community string) *Client {
	return &Client{
		target:    target,
		community: community,
		port:      161,
		timeout:   2 * time.Second,
		retries:   1,
		version:   SNMPv2c,
	}
}

// NewClientV3 creates a new SNMPv3 client
func NewClientV3(target, user string, security SNMPv3Security) *Client {
	return &Client{
		target:     target,
		port:       161,
		timeout:    2 * time.Second,
		retries:    1,
		version:    SNMPv3,
		v3User:     user,
		v3Security: security,
	}
}

// SetPort sets custom SNMP port
func (c *Client) SetPort(port uint16) {
	c.port = port
}

// SetTimeout sets SNMP timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// SetVersion sets SNMP version
func (c *Client) SetVersion(version SNMPVersion) {
	c.version = version
}

// SetV3Auth sets SNMPv3 authentication parameters
func (c *Client) SetV3Auth(protocol, password string) {
	c.v3AuthProtocol = protocol
	c.v3AuthPassword = password
}

// SetV3Priv sets SNMPv3 privacy parameters
func (c *Client) SetV3Priv(protocol, password string) {
	c.v3PrivProtocol = protocol
	c.v3PrivPassword = password
}

// connect creates and connects SNMP client
func (c *Client) connect() (*gosnmp.GoSNMP, error) {
	snmp := &gosnmp.GoSNMP{
		Target:  c.target,
		Port:    c.port,
		Timeout: c.timeout,
		Retries: c.retries,
	}

	switch c.version {
	case SNMPv1:
		snmp.Version = gosnmp.Version1
		snmp.Community = c.community
	case SNMPv2c:
		snmp.Version = gosnmp.Version2c
		snmp.Community = c.community
	case SNMPv3:
		snmp.Version = gosnmp.Version3
		snmp.SecurityModel = gosnmp.UserSecurityModel

		// Set security parameters
		sp := &gosnmp.UsmSecurityParameters{
			UserName: c.v3User,
		}

		switch c.v3Security {
		case NoAuthNoPriv:
			snmp.MsgFlags = gosnmp.NoAuthNoPriv
		case AuthNoPriv:
			snmp.MsgFlags = gosnmp.AuthNoPriv
			sp.AuthenticationProtocol = c.getAuthProtocol()
			sp.AuthenticationPassphrase = c.v3AuthPassword
		case AuthPriv:
			snmp.MsgFlags = gosnmp.AuthPriv
			sp.AuthenticationProtocol = c.getAuthProtocol()
			sp.AuthenticationPassphrase = c.v3AuthPassword
			sp.PrivacyProtocol = c.getPrivProtocol()
			sp.PrivacyPassphrase = c.v3PrivPassword
		}

		snmp.SecurityParameters = sp
	default:
		snmp.Version = gosnmp.Version2c
		snmp.Community = c.community
	}

	err := snmp.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", c.target, err)
	}

	return snmp, nil
}

func (c *Client) getAuthProtocol() gosnmp.SnmpV3AuthProtocol {
	switch c.v3AuthProtocol {
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
	switch c.v3PrivProtocol {
	case "AES":
		return gosnmp.AES
	case "AES192":
		return gosnmp.AES192
	case "AES256":
		return gosnmp.AES256
	case "AES192C":
		return gosnmp.AES192C
	case "AES256C":
		return gosnmp.AES256C
	default:
		return gosnmp.DES
	}
}

// Get performs SNMP GET for single OID
func (c *Client) Get(oid string) (interface{}, error) {
	snmp, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer snmp.Conn.Close()

	result, err := snmp.Get([]string{oid})
	if err != nil {
		return nil, fmt.Errorf("SNMP GET failed for %s: %w", oid, err)
	}

	if len(result.Variables) == 0 {
		return nil, fmt.Errorf("no result for OID %s", oid)
	}

	return decodeValue(result.Variables[0]), nil
}

// GetMultiple performs SNMP GET for multiple OIDs
func (c *Client) GetMultiple(oids []string) (map[string]interface{}, error) {
	snmp, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer snmp.Conn.Close()

	result, err := snmp.Get(oids)
	if err != nil {
		return nil, fmt.Errorf("SNMP GET failed: %w", err)
	}

	values := make(map[string]interface{})
	for _, v := range result.Variables {
		values[v.Name] = decodeValue(v)
	}

	return values, nil
}

// Walk performs SNMP WALK on OID subtree
func (c *Client) Walk(rootOid string) (map[string]interface{}, error) {
	snmp, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer snmp.Conn.Close()

	values := make(map[string]interface{})

	err = snmp.Walk(rootOid, func(pdu gosnmp.SnmpPDU) error {
		values[pdu.Name] = decodeValue(pdu)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("SNMP WALK failed for %s: %w", rootOid, err)
	}

	return values, nil
}

// Set performs SNMP SET operation
func (c *Client) Set(oid string, value interface{}, valueType gosnmp.Asn1BER) error {
	snmp, err := c.connect()
	if err != nil {
		return err
	}
	defer snmp.Conn.Close()

	pdu := gosnmp.SnmpPDU{
		Name:  oid,
		Type:  valueType,
		Value: value,
	}

	_, err = snmp.Set([]gosnmp.SnmpPDU{pdu})
	if err != nil {
		return fmt.Errorf("SNMP SET failed for %s: %w", oid, err)
	}

	return nil
}

// SetInteger sets an integer value via SNMP
func (c *Client) SetInteger(oid string, value int) error {
	return c.Set(oid, value, gosnmp.Integer)
}

// decodeValue converts SNMP PDU to Go value
func decodeValue(pdu gosnmp.SnmpPDU) interface{} {
	switch pdu.Type {
	case gosnmp.OctetString:
		return string(pdu.Value.([]byte))
	case gosnmp.Integer:
		return pdu.Value.(int)
	case gosnmp.Counter32:
		return pdu.Value.(uint)
	case gosnmp.Counter64:
		return pdu.Value.(uint64)
	case gosnmp.Gauge32:
		return pdu.Value.(uint)
	case gosnmp.TimeTicks:
		return pdu.Value.(uint32)
	case gosnmp.IPAddress:
		return pdu.Value.(string)
	default:
		return pdu.Value
	}
}

// TestConnection tests if SNMP is accessible
func (c *Client) TestConnection() error {
	// Try to get sysDescr
	_, err := c.Get(".1.3.6.1.2.1.1.1.0")
	return err
}
