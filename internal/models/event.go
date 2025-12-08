package models

import "time"

type EventLevel string

const (
	EventLevelInfo  EventLevel = "info"
	EventLevelWarn  EventLevel = "warn"
	EventLevelError EventLevel = "error"
)

type EventType string

const (
	EventTypeDeviceOnline      EventType = "device_online"
	EventTypeDeviceOffline     EventType = "device_offline"
	EventTypePortUp            EventType = "port_up"
	EventTypePortDown          EventType = "port_down"
	EventTypeCameraNoStream    EventType = "camera_no_stream"
	EventTypeAuthError         EventType = "auth_error"
	EventTypeHighLatency       EventType = "high_latency"
	EventTypeMonitoringError   EventType = "monitoring_error"
	EventTypeSystemStart       EventType = "system_start"
	EventTypeSystemStop        EventType = "system_stop"
)

type Event struct {
	ID        int64      `json:"id"`
	DeviceID  *int64     `json:"device_id,omitempty"`
	Type      EventType  `json:"type"`
	Level     EventLevel `json:"level"`
	Message   string     `json:"message"`
	CreatedAt time.Time  `json:"created_at"`
}

type EventFilter struct {
	DeviceID  *int64      `json:"device_id,omitempty"`
	Type      *EventType  `json:"type,omitempty"`
	Level     *EventLevel `json:"level,omitempty"`
	StartTime *time.Time  `json:"start_time,omitempty"`
	EndTime   *time.Time  `json:"end_time,omitempty"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
}
