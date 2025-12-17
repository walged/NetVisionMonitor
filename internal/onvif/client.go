package onvif

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Client is an ONVIF client for camera communication
type Client struct {
	Address  string
	Username string
	Password string
	Timeout  time.Duration
}

// DeviceInfo contains camera device information
type DeviceInfo struct {
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmware_version"`
	SerialNumber    string `json:"serial_number"`
	HardwareID      string `json:"hardware_id"`
}

// MediaProfile represents an ONVIF media profile
type MediaProfile struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

// StreamURI contains stream URL information
type StreamURI struct {
	URI       string `json:"uri"`
	ProfileToken string `json:"profile_token"`
}

// CameraInfo contains all discovered camera information
type CameraInfo struct {
	DeviceInfo   *DeviceInfo    `json:"device_info"`
	Profiles     []MediaProfile `json:"profiles"`
	StreamURI    string         `json:"stream_uri"`
	SnapshotURI  string         `json:"snapshot_uri"`
	ONVIFVersion string         `json:"onvif_version"`
}

// NewClient creates a new ONVIF client
func NewClient(host string, port int, username, password string) *Client {
	if port <= 0 {
		port = 80
	}
	return &Client{
		Address:  fmt.Sprintf("http://%s:%d", host, port),
		Username: username,
		Password: password,
		Timeout:  10 * time.Second,
	}
}

