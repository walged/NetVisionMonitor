package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/logger"
	"netvisionmonitor/internal/models"
	"netvisionmonitor/internal/monitoring/camera"
	"netvisionmonitor/internal/monitoring/ping"
	"netvisionmonitor/internal/monitoring/snmp"
)

// Monitor manages the monitoring cycle
type Monitor struct {
	db           *database.Database
	pool         *WorkerPool
	interval     time.Duration
	pingTimeout  time.Duration
	snmpTimeout  time.Duration

	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
	mu           sync.RWMutex

	// Callbacks
	onStatusChange func(deviceID int64, oldStatus, newStatus string)
	onEvent        func(event *models.Event)
}

// Config holds monitor configuration
type Config struct {
	Interval     time.Duration
	PingTimeout  time.Duration
	SNMPTimeout  time.Duration
	Workers      int
}

// DefaultConfig returns default monitor configuration
func DefaultConfig() Config {
	return Config{
		Interval:    30 * time.Second,
		PingTimeout: 3 * time.Second,
		SNMPTimeout: 5 * time.Second,
		Workers:     10,
	}
}

// NewMonitor creates a new monitor instance
func NewMonitor(db *database.Database, cfg Config) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Monitor{
		db:          db,
		interval:    cfg.Interval,
		pingTimeout: cfg.PingTimeout,
		snmpTimeout: cfg.SNMPTimeout,
		ctx:         ctx,
		cancel:      cancel,
	}

	m.pool = NewWorkerPool(cfg.Workers, m.handleResult)

	return m
}

// SetStatusChangeHandler sets callback for status changes
func (m *Monitor) SetStatusChangeHandler(handler func(deviceID int64, oldStatus, newStatus string)) {
	m.onStatusChange = handler
}

// SetEventHandler sets callback for events
func (m *Monitor) SetEventHandler(handler func(event *models.Event)) {
	m.onEvent = handler
}

// Start begins the monitoring cycle
func (m *Monitor) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	// Create new context for this monitoring session
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.mu.Unlock()

	m.pool.Start()

	go m.monitorLoop()

	logger.Info("Monitor started with %v interval", m.interval)
}

// Stop halts the monitoring cycle
func (m *Monitor) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	m.cancel()
	m.pool.Stop()

	logger.Info("Monitor stopped")
}

// IsRunning returns whether the monitor is running
func (m *Monitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// SetInterval updates the monitoring interval
func (m *Monitor) SetInterval(interval time.Duration) {
	m.mu.Lock()
	m.interval = interval
	m.mu.Unlock()
}

// RunOnce performs a single monitoring cycle
func (m *Monitor) RunOnce() {
	m.mu.RLock()
	wasRunning := m.running
	m.mu.RUnlock()

	// If monitor is not running, we need to temporarily start the pool
	if !wasRunning {
		// Create temporary context for this run
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		m.mu.Lock()
		m.ctx = ctx
		m.mu.Unlock()

		m.pool.Start()
		defer m.pool.Stop()
	}

	m.checkAllDevices()

	// If we started the pool temporarily, wait a bit for results to process
	if !wasRunning {
		time.Sleep(500 * time.Millisecond)
	}
}

// monitorLoop runs the main monitoring loop
func (m *Monitor) monitorLoop() {
	// Initial check
	m.checkAllDevices()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkAllDevices()
		case <-m.ctx.Done():
			return
		}
	}
}

// checkAllDevices submits monitoring tasks for all devices
func (m *Monitor) checkAllDevices() {
	deviceRepo := database.NewDeviceRepository(m.db.DB())

	devices, err := deviceRepo.GetAll()
	if err != nil {
		logger.Error("Error fetching devices: %v", err)
		return
	}

	logger.Debug("Checking %d devices", len(devices))

	for _, device := range devices {
		task := m.createTask(device)
		m.pool.Submit(task)
	}
}

// createTask creates a monitoring task for a device
func (m *Monitor) createTask(device models.Device) Task {
	return Task{
		DeviceID:   device.ID,
		DeviceType: string(device.Type),
		IPAddress:  device.IPAddress,
		Execute: func(ctx context.Context) error {
			return m.checkDevice(ctx, device)
		},
	}
}

