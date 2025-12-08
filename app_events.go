package main

import (
	"fmt"
	"time"

	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/models"
)

// EventFilterInput is used for filtering events from frontend
type EventFilterInput struct {
	DeviceID  *int64  `json:"device_id,omitempty"`
	Type      *string `json:"type,omitempty"`
	Level     *string `json:"level,omitempty"`
	StartTime *string `json:"start_time,omitempty"` // ISO 8601 format
	EndTime   *string `json:"end_time,omitempty"`   // ISO 8601 format
	Page      int     `json:"page"`
	PageSize  int     `json:"page_size"`
	Limit     int     `json:"limit"`
	Offset    int     `json:"offset"`
}

// GetEvents returns events with simple limit/offset
func (a *App) GetEvents(limit, offset int) ([]models.Event, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 100
	}

	repo := database.NewEventRepository(a.db.DB())
	return repo.GetFiltered(models.EventFilter{
		Limit:  limit,
		Offset: offset,
	})
}

// GetEventsPaginated returns paginated events with filters
func (a *App) GetEventsPaginated(filter EventFilterInput) (*database.EventListResult, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	f := models.EventFilter{}

	if filter.DeviceID != nil {
		f.DeviceID = filter.DeviceID
	}

	if filter.Type != nil {
		t := models.EventType(*filter.Type)
		f.Type = &t
	}

	if filter.Level != nil {
		l := models.EventLevel(*filter.Level)
		f.Level = &l
	}

	if filter.StartTime != nil {
		t, err := time.Parse(time.RFC3339, *filter.StartTime)
		if err == nil {
			f.StartTime = &t
		}
	}

	if filter.EndTime != nil {
		t, err := time.Parse(time.RFC3339, *filter.EndTime)
		if err == nil {
			f.EndTime = &t
		}
	}

	page := filter.Page
	pageSize := filter.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	repo := database.NewEventRepository(a.db.DB())
	return repo.GetFilteredWithTotal(f, page, pageSize)
}

// GetEventsFiltered returns filtered events
func (a *App) GetEventsFiltered(filter EventFilterInput) ([]models.Event, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	f := models.EventFilter{
		DeviceID: filter.DeviceID,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	}

	if filter.Type != nil {
		t := models.EventType(*filter.Type)
		f.Type = &t
	}

	if filter.Level != nil {
		l := models.EventLevel(*filter.Level)
		f.Level = &l
	}

	if filter.StartTime != nil {
		t, err := time.Parse(time.RFC3339, *filter.StartTime)
		if err == nil {
			f.StartTime = &t
		}
	}

	if filter.EndTime != nil {
		t, err := time.Parse(time.RFC3339, *filter.EndTime)
		if err == nil {
			f.EndTime = &t
		}
	}

	repo := database.NewEventRepository(a.db.DB())
	return repo.GetFiltered(f)
}

// GetRecentEvents returns recent events
func (a *App) GetRecentEvents(limit int) ([]models.Event, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 100
	}

	repo := database.NewEventRepository(a.db.DB())
	return repo.GetRecent(limit)
}

// GetDeviceEvents returns events for a specific device
func (a *App) GetDeviceEvents(deviceID int64, limit int) ([]models.Event, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 50
	}

	repo := database.NewEventRepository(a.db.DB())
	return repo.GetByDevice(deviceID, limit)
}

// SearchEvents searches events by message
func (a *App) SearchEvents(query string, limit int) ([]models.Event, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 100
	}

	repo := database.NewEventRepository(a.db.DB())
	return repo.Search(query, limit)
}

// DeleteEvent deletes an event by ID
func (a *App) DeleteEvent(id int64) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	repo := database.NewEventRepository(a.db.DB())
	return repo.Delete(id)
}

// ClearOldEvents removes events older than specified days
func (a *App) ClearOldEvents(days int) (int64, error) {
	if a.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	if days <= 0 {
		days = 30
	}

	before := time.Now().AddDate(0, 0, -days)
	repo := database.NewEventRepository(a.db.DB())
	return repo.DeleteOlderThan(before)
}

// ClearEvents removes all events
func (a *App) ClearEvents() error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	repo := database.NewEventRepository(a.db.DB())
	return repo.DeleteAll()
}

// GetEventStats returns event statistics
func (a *App) GetEventStats() (map[string]int, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	repo := database.NewEventRepository(a.db.DB())

	total, err := repo.Count()
	if err != nil {
		return nil, err
	}

	byLevel, err := repo.CountByLevel()
	if err != nil {
		return nil, err
	}

	byLevel["total"] = total
	return byLevel, nil
}

// CreateEvent creates a new event (internal use)
func (a *App) createEvent(deviceID *int64, eventType models.EventType, level models.EventLevel, message string) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	event := &models.Event{
		DeviceID: deviceID,
		Type:     eventType,
		Level:    level,
		Message:  message,
	}

	repo := database.NewEventRepository(a.db.DB())
	return repo.Create(event)
}
