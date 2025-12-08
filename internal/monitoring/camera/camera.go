package camera

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client handles camera connectivity checks
type Client struct {
	Timeout time.Duration
}

// NewClient creates a new camera client
func NewClient(timeout time.Duration) *Client {
	return &Client{
		Timeout: timeout,
	}
}

// CameraStatus contains camera check results
type CameraStatus struct {
	Available     bool
	RTSPAvailable bool
	ONVIFAvailable bool
	SnapshotOK    bool
	Latency       time.Duration
	Error         string
}

// CheckRTSP checks if RTSP stream is accessible
func (c *Client) CheckRTSP(ctx context.Context, rtspURL string) (bool, time.Duration, error) {
	if rtspURL == "" {
		return false, 0, fmt.Errorf("RTSP URL is empty")
	}

	// Parse RTSP URL
	u, err := url.Parse(rtspURL)
	if err != nil {
		return false, 0, fmt.Errorf("invalid RTSP URL: %w", err)
	}

	// Default RTSP port
	host := u.Host
	if !strings.Contains(host, ":") {
		host = host + ":554"
	}

	start := time.Now()

	// Try TCP connection to RTSP port
	dialer := net.Dialer{
		Timeout: c.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", host)
	latency := time.Since(start)

	if err != nil {
		return false, latency, fmt.Errorf("RTSP connection failed: %w", err)
	}
	defer conn.Close()

	// Send RTSP OPTIONS request
	conn.SetDeadline(time.Now().Add(c.Timeout))

	request := fmt.Sprintf("OPTIONS %s RTSP/1.0\r\nCSeq: 1\r\n\r\n", rtspURL)
	_, err = conn.Write([]byte(request))
	if err != nil {
		return false, latency, fmt.Errorf("RTSP write failed: %w", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return false, latency, fmt.Errorf("RTSP read failed: %w", err)
	}

	response := string(buf[:n])
	if strings.Contains(response, "RTSP/1.0 200") {
		return true, latency, nil
	}

	// Any response means RTSP is available (even if auth required)
	if strings.Contains(response, "RTSP/1.0") {
		return true, latency, nil
	}

	return false, latency, fmt.Errorf("invalid RTSP response")
}

// CheckONVIF checks if ONVIF service is accessible
func (c *Client) CheckONVIF(ctx context.Context, host string, port int) (bool, time.Duration, error) {
	if port <= 0 {
		port = 80
	}

	address := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	dialer := net.Dialer{
		Timeout: c.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	latency := time.Since(start)

	if err != nil {
		return false, latency, fmt.Errorf("ONVIF connection failed: %w", err)
	}
	defer conn.Close()

	// Try to access ONVIF device service
	onvifPath := fmt.Sprintf("http://%s/onvif/device_service", address)

	client := &http.Client{
		Timeout: c.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", onvifPath, nil)
	if err != nil {
		return false, latency, err
	}

	resp, err := client.Do(req)
	if err != nil {
		// Connection is open but ONVIF might not be available
		return false, latency, nil
	}
	defer resp.Body.Close()

	// Any response from ONVIF endpoint means it's available
	return resp.StatusCode != 404, latency, nil
}

// CheckSnapshot checks if snapshot URL is accessible
func (c *Client) CheckSnapshot(ctx context.Context, snapshotURL string) (bool, time.Duration, error) {
	if snapshotURL == "" {
		return false, 0, nil // Not an error, just not configured
	}

	start := time.Now()

	client := &http.Client{
		Timeout: c.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", snapshotURL, nil)
	if err != nil {
		return false, 0, fmt.Errorf("invalid snapshot URL: %w", err)
	}

	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return false, latency, fmt.Errorf("snapshot request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if response is an image
	contentType := resp.Header.Get("Content-Type")
	isImage := strings.HasPrefix(contentType, "image/")

	if resp.StatusCode == 200 && isImage {
		return true, latency, nil
	}

	// 401/403 means camera is there but needs auth
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return true, latency, nil
	}

	return false, latency, fmt.Errorf("snapshot returned status %d", resp.StatusCode)
}

// CheckAvailability performs a comprehensive camera check
func (c *Client) CheckAvailability(ctx context.Context, ipAddress string, rtspURL string, onvifPort int, snapshotURL string) (*CameraStatus, error) {
	status := &CameraStatus{}

	var latencies []time.Duration
	hasAnyConfig := rtspURL != "" || onvifPort > 0 || snapshotURL != ""

	// Check RTSP
	if rtspURL != "" {
		available, latency, err := c.CheckRTSP(ctx, rtspURL)
		status.RTSPAvailable = available
		if available {
			latencies = append(latencies, latency)
		}
		if err != nil && status.Error == "" {
			status.Error = err.Error()
		}
	}

	// Check ONVIF
	if onvifPort > 0 {
		available, latency, _ := c.CheckONVIF(ctx, ipAddress, onvifPort)
		status.ONVIFAvailable = available
		if available {
			latencies = append(latencies, latency)
		}
	}

	// Check Snapshot
	if snapshotURL != "" {
		available, latency, _ := c.CheckSnapshot(ctx, snapshotURL)
		status.SnapshotOK = available
		if available {
			latencies = append(latencies, latency)
		}
	}

	// If no specific camera config is set, fall back to simple TCP port check
	// This allows cameras without RTSP/ONVIF/Snapshot to still be monitored
	if !hasAnyConfig {
		start := time.Now()
		dialer := net.Dialer{Timeout: c.Timeout}

		// Try HTTP port first (most cameras have web interface)
		conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:80", ipAddress))
		if err == nil {
			conn.Close()
			status.Available = true
			status.Latency = time.Since(start)
			return status, nil
		}

		// Try RTSP port
		conn, err = dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:554", ipAddress))
		if err == nil {
			conn.Close()
			status.Available = true
			status.Latency = time.Since(start)
			return status, nil
		}

		// Camera not reachable
		status.Error = "Camera not reachable (ports 80/554 closed)"
		return status, nil
	}

	// Calculate average latency
	if len(latencies) > 0 {
		var total time.Duration
		for _, l := range latencies {
			total += l
		}
		status.Latency = total / time.Duration(len(latencies))
	}

	// Camera is available if any check passed
	status.Available = status.RTSPAvailable || status.ONVIFAvailable || status.SnapshotOK

	return status, nil
}