// checkDevice performs monitoring check for a single device
func (m *Monitor) checkDevice(ctx context.Context, device models.Device) error {
	switch device.Type {
	case models.DeviceTypeSwitch:
		return m.checkSwitch(ctx, device)
	case models.DeviceTypeServer:
		return m.checkServer(ctx, device)
	case models.DeviceTypeCamera:
		return m.checkCamera(ctx, device)
	default:
		return m.checkPing(ctx, device)
	}
}

// checkSwitch checks a switch via SNMP, falls back to ping if SNMP fails
func (m *Monitor) checkSwitch(ctx context.Context, device models.Device) error {
	// Get switch details
	switchRepo := database.NewSwitchRepository(m.db.DB())
	sw, err := switchRepo.GetByDeviceID(device.ID)
	if err != nil || sw == nil {
		// Fallback to ping
		return m.checkPing(ctx, device)
	}

	var client *snmp.Client

	if sw.SNMPVersion == "v3" {
		// Use SNMPv3
		logger.Debug("checkSwitch: using SNMPv3 for %s - user=%s, security=%s, authProto=%s, privProto=%s",
			device.IPAddress, sw.SNMPv3User, sw.SNMPv3Security, sw.SNMPv3AuthProto, sw.SNMPv3PrivProto)
		client = snmp.NewClientV3(
			device.IPAddress,
			sw.SNMPv3User,
			sw.SNMPv3Security,
			sw.SNMPv3AuthProto,
			sw.SNMPv3AuthPass,
			sw.SNMPv3PrivProto,
			sw.SNMPv3PrivPass,
			m.snmpTimeout,
		)
	} else {
		// Use SNMPv1/v2c
		community := sw.SNMPCommunity
		if community == "" {
			community = "public"
		}
		client = snmp.NewClient(device.IPAddress, community, sw.SNMPVersion, m.snmpTimeout)
	}

	available, _, err := client.CheckAvailability(ctx)
	if err != nil || !available {
		// SNMP failed - fallback to ping
		logger.Debug("SNMP check failed for %s, falling back to ping: %v", device.IPAddress, err)
		return m.checkPing(ctx, device)
	}

	// Optionally update port statuses
	go m.updatePortStatuses(ctx, device.ID, client, sw.PortCount)

	return nil
}

// updatePortStatuses updates switch port statuses
func (m *Monitor) updatePortStatuses(ctx context.Context, deviceID int64, client *snmp.Client, portCount int) {
	interfaces, err := client.GetAllInterfaceStatuses(ctx, portCount)
	if err != nil {
		logger.Debug("Failed to get interface statuses for device %d: %v", deviceID, err)
		return
	}

	switchRepo := database.NewSwitchRepository(m.db.DB())
	ports, err := switchRepo.GetPorts(deviceID)
	if err != nil {
		return
	}

	// Create a map of port numbers to interface statuses
	statusMap := make(map[int]string)
	for _, iface := range interfaces {
		statusMap[iface.Index] = iface.Status
	}

	// Update port statuses
	for _, port := range ports {
		if status, ok := statusMap[port.PortNumber]; ok {
			if port.Status != status {
				// Status changed
				oldStatus := port.Status
				switchRepo.UpdatePortStatus(port.ID, status)

				// Emit event
				if m.onEvent != nil {
					eventType := models.EventTypePortUp
					level := models.EventLevelInfo
					if status == "down" {
						eventType = models.EventTypePortDown
						level = models.EventLevelWarn
					}

					m.onEvent(&models.Event{
						DeviceID: &deviceID,
						Type:     eventType,
						Level:    level,
						Message:  fmt.Sprintf("Port %d changed from %s to %s", port.PortNumber, oldStatus, status),
					})
				}
			}
		}
	}
}

