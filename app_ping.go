package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"netvisionmonitor/internal/monitoring/ping"
)

// PingResult represents the result of a ping operation
type PingResult struct {
	Success     bool    `json:"success"`
	PacketsSent int     `json:"packets_sent"`
	PacketsRecv int     `json:"packets_recv"`
	PacketLoss  float64 `json:"packet_loss"`
	AvgLatency  float64 `json:"avg_latency_ms"`
	MinLatency  float64 `json:"min_latency_ms"`
	MaxLatency  float64 `json:"max_latency_ms"`
	Error       string  `json:"error,omitempty"`
}

// PingDevice performs an internal ping to the specified IP address
func (a *App) PingDevice(ipAddress string) (*PingResult, error) {
	if ipAddress == "" {
		return nil, fmt.Errorf("IP address is required")
	}

	pinger := ping.NewPinger(3 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := pinger.Ping(ctx, ipAddress)
	if err != nil {
		return &PingResult{
			Success:     false,
			PacketsSent: 3,
			PacketsRecv: 0,
			PacketLoss:  100,
			Error:       err.Error(),
		}, nil
	}

	return &PingResult{
		Success:     result.Success,
		PacketsSent: result.PacketsSent,
		PacketsRecv: result.PacketsRecv,
		PacketLoss:  result.PacketLoss,
		AvgLatency:  float64(result.AvgLatency.Milliseconds()),
		MinLatency:  float64(result.MinLatency.Milliseconds()),
		MaxLatency:  float64(result.MaxLatency.Milliseconds()),
	}, nil
}

// OpenPingCmd opens Windows command prompt with ping command
func (a *App) OpenPingCmd(ipAddress string) error {
	if ipAddress == "" {
		return fmt.Errorf("IP address is required")
	}

	// Open cmd.exe with ping -t command (continuous ping)
	cmd := exec.Command("cmd", "/c", "start", "cmd", "/k", "ping", "-t", ipAddress)
	return cmd.Start()
}
