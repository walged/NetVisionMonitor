package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"netvisionmonitor/internal/models"
)

// EventRepository handles event database operations
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Create inserts a new event
func (r *EventRepository) Create(event *models.Event) error {
	result, err := r.db.Exec(`
		INSERT INTO events (device_id, type, level, message, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		event.DeviceID, event.Type, event.Level, event.Message, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	event.ID = id
	return nil
}

// GetByID retrieves an event by ID
func (r *EventRepository) GetByID(id int64) (*models.Event, error) {
	event := &models.Event{}
	err := r.db.QueryRow(`
		SELECT id, device_id, type, level, message, created_at
		FROM events WHERE id = ?`, id,
	).Scan(&event.ID, &event.DeviceID, &event.Type, &event.Level, &event.Message, &event.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return event, nil
}

// GetFiltered retrieves events with filters
func (r *EventRepository) GetFiltered(filter models.EventFilter) ([]models.Event, error) {
	query := "SELECT id, device_id, type, level, message, created_at FROM events WHERE 1=1"
	var args []interface{}

	if filter.DeviceID != nil {
		query += " AND device_id = ?"
		args = append(args, *filter.DeviceID)
	}
	if filter.Type != nil {
		query += " AND type = ?"
		args = append(args, *filter.Type)
	}
	if filter.Level != nil {
		query += " AND level = ?"
		args = append(args, *filter.Level)
	}
	if filter.StartTime != nil {
		query += " AND created_at >= ?"
		args = append(args, *filter.StartTime)
	}
	if filter.EndTime != nil {
		query += " AND created_at <= ?"
		args = append(args, *filter.EndTime)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		err := rows.Scan(&e.ID, &e.DeviceID, &e.Type, &e.Level, &e.Message, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

// GetRecent retrieves recent events
func (r *EventRepository) GetRecent(limit int) ([]models.Event, error) {
	return r.GetFiltered(models.EventFilter{Limit: limit})
}

// GetByDevice retrieves events for a specific device
func (r *EventRepository) GetByDevice(deviceID int64, limit int) ([]models.Event, error) {
	return r.GetFiltered(models.EventFilter{DeviceID: &deviceID, Limit: limit})
}

// Delete removes an event by ID
func (r *EventRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM events WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// DeleteOlderThan removes events older than specified time
func (r *EventRepository) DeleteOlderThan(before time.Time) (int64, error) {
	result, err := r.db.Exec("DELETE FROM events WHERE created_at < ?", before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old events: %w", err)
	}
	return result.RowsAffected()
}

// DeleteByDevice removes all events for a device
func (r *EventRepository) DeleteByDevice(deviceID int64) error {
	_, err := r.db.Exec("DELETE FROM events WHERE device_id = ?", deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete device events: %w", err)
	}
	return nil
}

// DeleteAll removes all events
func (r *EventRepository) DeleteAll() error {
	_, err := r.db.Exec("DELETE FROM events")
	if err != nil {
		return fmt.Errorf("failed to delete all events: %w", err)
	}
	return nil
}

// Count returns the total number of events
func (r *EventRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}
	return count, nil
}

// CountByLevel returns event counts by level
func (r *EventRepository) CountByLevel() (map[string]int, error) {
	rows, err := r.db.Query("SELECT level, COUNT(*) FROM events GROUP BY level")
	if err != nil {
		return nil, fmt.Errorf("failed to count events by level: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		counts[level] = count
	}
	return counts, nil
}

// EventListResult contains paginated events list result
type EventListResult struct {
	Events     []models.Event `json:"events"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// GetFilteredWithTotal retrieves events with filters and returns total count
func (r *EventRepository) GetFilteredWithTotal(filter models.EventFilter, page, pageSize int) (*EventListResult, error) {
	whereClause := "WHERE 1=1"
	var args []interface{}

	if filter.DeviceID != nil {
		whereClause += " AND device_id = ?"
		args = append(args, *filter.DeviceID)
	}
	if filter.Type != nil {
		whereClause += " AND type = ?"
		args = append(args, *filter.Type)
	}
	if filter.Level != nil {
		whereClause += " AND level = ?"
		args = append(args, *filter.Level)
	}
	if filter.StartTime != nil {
		whereClause += " AND created_at >= ?"
		args = append(args, *filter.StartTime)
	}
	if filter.EndTime != nil {
		whereClause += " AND created_at <= ?"
		args = append(args, *filter.EndTime)
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM events " + whereClause
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}

	// Calculate pagination
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	if page <= 0 {
		page = 1
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * pageSize

	// Query events
	query := "SELECT id, device_id, type, level, message, created_at FROM events " +
		whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		err := rows.Scan(&e.ID, &e.DeviceID, &e.Type, &e.Level, &e.Message, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}

	return &EventListResult{
		Events:     events,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Search searches events by message content
func (r *EventRepository) Search(query string, limit int) ([]models.Event, error) {
	rows, err := r.db.Query(`
		SELECT id, device_id, type, level, message, created_at
		FROM events WHERE message LIKE ?
		ORDER BY created_at DESC LIMIT ?`,
		"%"+strings.ReplaceAll(query, "%", "\\%")+"%", limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		err := rows.Scan(&e.ID, &e.DeviceID, &e.Type, &e.Level, &e.Message, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}
