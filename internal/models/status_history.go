package models

import "time"

// StatusHistory stores device status changes over time
type StatusHistory struct {
	ID        int64     `json:"id" db:"id"`
	DeviceID  int64     `json:"device_id" db:"device_id"`
	Status    string    `json:"status" db:"status"`
	Latency   int64     `json:"latency" db:"latency"` // Response time in ms
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// DeviceStats contains aggregated statistics for a device
type DeviceStats struct {
	DeviceID       int64   `json:"device_id"`
	TotalChecks    int64   `json:"total_checks"`
	OnlineCount    int64   `json:"online_count"`
	OfflineCount   int64   `json:"offline_count"`
	UptimePercent  float64 `json:"uptime_percent"`
	AvgLatency     float64 `json:"avg_latency"`
	MinLatency     int64   `json:"min_latency"`
	MaxLatency     int64   `json:"max_latency"`
	LastOnline     *string `json:"last_online,omitempty"`
	LastOffline    *string `json:"last_offline,omitempty"`
	CurrentStreak  int64   `json:"current_streak"`  // Current online/offline streak in checks
	StreakStatus   string  `json:"streak_status"`   // "online" or "offline"
}

// LatencyPoint represents a single point in latency graph
type LatencyPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Latency   int64     `json:"latency"`
	Status    string    `json:"status"`
}

// UptimePoint represents uptime percentage over time period
type UptimePoint struct {
	Period        string  `json:"period"` // "hour", "day", etc.
	PeriodStart   string  `json:"period_start"`
	UptimePercent float64 `json:"uptime_percent"`
	TotalChecks   int64   `json:"total_checks"`
}