// createPasswordDigest creates WS-Security password digest
func createPasswordDigest(password, nonce string, created time.Time) string {
	nonceBytes, _ := base64.StdEncoding.DecodeString(nonce)
	createdStr := created.UTC().Format("2006-01-02T15:04:05Z")

	// SHA1(nonce + created + password)
	h := sha1.New()
	h.Write(nonceBytes)
	h.Write([]byte(createdStr))
	h.Write([]byte(password))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// createSecurityHeader creates WS-Security SOAP header
func (c *Client) createSecurityHeader() string {
	if c.Username == "" {
		return ""
	}

	nonce := make([]byte, 16)
	for i := range nonce {
		nonce[i] = byte(i * 17)
	}
	nonceB64 := base64.StdEncoding.EncodeToString(nonce)
	created := time.Now().UTC()
	createdStr := created.Format("2006-01-02T15:04:05Z")
	digest := createPasswordDigest(c.Password, nonceB64, created)

	return fmt.Sprintf(`
	<Security xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd" mustUnderstand="1">
		<UsernameToken>
			<Username>%s</Username>
			<Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">%s</Password>
			<Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">%s</Nonce>
			<Created xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">%s</Created>
		</UsernameToken>
	</Security>`, c.Username, digest, nonceB64, createdStr)
}

// createSOAPEnvelope creates a SOAP envelope with optional security header
func (c *Client) createSOAPEnvelope(body string) string {
	security := c.createSecurityHeader()
	headerContent := ""
	if security != "" {
		headerContent = security
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
	xmlns:tds="http://www.onvif.org/ver10/device/wsdl"
	xmlns:trt="http://www.onvif.org/ver10/media/wsdl"
	xmlns:tt="http://www.onvif.org/ver10/schema">
	<s:Header>%s</s:Header>
	<s:Body>%s</s:Body>
</s:Envelope>`, headerContent, body)
}

// doRequest performs a SOAP request
func (c *Client) doRequest(ctx context.Context, endpoint, body string) ([]byte, error) {
	envelope := c.createSOAPEnvelope(body)

	req, err := http.NewRequestWithContext(ctx, "POST", c.Address+endpoint, bytes.NewBufferString(envelope))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ONVIF error: status %d, body: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// GetDeviceInformation retrieves device information
func (c *Client) GetDeviceInformation(ctx context.Context) (*DeviceInfo, error) {
	body := `<tds:GetDeviceInformation/>`

	data, err := c.doRequest(ctx, "/onvif/device_service", body)
	if err != nil {
		return nil, err
	}

	// Parse response
	type GetDeviceInformationResponse struct {
		Manufacturer    string `xml:"Body>GetDeviceInformationResponse>Manufacturer"`
		Model           string `xml:"Body>GetDeviceInformationResponse>Model"`
		FirmwareVersion string `xml:"Body>GetDeviceInformationResponse>FirmwareVersion"`
		SerialNumber    string `xml:"Body>GetDeviceInformationResponse>SerialNumber"`
		HardwareId      string `xml:"Body>GetDeviceInformationResponse>HardwareId"`
	}

	var resp GetDeviceInformationResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device info: %w", err)
	}

	return &DeviceInfo{
		Manufacturer:    resp.Manufacturer,
		Model:           resp.Model,
		FirmwareVersion: resp.FirmwareVersion,
		SerialNumber:    resp.SerialNumber,
		HardwareID:      resp.HardwareId,
	}, nil
}

// mediaServiceEndpoints contains common ONVIF media service endpoints
var mediaServiceEndpoints = []string{
	"/onvif/media_service",
	"/onvif/Media",
	"/onvif/media",
	"/onvif/services/media",
	"/Media",
	"/media",
}

// GetProfiles retrieves media profiles
func (c *Client) GetProfiles(ctx context.Context) ([]MediaProfile, error) {
	body := `<trt:GetProfiles/>`

	var data []byte
	var err error

	// Try all known media service endpoints
	for _, endpoint := range mediaServiceEndpoints {
		data, err = c.doRequest(ctx, endpoint, body)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	// Parse response
	type Profile struct {
		Token string `xml:"token,attr"`
		Name  string `xml:"Name"`
	}
	type GetProfilesResponse struct {
		Profiles []Profile `xml:"Body>GetProfilesResponse>Profiles"`
	}

	var resp GetProfilesResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse profiles: %w", err)
	}

	profiles := make([]MediaProfile, len(resp.Profiles))
	for i, p := range resp.Profiles {
		profiles[i] = MediaProfile{
			Token: p.Token,
			Name:  p.Name,
		}
	}

	return profiles, nil
}

// GetStreamURI retrieves RTSP stream URI for a profile
func (c *Client) GetStreamURI(ctx context.Context, profileToken string) (string, error) {
	body := fmt.Sprintf(`
	<trt:GetStreamUri>
		<trt:StreamSetup>
			<tt:Stream>RTP-Unicast</tt:Stream>
			<tt:Transport>
				<tt:Protocol>RTSP</tt:Protocol>
			</tt:Transport>
		</trt:StreamSetup>
		<trt:ProfileToken>%s</trt:ProfileToken>
	</trt:GetStreamUri>`, profileToken)

	var data []byte
	var err error

	// Try all known media service endpoints
	for _, endpoint := range mediaServiceEndpoints {
		data, err = c.doRequest(ctx, endpoint, body)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", err
	}

	// Parse response
	type GetStreamUriResponse struct {
		Uri string `xml:"Body>GetStreamUriResponse>MediaUri>Uri"`
	}

	var resp GetStreamUriResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse stream URI: %w", err)
	}

	return resp.Uri, nil
}

// GetSnapshotURI retrieves snapshot URI for a profile
func (c *Client) GetSnapshotURI(ctx context.Context, profileToken string) (string, error) {
	body := fmt.Sprintf(`<trt:GetSnapshotUri>
		<trt:ProfileToken>%s</trt:ProfileToken>
	</trt:GetSnapshotUri>`, profileToken)

	var data []byte
	var err error

	// Try all known media service endpoints
	for _, endpoint := range mediaServiceEndpoints {
		data, err = c.doRequest(ctx, endpoint, body)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", err
	}

	// Parse response
	type GetSnapshotUriResponse struct {
		Uri string `xml:"Body>GetSnapshotUriResponse>MediaUri>Uri"`
	}

	var resp GetSnapshotUriResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse snapshot URI: %w", err)
	}

	return resp.Uri, nil
}

// DiscoverCamera performs full camera discovery via ONVIF
func (c *Client) DiscoverCamera(ctx context.Context) (*CameraInfo, error) {
	info := &CameraInfo{}

	// Get device information
	devInfo, err := c.GetDeviceInformation(ctx)
	if err != nil {
		// Try without auth first if auth fails
		if c.Username != "" {
			noAuthClient := &Client{
				Address: c.Address,
				Timeout: c.Timeout,
			}
			devInfo, err = noAuthClient.GetDeviceInformation(ctx)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get device info: %w", err)
		}
	}
	info.DeviceInfo = devInfo

	// Get profiles
	profiles, err := c.GetProfiles(ctx)
	if err != nil {
		// Some cameras may not require auth for profiles
		return info, nil
	}
	info.Profiles = profiles

	if len(profiles) == 0 {
		return info, nil
	}

	// Get stream URI for first profile
	streamURI, err := c.GetStreamURI(ctx, profiles[0].Token)
	if err == nil {
		info.StreamURI = c.addCredentialsToURI(streamURI)
	}

	// Get snapshot URI for first profile
	snapshotURI, err := c.GetSnapshotURI(ctx, profiles[0].Token)
	if err == nil {
		info.SnapshotURI = snapshotURI
	}

	return info, nil
}

// addCredentialsToURI adds username and password to RTSP URI
func (c *Client) addCredentialsToURI(uri string) string {
	if c.Username == "" || uri == "" {
		return uri
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return uri
	}

	parsed.User = url.UserPassword(c.Username, c.Password)
	return parsed.String()
}

// TestConnection tests ONVIF connectivity
func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.GetDeviceInformation(ctx)
	return err
}

// GetCapabilities retrieves device capabilities
func (c *Client) GetCapabilities(ctx context.Context) (map[string]string, error) {
	body := `<tds:GetCapabilities>
		<tds:Category>All</tds:Category>
	</tds:GetCapabilities>`

	data, err := c.doRequest(ctx, "/onvif/device_service", body)
	if err != nil {
		return nil, err
	}

	capabilities := make(map[string]string)

	// Extract capability URLs
	dataStr := string(data)
	if strings.Contains(dataStr, "Media") {
		capabilities["media"] = "supported"
	}
	if strings.Contains(dataStr, "PTZ") {
		capabilities["ptz"] = "supported"
	}
	if strings.Contains(dataStr, "Events") {
		capabilities["events"] = "supported"
	}

	return capabilities, nil
}

// Probe sends WS-Discovery probe to find ONVIF devices on network
func Probe(ctx context.Context, timeout time.Duration) ([]string, error) {
	// This is a simplified probe - full WS-Discovery is more complex
	// For now, we support manual IP entry
	return nil, fmt.Errorf("network discovery not implemented - use manual IP entry")
}

// GenerateUUID generates a UUID for SOAP message IDs
func GenerateUUID() string {
	return uuid.New().String()
}