// checkServer checks a server via ping and TCP
func (m *Monitor) checkServer(ctx context.Context, device models.Device) error {
	// Get server details
	serverRepo := database.NewServerRepository(m.db.DB())
	srv, err := serverRepo.GetByDeviceID(device.ID)
	if err != nil || srv == nil {
		// Fallback to ping
		return m.checkPing(ctx, device)
	}

	pinger := ping.NewPinger(m.pingTimeout)

	// First check basic connectivity
	result, err := pinger.Ping(ctx, device.IPAddress)
	if err != nil || !result.Success {
		return err
	}

	// Check TCP ports if configured
	if srv.TCPPorts != "" && srv.TCPPorts != "[]" {
		var ports []int
		if err := json.Unmarshal([]byte(srv.TCPPorts), &ports); err == nil && len(ports) > 0 {
			portStatus := pinger.CheckMultiplePorts(ctx, device.IPAddress, ports)

			// All ports should be open for full success
			for _, open := range portStatus {
				if !open {
					return nil // Still online, but some ports closed
				}
			}
		}
	}

	return nil
}

// checkCamera checks a camera via RTSP/ONVIF or falls back to ping
func (m *Monitor) checkCamera(ctx context.Context, device models.Device) error {
	// Get camera details
	cameraRepo := database.NewCameraRepository(m.db.DB())
	cam, err := cameraRepo.GetByDeviceID(device.ID)
	if err != nil || cam == nil {
		// No camera config - fallback to ping
		return m.checkPing(ctx, device)
	}

	// Check if any camera-specific config is set
	hasConfig := cam.RTSPURL != "" || cam.ONVIFPort > 0 || cam.SnapshotURL != ""
	if !hasConfig {
		// No specific camera config - use ping
		return m.checkPing(ctx, device)
	}

	client := camera.NewClient(m.pingTimeout)

	status, err := client.CheckAvailability(ctx, device.IPAddress, cam.RTSPURL, cam.ONVIFPort, cam.SnapshotURL)
	if err != nil {
		// Camera check failed, try ping as fallback
		pingErr := m.checkPing(ctx, device)
		if pingErr == nil {
			return nil // Device is reachable via ping
		}
		return err
	}

	if !status.Available {
		// Camera-specific checks failed, try ping as fallback
		pingErr := m.checkPing(ctx, device)
		if pingErr == nil {
			return nil // Device is reachable via ping
		}
		return &cameraError{message: status.Error}
	}

	return nil
}

// checkPing performs a simple ping check
func (m *Monitor) checkPing(ctx context.Context, device models.Device) error {
	pinger := ping.NewPinger(m.pingTimeout)

	result, err := pinger.Ping(ctx, device.IPAddress)
	if err != nil {
		return err
	}

	if !result.Success {
		return &pingError{packetLoss: result.PacketLoss}
	}

	return nil
}

// handleResult processes monitoring results
func (m *Monitor) handleResult(result Result) {
	deviceRepo := database.NewDeviceRepository(m.db.DB())
	historyRepo := database.NewStatusHistoryRepository(m.db.DB())

	// Get current device status
	device, err := deviceRepo.GetByID(result.DeviceID)
	if err != nil || device == nil {
		return
	}

	oldStatus := string(device.Status)
	newStatus := result.Status

	// Update device status in database
	deviceRepo.UpdateStatus(result.DeviceID, models.DeviceStatus(newStatus))

	// Record status in history (convert latency to milliseconds)
	latencyMs := result.Latency.Milliseconds()
	historyRepo.Record(result.DeviceID, newStatus, latencyMs)

	// Check for status change
	if oldStatus != newStatus && oldStatus != "unknown" {
		if m.onStatusChange != nil {
			m.onStatusChange(result.DeviceID, oldStatus, newStatus)
		}

		// Create event
		if m.onEvent != nil {
			eventType := models.EventTypeDeviceOnline
			level := models.EventLevelInfo
			message := device.Name + " is now online"

			if newStatus == "offline" {
				eventType = models.EventTypeDeviceOffline
				level = models.EventLevelError
				message = device.Name + " is now offline"
				if result.Error != nil {
					message += ": " + result.Error.Error()
				}
			}

			m.onEvent(&models.Event{
				DeviceID: &result.DeviceID,
				Type:     eventType,
				Level:    level,
				Message:  message,
			})
		}
	}
}

// Custom error types
type pingError struct {
	packetLoss float64
}

func (e *pingError) Error() string {
	return fmt.Sprintf("ping failed with %.0f%% packet loss", e.packetLoss)
}

type cameraError struct {
	message string
}

func (e *cameraError) Error() string {
	if e.message != "" {
		return e.message
	}
	return "camera unavailable"
}
