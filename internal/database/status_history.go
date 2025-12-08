package database

import (
	"database/sql"
	"fmt"
	"time"

	"netvisionmonitor/internal/models"
)

type StatusHistoryRepository struct {
	db *sql.DB
}

func NewStatusHistoryRepository(db *sql.DB) *StatusHistoryRepository {
	return &StatusHistoryRepository{db: db}
}

// Record saves a new status check result
func (r *StatusHistoryRepository) Record(deviceID int64, status string, latency int64) error {
	_, err := r.db.Exec(`
		INSERT INTO status_history (device_id, status, latency, created_at)
		VALUES (?, ?, ?, ?)
	`, deviceID, status, latency, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record status history: %w", err)
	}
	return nil
}

// GetHistory returns status history for a device
func (r *StatusHistoryRepository) GetHistory(deviceID int64, limit int) ([]models.StatusHistory, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.Query(`
		SELECT id, device_id, status, latency, created_at
		FROM status_history
		WHERE device_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.StatusHistory
	for rows.Next() {
		var h models.StatusHistory
		if err := rows.Scan(&h.ID, &h.DeviceID, &h.Status, &h.Latency, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

// GetLatencyPoints returns latency data points for graphing
func (r *StatusHistoryRepository) GetLatencyPoints(deviceID int64, hours int) ([]models.LatencyPoint, error) {
	if hours <= 0 {
		hours = 24
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	rows, err := r.db.Query(`
		SELECT created_at, latency, status
		FROM status_history
		WHERE device_id = ? AND created_at >= ?
		ORDER BY created_at ASC
	`, deviceID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.LatencyPoint
	for rows.Next() {
		var p models.LatencyPoint
		if err := rows.Scan(&p.Timestamp, &p.Latency, &p.Status); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

// GetStats returns aggregated statistics for a device
func (r *StatusHistoryRepository) GetStats(deviceID int64) (*models.DeviceStats, error) {
	stats := &models.DeviceStats{DeviceID: deviceID}

	// Get totals
	err := r.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'online' THEN 1 ELSE 0 END) as online,
			SUM(CASE WHEN status = 'offline' THEN 1 ELSE 0 END) as offline,
			COALESCE(AVG(CASE WHEN status = 'online' THEN latency END), 0) as avg_latency,
			COALESCE(MIN(CASE WHEN status = 'online' THEN latency END), 0) as min_latency,
			COALESCE(MAX(CASE WHEN status = 'online' THEN latency END), 0) as max_latency
		FROM status_history
		WHERE device_id = ?
	`, deviceID).Scan(
		&stats.TotalChecks,
		&stats.OnlineCount,
		&stats.OfflineCount,
		&stats.AvgLatency,
		&stats.MinLatency,
		&stats.MaxLatency,
	)
	if err != nil {
		return nil, err
	}

	// Calculate uptime
	if stats.TotalChecks > 0 {
		stats.UptimePercent = float64(stats.OnlineCount) / float64(stats.TotalChecks) * 100
	}

	// Get last online time
	var lastOnline sql.NullTime
	err = r.db.QueryRow(`
		SELECT created_at FROM status_history
		WHERE device_id = ? AND status = 'online'
		ORDER BY created_at DESC LIMIT 1
	`, deviceID).Scan(&lastOnline)
	if err == nil && lastOnline.Valid {
		str := lastOnline.Time.Format("2006-01-02 15:04:05")
		stats.LastOnline = &str
	}

	// Get last offline time
	var lastOffline sql.NullTime
	err = r.db.QueryRow(`
		SELECT created_at FROM status_history
		WHERE device_id = ? AND status = 'offline'
		ORDER BY created_at DESC LIMIT 1
	`, deviceID).Scan(&lastOffline)
	if err == nil && lastOffline.Valid {
		str := lastOffline.Time.Format("2006-01-02 15:04:05")
		stats.LastOffline = &str
	}

	// Get current streak
	rows, err := r.db.Query(`
		SELECT status FROM status_history
		WHERE device_id = ?
		ORDER BY created_at DESC
		LIMIT 100
	`, deviceID)
	if err == nil {
		defer rows.Close()
		var streak int64 = 0
		var streakStatus string = ""
		first := true

		for rows.Next() {
			var status string
			if err := rows.Scan(&status); err != nil {
				continue // Skip corrupted rows
			}
			if first {
				streakStatus = status
				first = false
			}
			if status == streakStatus {
				streak++
			} else {
				break
			}
		}
		stats.CurrentStreak = streak
		stats.StreakStatus = streakStatus
	}

	return stats, nil
}

// GetUptimeByPeriod returns uptime grouped by time period
func (r *StatusHistoryRepository) GetUptimeByPeriod(deviceID int64, period string, count int) ([]models.UptimePoint, error) {
	var format string
	switch period {
	case "hour":
		format = "%Y-%m-%d %H:00"
	case "day":
		format = "%Y-%m-%d"
	case "week":
		format = "%Y-W%W"
	default:
		format = "%Y-%m-%d"
		period = "day"
	}

	if count <= 0 {
		count = 30
	}

	rows, err := r.db.Query(`
		SELECT
			strftime(?, created_at) as period_start,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'online' THEN 1 ELSE 0 END) as online
		FROM status_history
		WHERE device_id = ?
		GROUP BY strftime(?, created_at)
		ORDER BY period_start DESC
		LIMIT ?
	`, format, deviceID, format, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.UptimePoint
	for rows.Next() {
		var p models.UptimePoint
		var total, online int64
		if err := rows.Scan(&p.PeriodStart, &total, &online); err != nil {
			return nil, err
		}
		p.Period = period
		p.TotalChecks = total
		if total > 0 {
			p.UptimePercent = float64(online) / float64(total) * 100
		}
		points = append(points, p)
	}

	// Reverse to get chronological order
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}

// DeleteOlderThan removes old history records
func (r *StatusHistoryRepository) DeleteOlderThan(before time.Time) (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM status_history WHERE created_at < ?
	`, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetRecentChanges returns recent status changes (transitions)
func (r *StatusHistoryRepository) GetRecentChanges(deviceID int64, limit int) ([]models.StatusHistory, error) {
	if limit <= 0 {
		limit = 20
	}

	// Get status changes by comparing with previous record
	rows, err := r.db.Query(`
		WITH ranked AS (
			SELECT
				id, device_id, status, latency, created_at,
				LAG(status) OVER (ORDER BY created_at) as prev_status
			FROM status_history
			WHERE device_id = ?
		)
		SELECT id, device_id, status, latency, created_at
		FROM ranked
		WHERE status != prev_status OR prev_status IS NULL
		ORDER BY created_at DESC
		LIMIT ?
	`, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []models.StatusHistory
	for rows.Next() {
		var h models.StatusHistory
		if err := rows.Scan(&h.ID, &h.DeviceID, &h.Status, &h.Latency, &h.CreatedAt); err != nil {
			return nil, err
		}
		changes = append(changes, h)
	}
	return changes, nil
}
