package ping

import (
	"context"
	"fmt"
	"net"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// Pinger handles ICMP ping operations
type Pinger struct {
	Timeout time.Duration
	Count   int
}

// NewPinger creates a new pinger
func NewPinger(timeout time.Duration) *Pinger {
	return &Pinger{
		Timeout: timeout,
		Count:   3,
	}
}

// PingResult contains ping results
type PingResult struct {
	Success     bool
	PacketsSent int
	PacketsRecv int
	AvgLatency  time.Duration
	MinLatency  time.Duration
	MaxLatency  time.Duration
	PacketLoss  float64
}

// Ping performs an ICMP ping to the target
func (p *Pinger) Ping(ctx context.Context, target string) (*PingResult, error) {
	// Try ICMP ping first
	result, err := p.ICMPPing(ctx, target)
	if err == nil && result.Success {
		return result, nil
	}

	// Fall back to TCP ping on common ports if ICMP fails
	// This handles cases where ICMP is blocked but device is reachable
	tcpResult, tcpErr := p.TCPPingMulti(ctx, target, []int{80, 443, 22, 8080})
	if tcpErr == nil && tcpResult.Success {
		return tcpResult, nil
	}

	// Return the ICMP error if both failed
	if err != nil {
		return result, err
	}
	return result, fmt.Errorf("device unreachable")
}

// ICMPPing performs a real ICMP ping
func (p *Pinger) ICMPPing(ctx context.Context, target string) (*PingResult, error) {
	result := &PingResult{
		PacketsSent: p.Count,
	}

	pinger, err := probing.NewPinger(target)
	if err != nil {
		return result, fmt.Errorf("failed to create pinger: %w", err)
	}

	// Configure pinger
	pinger.Count = p.Count
	pinger.Timeout = p.Timeout
	pinger.SetPrivileged(true) // Use privileged mode on Windows

	// Run with context
	done := make(chan error, 1)
	go func() {
		done <- pinger.Run()
	}()

	select {
	case <-ctx.Done():
		pinger.Stop()
		return result, ctx.Err()
	case err := <-done:
		if err != nil {
			return result, err
		}
	}

	stats := pinger.Statistics()

	result.PacketsSent = stats.PacketsSent
	result.PacketsRecv = stats.PacketsRecv
	result.PacketLoss = stats.PacketLoss
	result.AvgLatency = stats.AvgRtt
	result.MinLatency = stats.MinRtt
	result.MaxLatency = stats.MaxRtt
	result.Success = stats.PacketsRecv > 0

	if !result.Success {
		return result, fmt.Errorf("all %d packets lost", p.Count)
	}

	return result, nil
}

// TCPPing performs a TCP connection test on a single port
func (p *Pinger) TCPPing(ctx context.Context, target string, port int) (*PingResult, error) {
	result := &PingResult{
		PacketsSent: p.Count,
	}

	var totalLatency time.Duration
	var minLatency, maxLatency time.Duration
	successCount := 0

	address := fmt.Sprintf("%s:%d", target, port)

	for i := 0; i < p.Count; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		start := time.Now()

		dialer := net.Dialer{
			Timeout: p.Timeout,
		}

		conn, err := dialer.DialContext(ctx, "tcp", address)
		latency := time.Since(start)

		if err == nil {
			conn.Close()
			successCount++
			totalLatency += latency

			if minLatency == 0 || latency < minLatency {
				minLatency = latency
			}
			if latency > maxLatency {
				maxLatency = latency
			}
		}

		// Small delay between attempts
		if i < p.Count-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	result.PacketsRecv = successCount
	result.Success = successCount > 0

	if successCount > 0 {
		result.AvgLatency = totalLatency / time.Duration(successCount)
		result.MinLatency = minLatency
		result.MaxLatency = maxLatency
	}

	result.PacketLoss = float64(p.Count-successCount) / float64(p.Count) * 100

	if !result.Success {
		return result, fmt.Errorf("all %d packets lost", p.Count)
	}

	return result, nil
}

// TCPPingMulti tries TCP ping on multiple ports, returns success if any port responds
func (p *Pinger) TCPPingMulti(ctx context.Context, target string, ports []int) (*PingResult, error) {
	for _, port := range ports {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := p.TCPPing(ctx, target, port)
		if err == nil && result.Success {
			return result, nil
		}
	}

	return &PingResult{
		PacketsSent: p.Count,
		Success:     false,
		PacketLoss:  100,
	}, fmt.Errorf("no TCP ports responded")
}

// TCPPortCheck checks if a specific TCP port is open
func (p *Pinger) TCPPortCheck(ctx context.Context, target string, port int) (bool, time.Duration, error) {
	address := fmt.Sprintf("%s:%d", target, port)

	start := time.Now()

	dialer := net.Dialer{
		Timeout: p.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	latency := time.Since(start)

	if err != nil {
		return false, latency, err
	}

	conn.Close()
	return true, latency, nil
}

// CheckMultiplePorts checks multiple TCP ports
func (p *Pinger) CheckMultiplePorts(ctx context.Context, target string, ports []int) map[int]bool {
	results := make(map[int]bool)

	for _, port := range ports {
		select {
		case <-ctx.Done():
			return results
		default:
		}

		open, _, _ := p.TCPPortCheck(ctx, target, port)
		results[port] = open
	}

	return results
}
